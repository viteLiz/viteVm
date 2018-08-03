package vm

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto"
	"math/big"
)

func opStop(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	return nil, nil
}

func opAdd(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	U256(y.Add(x, y))

	vm.intPool.put(x)
	return nil, nil
}

func opMul(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	U256(y.Mul(x, y))

	vm.intPool.put(x)
	return nil, nil
}

func opSub(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	U256(y.Sub(x, y))

	vm.intPool.put(x)
	return nil, nil
}

func opDiv(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if y.Sign() != 0 {
		U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}

	vm.intPool.put(x)
	return nil, nil
}

func opSdiv(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := S256(stack.pop()), S256(stack.pop())
	res := vm.intPool.getZero()

	if y.Sign() == 0 || x.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() != y.Sign() {
			res.Div(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Div(x.Abs(x), y.Abs(y))
		}
		stack.push(U256(res))
	}

	vm.intPool.put(x, y)
	return nil, nil
}

func opMod(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if y.Sign() == 0 {
		stack.push(x.SetUint64(0))
	} else {
		stack.push(U256(x.Mod(x, y)))
	}

	vm.intPool.put(y)
	return nil, nil
}

func opSmod(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := S256(stack.pop()), S256(stack.pop())
	res := vm.intPool.getZero()

	if y.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() < 0 {
			res.Mod(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Mod(x.Abs(x), y.Abs(y))
		}
		stack.push(U256(res))
	}

	vm.intPool.put(x, y)
	return nil, nil
}

func opAddmod(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		x.Mod(x, z)
		stack.push(U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}

	vm.intPool.put(y, z)
	return nil, nil
}

func opMulmod(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		x.Mod(x, z)
		stack.push(U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}

	vm.intPool.put(y, z)
	return nil, nil
}

func opExp(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.pop()
	stack.push(Exp(base, exponent))

	vm.intPool.put(base, exponent)
	return nil, nil
}

func opSignExtend(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	back := stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := back.Lsh(big1, bit)
		mask.Sub(mask, big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}
		stack.push(U256(num))
	}

	vm.intPool.put(back)
	return nil, nil
}

func opLt(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	vm.intPool.put(x)
	return nil, nil
}

func opGt(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	vm.intPool.put(x)
	return nil, nil
}

func opSlt(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(1)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(0)

	default:
		if x.Cmp(y) < 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	vm.intPool.put(x)
	return nil, nil
}

func opSgt(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(0)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(1)

	default:
		if x.Cmp(y) > 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	vm.intPool.put(x)
	return nil, nil
}

func opEq(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	vm.intPool.put(x)
	return nil, nil
}

func opIszero(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x := stack.peek()
	if x.Sign() > 0 {
		x.SetUint64(0)
	} else {
		x.SetUint64(1)
	}
	return nil, nil
}

func opAnd(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.And(x, y))

	vm.intPool.put(y)
	return nil, nil
}

func opOr(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(x, y)

	vm.intPool.put(x)
	return nil, nil
}

func opXor(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(x, y)

	vm.intPool.put(x)
	return nil, nil
}

func opNot(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	x := stack.peek()
	U256(x.Not(x))
	return nil, nil
}

func opByte(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	if th.Cmp(big32) < 0 {
		b := Byte(val, 32, int(th.Int64()))
		val.SetUint64(uint64(b))
	} else {
		val.SetUint64(0)
	}

	vm.intPool.put(th)
	return nil, nil
}

func opSHL(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	shift, value := U256(stack.pop()), U256(stack.peek())
	if shift.Cmp(Big256) >= 0 {
		value.SetUint64(0)
	} else {
		U256(value.Lsh(value, uint(shift.Uint64())))
	}

	vm.intPool.put(shift)
	return nil, nil
}

func opSHR(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	shift, value := U256(stack.pop()), U256(stack.peek())
	if shift.Cmp(Big256) >= 0 {
		value.SetUint64(0)
	} else {
		U256(value.Rsh(value, uint(shift.Uint64())))
	}

	vm.intPool.put(shift)
	return nil, nil
}

func opSAR(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	shift, value := U256(stack.pop()), S256(stack.pop())
	if shift.Cmp(Big256) >= 0 {
		if value.Sign() > 0 {
			value.SetUint64(0)
		} else {
			value.SetInt64(-1)
		}
		stack.push(U256(value))
	} else {
		stack.push(U256(value.Rsh(value, uint(shift.Uint64()))))
	}

	vm.intPool.put(shift)
	return nil, nil
}

func opBlake2b(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	data := memory.get(offset.Int64(), size.Int64())
	hash := crypto.Hash256(data)
	stack.push(vm.intPool.get().SetBytes(hash))

	vm.intPool.put(offset, size)
	return nil, nil
}

