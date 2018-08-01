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

type Context struct {
	Depth    int
	TxType   int
	ReadOnly bool

	SnapshotTimestamp *big.Int
	AccountHeight     *big.Int
	SnapshotHeight    *big.Int
}

type VM struct {
	Context
	VMConfig
	StateDb database

	abort          int32
	returnData     []byte
	intPool        *intPool
	instructionSet [256]operation
}

func NewVM(ctx Context) *VM {
	vm := &VM{Context: ctx, instructionSet: simpleInstructionSet}
	return vm
}

type SendTransaction struct {
	To          types.Address
	TokenTypeId types.TokenTypeId
	Amount      *big.Int
	Data        []byte
}

var (
	defaultQuota    = uint64(1000000)
	viteTokenTypeId = types.TokenTypeId{}
)

func canTransfer(db database, addr types.Address, tokenTypeId types.TokenTypeId, tokenAmount *big.Int, feeAmount *big.Int) bool {
	return tokenAmount.Cmp(db.GetBalance(addr, tokenTypeId)) <= 0 && feeAmount.Cmp(db.GetBalance(addr, viteTokenTypeId)) <= 0
}

func (vm *VM) Create(from types.Address, data []byte, tokenTypeId types.TokenTypeId, amount *big.Int) (contractAddr types.Address, quotaUsed uint64, err error) {
	// TODO calculate quota
	quotaInit := defaultQuota
	quotaLeft := quotaInit
	cost, err := intrinsicGasCost(data, true)
	if err != nil {
		return types.Address{}, 0, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return types.Address{}, 0, err
	}

	if vm.TxType == 1 {
		// send
		// TODO calculate create fee
		createFee := big0
		if !canTransfer(vm.StateDb, from, tokenTypeId, amount, createFee) {
			return types.Address{}, quotaInit - quotaLeft, ErrInsufficientBalance
		}
		// sub balance and create cost
		vm.StateDb.SubBalance(from, tokenTypeId, amount)
		vm.StateDb.SubBalance(from, viteTokenTypeId, createFee)
		return types.Address{}, quotaInit - quotaLeft, nil
	} else {
		// receive
		contractAddr, _, err := types.CreateAddress()
		if err != nil || vm.StateDb.IsExistAddress(contractAddr) {
			return types.Address{}, quotaInit - quotaLeft, ErrContractAddressCreationFail
		}

		errorRevertId := vm.StateDb.Snapshot()
		vm.StateDb.CreateAccount(contractAddr)
		vm.StateDb.AddBalance(contractAddr, tokenTypeId, amount)
		revertId := vm.StateDb.Snapshot()
		contract := NewContract(from, contractAddr, tokenTypeId, amount, quotaLeft, nil)
		contract.SetCallCode(contractAddr, types.DataHash(data), data)
		code, err := run(vm, contract, nil)
		quotaLeft = contract.quotaLeft
		if err == nil {
			codeCost := uint64(len(code)) * ContractCodeGas
			quotaLeft, err = useQuota(quotaLeft, codeCost)
			if err == nil {
				vm.StateDb.SetContractCode(contractAddr, code)
				return contractAddr, quotaInit - quotaLeft, nil
			}
		}
		if err == errExecutionReverted {
			vm.StateDb.RevertToSnapShot(revertId)
			return types.Address{}, quotaInit - quotaLeft, err
		} else {
			vm.StateDb.RevertToSnapShot(errorRevertId)
			return types.Address{}, quotaInit - quotaLeft, err
		}
	}
}

func (vm *VM) Call(from types.Address, to types.Address, data []byte, tokenTypeId types.TokenTypeId, amount *big.Int) (sendList []SendTransaction, quotaUsed uint64, err error) {
	// TODO calculate quota
	quotaInit := defaultQuota
	quotaLeft := quotaInit
	cost, err := intrinsicGasCost(data, false)
	if err != nil {
		return nil, 0, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return nil, 0, err
	}

	if vm.TxType == 1 {
		// send
		if !canTransfer(vm.StateDb, from, tokenTypeId, amount, big0) {
			return nil, quotaInit - quotaLeft, ErrInsufficientBalance
		}
		// sub balance
		vm.StateDb.SubBalance(from, tokenTypeId, amount)
		return nil, quotaInit - quotaLeft, nil
	} else {
		// receive
		errorRevertId := vm.StateDb.Snapshot()
		if !vm.StateDb.IsExistAddress(to) {
			vm.StateDb.CreateAccount(to)
		}
		vm.StateDb.AddBalance(to, tokenTypeId, amount)
		revertId := vm.StateDb.Snapshot()
		contract := NewContract(from, to, tokenTypeId, amount, quotaLeft, nil)
		contract.SetCallCode(to, vm.StateDb.GetContractCodeHash(to), vm.StateDb.GetContractCode(to))
		_, err := run(vm, contract, data)
		quotaLeft = contract.quotaLeft
		if err == nil {
			// TODO return sendList
			return nil, quotaInit - quotaLeft, nil
		} else if err == errExecutionReverted {
			vm.StateDb.RevertToSnapShot(revertId)
			return nil, quotaInit - quotaLeft, err
		} else {
			vm.StateDb.RevertToSnapShot(errorRevertId)
			return nil, quotaInit - quotaLeft, err
		}
	}
}

func (vm *VM) getHash(num uint64) types.Hash {
	// TODO get snapshot block hash
	return types.Hash{}
}

func run(vm *VM, contract *Contract, data []byte) (ret []byte, err error) {
	if len(contract.code) == 0 {
		return nil, nil
	}

	if vm.intPool == nil {
		vm.intPool = poolOfIntPools.get()
		defer func() {
			poolOfIntPools.put(vm.intPool)
			vm.intPool = nil
		}()
	}

	// TODO TBD
	vm.Depth++
	defer func() { vm.Depth-- }()

	vm.returnData = nil

	var (
		op    OpCode
		mem   = newMemory()
		stack = newStack()
		pc    = uint64(0)
		cost  uint64
	)
	contract.data = data

	for atomic.LoadInt32(&vm.abort) == 0 {
		currentPc := pc
		op = contract.GetOp(pc)
		operation := vm.instructionSet[op]
		if !operation.valid {
			return nil, fmt.Errorf("invalid opcode 0x%x", int(op))
		}
		if err := operation.validateStack(stack); err != nil {
			return nil, err
		}

		var memorySize uint64
		if operation.memorySize != nil {
			memSize, overflow := bigUint64(operation.memorySize(stack))
			if overflow {
				return nil, errGasUintOverflow
			}
			if memorySize, overflow = SafeMul(toWordSize(memSize), 32); overflow {
				return nil, errGasUintOverflow
			}
		}
		cost, err = operation.gasCost(vm, contract, stack, mem, memorySize)
		if err != nil || !contract.UseQuota(cost) {
			return nil, ErrOutOfQuota
		}
		if memorySize > 0 {
			mem.resize(memorySize)
		}

		res, err := operation.execute(&pc, vm, contract, mem, stack)

		if vm.Debug {
			fmt.Println("--------------------")
			fmt.Printf("op: %v, pc: %v\nstack: [%v]\nmemory: [%v]\n", opCodeToString[op], currentPc, stack.ToString(), mem.toString())
			fmt.Println("--------------------")
		}

		switch {
		case err != nil:
			return nil, err
		case operation.halts:
			return res, nil
		case operation.reverts:
			return res, errExecutionReverted
		case !operation.jumps:
			pc++
		}
	}
	return nil, nil
}
