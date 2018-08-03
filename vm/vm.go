/**
Package vm implements the vite virtual machine
*/
package vm

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
	"sync/atomic"
)

type VMConfig struct {
	Debug bool
}

type Transaction struct {
	From        types.Address
	To          types.Address
	TxType      int
	TokenTypeId types.TokenTypeId
	Amount      *big.Int
	Data        []byte
	Depth       uint64

	SnapshotTimestamp *big.Int
	AccountHeight     *big.Int
	SnapshotHeight    *big.Int
}

type VM struct {
	Transaction
	VMConfig
	StateDb Database

	abort          int32
	intPool        *intPool
	instructionSet [256]operation
}

func NewVM(tx Transaction) *VM {
	vm := &VM{Transaction: tx, instructionSet: simpleInstructionSet}
	return vm
}

func (vm *VM) Cancel() {
	atomic.StoreInt32(&vm.abort, 1)
}

var (
	defaultQuota    = uint64(1000000)
	viteTokenTypeId = types.TokenTypeId{}
)

func canTransfer(db Database, addr types.Address, tokenTypeId types.TokenTypeId, tokenAmount *big.Int, feeAmount *big.Int) bool {
	return tokenAmount.Cmp(db.GetBalance(addr, tokenTypeId)) <= 0 && feeAmount.Cmp(db.GetBalance(addr, viteTokenTypeId)) <= 0
}

func (vm *VM) Create() (contractAddr types.Address, quotaUsed uint64, err error) {
	// check can make transaction
	// TODO calculate quota
	quotaInit := defaultQuota
	quotaLeft := quotaInit
	cost, err := intrinsicGasCost(vm.Data, true)
	if err != nil {
		return types.Address{}, 0, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return types.Address{}, 0, err
	}

	if vm.TxType == 1 {
		// send
		// TODO calculate service fee for create contract, use 0 for test
		createFee := big0
		if !canTransfer(vm.StateDb, vm.From, vm.TokenTypeId, vm.Amount, createFee) {
			return types.Address{}, quotaInit - quotaLeft, ErrInsufficientBalance
		}
		// sub balance and service fee
		vm.StateDb.SubBalance(vm.From, vm.TokenTypeId, vm.Amount)
		vm.StateDb.SubBalance(vm.From, viteTokenTypeId, createFee)
		return types.Address{}, quotaInit - quotaLeft, nil
	} else {
		// receive
		// check depth, do nothing if reach the max depth
		if vm.Depth > callCreateDepth {
			// TODO refund?
			return types.Address{}, quotaInit - quotaLeft, ErrDepth
		}
		// create a random address
		contractAddr, _, err := types.CreateAddress()
		if err != nil || vm.StateDb.IsExistAddress(contractAddr) {
			return types.Address{}, quotaInit - quotaLeft, ErrContractAddressCreationFail
		}

		errorRevertId := vm.StateDb.Snapshot()

		// create contract account
		vm.StateDb.CreateAccount(contractAddr)
		vm.StateDb.AddBalance(contractAddr, vm.TokenTypeId, vm.Amount)

		// init contract state and set contract code
		contract := NewContract(vm.From, contractAddr, vm.TokenTypeId, vm.Amount, nil)
		contract.SetCallCode(contractAddr, types.DataHash(vm.Data), vm.Data)
		code, quotaLeft, err := run(vm, contract, quotaLeft)
		if err == nil {
			codeCost := uint64(len(code)) * contractCodeGas
			quotaLeft, err = useQuota(quotaLeft, codeCost)
			if err == nil {
				vm.StateDb.SetContractCode(contractAddr, code)
				return contractAddr, quotaInit - quotaLeft, nil
			}
		}

		// revert if error
		vm.StateDb.RevertToSnapShot(errorRevertId)
		if err != ErrOutOfQuota && vm.Amount.Cmp(big0) > 0 {
			// TODO how to refund
		}
		return types.Address{}, quotaInit - quotaLeft, err
	}
}

