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

type Log struct {
	Address types.Address
	Topics  []types.Hash
	Data    []byte
	Height  uint64
}

type VM struct {
	Transaction
	VMConfig
	StateDb Database

	abort          int32
	intPool        *intPool
	instructionSet [256]operation
	quotaLeft      uint64
	quotaReturn    uint64
	logs           []*Log
	txs            []*Transaction
	returnData     []byte
}

func NewVM(tx Transaction) *VM {
	vm := &VM{Transaction: tx, instructionSet: simpleInstructionSet, logs: make([]*Log, 0), txs: make([]*Transaction, 0)}
	return vm
}

func (vm *VM) Cancel() {
	atomic.StoreInt32(&vm.abort, 1)
}

var (
	viteTokenTypeId = types.TokenTypeId{}
)

func calcQuota() uint64 {
	// TODO calculate quota, use 1000000 for test
	return 1000000
}

func calcCreateContractFee() *big.Int {
	// TODO calculate service fee for create contract, use 0 for test
	return big0
}

func canTransfer(db Database, addr types.Address, tokenTypeId types.TokenTypeId, tokenAmount *big.Int, feeAmount *big.Int) bool {
	return tokenAmount.Cmp(db.GetBalance(addr, tokenTypeId)) <= 0 && feeAmount.Cmp(db.GetBalance(addr, viteTokenTypeId)) <= 0
}