func opAddress(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetBytes(contract.address.Bytes()))
	return nil, nil
}

func opBalance(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	addrBig, tokenTypeIdBig := stack.pop(), stack.pop()
	address, _ := types.BytesToAddress(addrBig.Bytes())
	tokenTypeId, _ := types.BytesToTokenTypeId(tokenTypeIdBig.Bytes())
	stack.push(vm.StateDb.GetBalance(address, tokenTypeId))

	vm.intPool.put(addrBig, tokenTypeIdBig)
	return nil, nil
}

func opCaller(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetBytes(contract.caller.Bytes()))
	return nil, nil
}

func opCallValue(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(contract.amount)
	return nil, nil
}

func opCallDataLoad(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetBytes(getDataBig(contract.data, stack.pop(), big32)))
	return nil, nil
}

func opCallDataSize(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetInt64(int64(len(contract.data))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	memory.set(memOffset.Uint64(), length.Uint64(), getDataBig(contract.data, dataOffset, length))

	vm.intPool.put(memOffset, dataOffset, length)
	return nil, nil
}

func opCodeSize(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetInt64(int64(len(contract.code))))
	return nil, nil
}

func opCodeCopy(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	memory.set(memOffset.Uint64(), length.Uint64(), getDataBig(contract.code, codeOffset, length))

	vm.intPool.put(memOffset, codeOffset, length)
	return nil, nil
}

func opBlockHash(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	num := stack.pop()
	n := vm.intPool.get().Sub(vm.SnapshotHeight, Big256)
	if num.Cmp(n) > 0 && num.Cmp(vm.SnapshotHeight) <= 0 {
		stack.push(vm.StateDb.GetHash(num.Uint64()).Big())
	} else {
		stack.push(vm.intPool.getZero())
	}

	vm.intPool.put(num, n)
	return nil, nil
}

func opTimestamp(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(U256(vm.intPool.get().Set(vm.SnapshotTimestamp)))
	return nil, nil
}

func opNumber(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(U256(vm.intPool.get().Set(vm.SnapshotHeight)))
	return nil, nil
}

func opPop(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	vm.intPool.put(stack.pop())
	return nil, nil
}

func opMload(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	offset := stack.pop()
	val := vm.intPool.get().SetBytes(memory.get(offset.Int64(), 32))
	stack.push(val)

	vm.intPool.put(offset)
	return nil, nil
}

func opMstore(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	// pop amount of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.set32(mStart.Uint64(), val)

	vm.intPool.put(mStart, val)
	return nil, nil
}

func opMstore8(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)

	return nil, nil
}

func opSLoad(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	loc := stack.peek()
	locHash, _ := types.BigToHash(loc)
	val := vm.StateDb.GetState(contract.address, locHash)
	loc.SetBytes(val.Bytes())
	return nil, nil
}

func opSStore(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	locBig, valBig := stack.pop(), stack.pop()
	loc, _ := types.BigToHash(locBig)
	val, _ := types.BigToHash(valBig)
	vm.StateDb.SetState(contract.address, loc, val)

	vm.intPool.put(locBig, valBig)
	return nil, nil
}

func opJump(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.jumpdests.has(contract.codeHash, contract.code, pos) {
		nop := contract.getOp(pos.Uint64())
		return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
	}
	*pc = pos.Uint64()

	vm.intPool.put(pos)
	return nil, nil
}

func opJumpi(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if cond.Sign() != 0 {
		if !contract.jumpdests.has(contract.codeHash, contract.code, pos) {
			nop := contract.getOp(pos.Uint64())
			return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	vm.intPool.put(pos, cond)
	return nil, nil
}

func opPc(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	stack.push(vm.intPool.get().SetInt64(int64(memory.len())))
	return nil, nil
}

func opJumpdest(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
		codeLen := len(contract.code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := vm.intPool.get()
		stack.push(integer.SetBytes(rightPadBytes(contract.code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
		stack.dup(vm.intPool, int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}

func makeLog(size int) executionFunc {
	return func(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
		topics := make([]types.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i], _ = types.BigToHash(stack.pop())
		}

		d := memory.get(mStart.Int64(), mSize.Int64())
		vm.StateDb.AddLog(&Log{
			Address: contract.address,
			Topics:  topics,
			Data:    d,
			Height:  vm.AccountHeight.Uint64(),
		})

		vm.intPool.put(mStart, mSize)
		return nil, nil
	}
}

func opReturn(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.getPtr(offset.Int64(), size.Int64())

	vm.intPool.put(offset, size)
	return ret, nil
}

func opRevert(pc *uint64, vm *VM, contract *contract, memory *memory, stack *stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.getPtr(offset.Int64(), size.Int64())

	vm.intPool.put(offset, size)
	return ret, nil
}