func (vm *VM) Call() (quotaUsed uint64, err error) {
	// TODO calculate quota
	quotaInit := defaultQuota
	quotaLeft := quotaInit
	cost, err := intrinsicGasCost(vm.Data, false)
	if err != nil {
		return 0, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return 0, err
	}

	if vm.TxType == 1 {
		// send
		if !canTransfer(vm.StateDb, vm.From, vm.TokenTypeId, vm.Amount, big0) {
			return quotaInit - quotaLeft, ErrInsufficientBalance
		}
		vm.StateDb.SubBalance(vm.From, vm.TokenTypeId, vm.Amount)
		return quotaInit - quotaLeft, nil
	} else {
		// receive
		if !vm.StateDb.IsExistAddress(vm.To) {
			vm.StateDb.CreateAccount(vm.To)
		}
		revertId := vm.StateDb.Snapshot()
		vm.StateDb.AddBalance(vm.To, vm.TokenTypeId, vm.Amount)
		if vm.Depth > callCreateDepth {
			return quotaInit - quotaLeft, ErrDepth
		}
		contract := NewContract(vm.From, vm.To, vm.TokenTypeId, vm.Amount, vm.Data)
		contract.SetCallCode(vm.To, vm.StateDb.GetContractCodeHash(vm.To), vm.StateDb.GetContractCode(vm.To))
		_, quotaLeft, err := run(vm, contract, quotaLeft)
		if err == nil {
			return quotaInit - quotaLeft, nil
		} else {
			vm.StateDb.RevertToSnapShot(revertId)
			if err != ErrOutOfQuota && vm.Amount.Cmp(big0) > 0 {
				vm.StateDb.AddBalance(vm.To, vm.TokenTypeId, vm.Amount)
				vm.StateDb.AddTransaction(&Transaction{
					From:              vm.To,
					To:                vm.From,
					TxType:            1,
					TokenTypeId:       vm.TokenTypeId,
					Amount:            vm.Amount,
					Depth:             vm.Depth + 1,
					SnapshotTimestamp: vm.SnapshotTimestamp,
					SnapshotHeight:    vm.SnapshotHeight,
				})
			}
			return quotaInit - quotaLeft, err
		}
	}
}

func run(vm *VM, c *Contract, quota uint64) (ret []byte, quotaLeft uint64, err error) {
	if len(c.code) == 0 {
		return nil, quota, nil
	}

	vm.intPool = poolOfIntPools.get()
	defer func() {
		poolOfIntPools.put(vm.intPool)
		vm.intPool = nil
	}()

	var (
		op   opCode
		mem  = newMemory()
		st   = newStack()
		pc   = uint64(0)
		cost uint64
	)

	for atomic.LoadInt32(&vm.abort) == 0 {
		currentPc := pc
		op = c.GetOp(pc)
		operation := vm.instructionSet[op]

		if !operation.valid {
			return nil, quota, fmt.Errorf("invalid opcode 0x%x", int(op))
		}

		if err := operation.validateStack(st); err != nil {
			return nil, quota, err
		}

		var memorySize uint64
		if operation.memorySize != nil {
			memSize, overflow := bigUint64(operation.memorySize(st))
			if overflow {
				return nil, quota, errGasUintOverflow
			}
			if memorySize, overflow = SafeMul(toWordSize(memSize), 32); overflow {
				return nil, quota, errGasUintOverflow
			}
		}

		cost, err = operation.gasCost(vm, c, st, mem, memorySize)
		if err != nil {
			return nil, quota, err
		}
		quota, err = useQuota(quota, cost)
		if err != nil {
			return nil, quota, err
		}

		if memorySize > 0 {
			mem.resize(memorySize)
		}

		res, err := operation.execute(&pc, vm, c, mem, st)

		if vm.Debug {
			fmt.Println("--------------------")
			fmt.Printf("op: %v, pc: %v\nstack: [%v]\nmemory: [%v]\nstorage: [%v]\n", opCodeToString[op], currentPc, st.toString(), mem.toString(), vm.StateDb.GetStatesString(c.self))
			fmt.Println("--------------------")
		}

		switch {
		case err != nil:
			return nil, quota, err
		case operation.halts:
			return res, quota, nil
		case operation.reverts:
			return res, quota, ErrExecutionReverted
		case !operation.jumps:
			pc++
		}
	}
	return nil, quota, nil
}
