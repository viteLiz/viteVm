package vm

import (
	"github.com/vitelabs/go-vite/common/types"
)

const (
	GasQuickStep   uint64 = 2
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 8
	GasSlowStep    uint64 = 10
	GasExtStep     uint64 = 20
	GasBalance     uint64 = 20
	GasSLoad       uint64 = 50

	GasReturn       uint64 = 0
	GasStop         uint64 = 0
	GasContractByte uint64 = 200

	ExpByte uint64 = 10
)

// memoryGasCosts calculates the quadratic gas for memory expansion. It does so
// only for the memory region that is expanded, not the total memory.
func memoryGasCost(mem *memory, newMemSize uint64) (uint64, error) {

	if newMemSize == 0 {
		return 0, nil
	}
	// The maximum that will fit in a uint64 is max_word_count - 1
	// anything above that will result in an overflow.
	// Additionally, a newMemSize which results in a
	// newMemSizeWords larger than 0x7ffffffff will cause the square operation
	// to overflow.
	// The constant ç is the highest number that can be used without
	// overflowing the gas calculation
	if newMemSize > 0xffffffffe0 {
		return 0, errGasUintOverflow
	}

	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * MemoryGas
		quadCoef := square / QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastGasCost
		mem.lastGasCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

func constGasFunc(gas uint64) gasFunc {
	return func(vm *VM, contrac *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
		return gas, nil
	}
}

func gasExp(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	expByteLen := uint64((stack.back(1).BitLen() + 7) / 8)

	var (
		gas      = expByteLen * ExpByte // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = SafeAdd(gas, GasSlowStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasBlake2b(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	if gas, overflow = SafeAdd(gas, Blake2bGas); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.back(1))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = SafeMul(toWordSize(wordGas), Blake2bWordGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCallDataCopy(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	words, overflow := bigUint64(stack.back(2))
	if overflow {
		return 0, errGasUintOverflow
	}

	if words, overflow = SafeMul(toWordSize(words), CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = SafeAdd(gas, words); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCodeCopy(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.back(2))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = SafeMul(toWordSize(wordGas), CopyGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMLoad(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore8(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasSStore(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	var (
		y    = stack.back(1)
		x, _ = types.BigToHash(stack.back(0))
		val  = vm.StateDb.GetState(contract.Address(), x)
	)
	if val == (types.Hash{}) && y.Sign() != 0 {
		return SstoreSetGas, nil
	} else if val != (types.Hash{}) && y.Sign() == 0 {
		vm.StateDb.AddRefund(SstoreRefundGas)
		return SstoreClearGas, nil
	} else {
		return SstoreResetGas, nil
	}
}

func gasPush(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasDup(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasSwap(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func makeGasLog(n uint64) gasFunc {
	return func(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
		requestedSize, overflow := bigUint64(stack.back(1))
		if overflow {
			return 0, errGasUintOverflow
		}

		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return 0, err
		}

		if gas, overflow = SafeAdd(gas, LogGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = SafeAdd(gas, n*LogTopicGas); overflow {
			return 0, errGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = SafeMul(requestedSize, LogDataGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = SafeAdd(gas, memorySizeGas); overflow {
			return 0, errGasUintOverflow
		}
		return gas, nil
	}
}

func gasReturn(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

func gasRevert(vm *VM, contract *Contract, stack *stack, mem *memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

func intrinsicGasCost(data []byte, isCreate bool) (uint64, error) {
	var gas uint64
	if isCreate {
		gas = TxGasContractCreation
	} else {
		gas = TxGas
	}
	if len(data) > 0 {
		var nonZeroByteCount uint64
		for _, byteCode := range data {
			if byteCode != 0 {
				nonZeroByteCount++
			}
		}
		if (maxUint64-gas)/TxDataNonZeroGas < nonZeroByteCount {
			return 0, errGasUintOverflow
		}
		gas += nonZeroByteCount * TxDataNonZeroGas

		zeroByteCount := uint64(len(data)) - nonZeroByteCount
		if (maxUint64-gas)/TxDataZeroGas < zeroByteCount {
			return 0, errGasUintOverflow
		}
		gas += zeroByteCount * TxDataZeroGas
	}
	return gas, nil
}