func (vm *VM) Create() (contractAddr types.Address, quota uint64, logs []*Log, txs []*Transaction, err error) {
	// check can make transaction
	quotaInit := calcQuota()
	vm.quotaLeft = quotaInit
	cost, err := intrinsicGasCost(vm.Data, true)
	if err != nil {
		return types.Address{}, 0, vm.logs, vm.txs, err
	}
	err = vm.useQuota(cost)
	if err != nil {
		return types.Address{}, 0, vm.logs, vm.txs, err
	}

	if vm.TxType == 1 {
		// send
		createFee := calcCreateContractFee()
		if !canTransfer(vm.StateDb, vm.From, vm.TokenTypeId, vm.Amount, createFee) {
			return types.Address{}, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, ErrInsufficientBalance
		}
		// sub balance and service fee
		vm.StateDb.SubBalance(vm.From, vm.TokenTypeId, vm.Amount)
		vm.StateDb.SubBalance(vm.From, viteTokenTypeId, createFee)
		return types.Address{}, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, nil
	} else {
		// receive
		// check depth, do nothing but refund if reach the max depth
		if vm.Depth > callCreateDepth {
			// TODO refund and delete account, solve at next version
			return types.Address{}, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, ErrDepth
		}
		// create a random address
		contractAddr, _, err := types.CreateAddress()
		if err != nil || vm.StateDb.IsExistAddress(contractAddr) {
			return types.Address{}, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, ErrContractAddressCreationFail
		}

		errorRevertId := vm.StateDb.Snapshot()

		// create contract account
		vm.StateDb.CreateAccount(contractAddr)
		vm.StateDb.AddBalance(contractAddr, vm.TokenTypeId, vm.Amount)

		// init contract state and set contract code
		contract := newContract(vm.From, contractAddr, vm.TokenTypeId, vm.Amount, nil)
		contract.setCallCode(contractAddr, types.DataHash(vm.Data), vm.Data)
		code, err := run(vm, contract)
		if err == nil {
			codeCost := uint64(len(code)) * contractCodeGas
			err = vm.useQuota(codeCost)
			if err == nil {
				vm.StateDb.SetContractCode(contractAddr, code)
				return contractAddr, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, nil
			}
		}

		// revert if out of quota, refund and delete account otherwise
		if err == ErrOutOfQuota {
			vm.StateDb.RevertToSnapShot(errorRevertId)
			return types.Address{}, quotaInit, vm.logs, vm.txs, err
		} else {
			if vm.Amount.Cmp(big0) > 0 {
				vm.txs = append(vm.txs, &Transaction{
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
			vm.StateDb.DeleteAccount(contractAddr)
		}
		return types.Address{}, quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, err
	}
}

func (vm *VM) Call() (quota uint64, logs []*Log, txs []*Transaction, err error) {
	quotaInit := calcQuota()
	vm.quotaLeft = quotaInit
	cost, err := intrinsicGasCost(vm.Data, false)
	if err != nil {
		return 0, vm.logs, vm.txs, err
	}
	err = vm.useQuota(cost)
	if err != nil {
		return 0, vm.logs, vm.txs, err
	}

	if vm.TxType == 1 {
		// send
		if !canTransfer(vm.StateDb, vm.From, vm.TokenTypeId, vm.Amount, big0) {
			return quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, ErrInsufficientBalance
		}
		vm.StateDb.SubBalance(vm.From, vm.TokenTypeId, vm.Amount)
		return quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, nil
	} else {
		// receive
		if !vm.StateDb.IsExistAddress(vm.To) {
			vm.StateDb.CreateAccount(vm.To)
		}
		revertId := vm.StateDb.Snapshot()
		vm.StateDb.AddBalance(vm.To, vm.TokenTypeId, vm.Amount)
		if vm.Depth > callCreateDepth {
			return quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, ErrDepth
		}
		contract := newContract(vm.From, vm.To, vm.TokenTypeId, vm.Amount, vm.Data)
		contract.setCallCode(vm.To, vm.StateDb.GetContractCodeHash(vm.To), vm.StateDb.GetContractCode(vm.To))
		_, err := run(vm, contract)
		if err == nil {
			return quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, nil
		} else {
			vm.StateDb.RevertToSnapShot(revertId)
			if err != ErrOutOfQuota && vm.Amount.Cmp(big0) > 0 {
				vm.StateDb.AddBalance(vm.To, vm.TokenTypeId, vm.Amount)
				vm.txs = append(vm.txs, &Transaction{
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
			if err == ErrOutOfQuota {
				return quotaInit, vm.logs, vm.txs, err
			} else {
				return quotaUsed(quotaInit, vm.quotaLeft, vm.quotaReturn), vm.logs, vm.txs, err
			}
		}
	}
}

func (vm *VM) delegateCall(contractAddr types.Address, data []byte) (ret []byte, err error) {
	revertId := vm.StateDb.Snapshot()
	contract := newContract(vm.From, vm.To, vm.TokenTypeId, vm.Amount, data)
	contract.setCallCode(contractAddr, vm.StateDb.GetContractCodeHash(contractAddr), vm.StateDb.GetContractCode(contractAddr))
	ret, err = run(vm, contract)
	if err != nil {
		vm.StateDb.RevertToSnapShot(revertId)
	}
	return ret, err
}

func run(vm *VM, c *contract) (ret []byte, err error) {
	if len(c.code) == 0 {
		return nil, nil
	}

	vm.intPool = poolOfIntPools.get()
	defer func() {
		poolOfIntPools.put(vm.intPool)
		vm.intPool = nil
	}()

	vm.returnData = nil

	var (
		op   opCode
		mem  = newMemory()
		st   = newStack()
		pc   = uint64(0)
		cost uint64
	)

	for atomic.LoadInt32(&vm.abort) == 0 {
		currentPc := pc
		op = c.getOp(pc)
		operation := vm.instructionSet[op]

		if !operation.valid {
			return nil, fmt.Errorf("invalid opcode 0x%x", int(op))
		}

		if err := operation.validateStack(st); err != nil {
			return nil, err
		}

		var memorySize uint64
		if operation.memorySize != nil {
			memSize, overflow := bigUint64(operation.memorySize(st))
			if overflow {
				return nil, errGasUintOverflow
			}
			if memorySize, overflow = SafeMul(toWordSize(memSize), 32); overflow {
				return nil, errGasUintOverflow
			}
		}

		cost, err = operation.gasCost(vm, c, st, mem, memorySize)
		if err != nil {
			return nil, err
		}
		err = vm.useQuota(cost)
		if err != nil {
			return nil, err
		}

		if memorySize > 0 {
			mem.resize(memorySize)
		}

		res, err := operation.execute(&pc, vm, c, mem, st)

		if vm.Debug {
			fmt.Println("--------------------")
			fmt.Printf("op: %v, pc: %v\nstack: [%v]\nmemory: [%v]\nstorage: [%v]\n", opCodeToString[op], currentPc, st.string(), mem.string(), vm.StateDb.GetStatesString(c.address))
			fmt.Println("--------------------")
		}

		if operation.returns {
			vm.returnData = res
		}

		switch {
		case err != nil:
			vm.quotaReturn = 0
			vm.logs = vm.logs[:0]
			vm.txs = vm.txs[:0]
			return nil, err
		case operation.halts:
			return res, nil
		case operation.reverts:
			vm.quotaReturn = 0
			vm.logs = vm.logs[:0]
			vm.txs = vm.txs[:0]
			return res, ErrExecutionReverted
		case !operation.jumps:
			pc++
		}
	}
	return nil, nil
}

func quotaUsed(quotaInit, quotaLeft, quotaReturn uint64) uint64 {
	return quotaInit - quotaLeft + min(quotaReturn, (quotaInit-quotaLeft)/2)
}

func (vm *VM) useQuota(cost uint64) error {
	if vm.quotaLeft < cost {
		return ErrOutOfQuota
	}
	vm.quotaLeft = vm.quotaLeft - cost
	return nil
}
