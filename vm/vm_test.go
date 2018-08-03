package vm

import (
	"bytes"
	"encoding/hex"
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
	"testing"
)

func TestRun(t *testing.T) {
	vm := NewVM(Transaction{})
	vm.StateDb = &testDatabase{}
	vm.Debug = true
	// return 1+2
	inputdata, _ := hex.DecodeString("6001600201602080919052602090F3")
	contract := newContract(types.Address{}, types.Address{}, types.TokenTypeId{}, new(big.Int), inputdata)
	contract.setCallCode(types.Address{}, types.Hash{}, inputdata)
	ret, _, _ := run(vm, contract, 10000)
	expectedRet, _ := hex.DecodeString("03")
	expectedRet = leftPadBytes(expectedRet, 32)
	if bytes.Compare(ret, expectedRet) != 0 {
		t.Fatalf("expected [%v], get [%v]", expectedRet, ret)
	} else {
		t.Log("return [%v]", ret)
	}
}

func TestVM_CreateSend(t *testing.T) {
	inputdata, _ := hex.DecodeString("608060405260008055348015601357600080fd5b5060358060216000396000f3006080604052600080fd00a165627a7a723058207c31c74808fe0f95820eb3c48eac8e3e10ef27058dc6ca159b547fccde9290790029")
	vm := NewVM(Transaction{Depth: 1, TxType: 1, TokenTypeId: types.CreateTokenTypeId(), Amount: big.NewInt(10), Data: inputdata})
	vm.StateDb = &testDatabase{}
	vm.Debug = true
	addr, quotaUsed, _, _, err := vm.Create()
	empthAddress := types.Address{}
	if addr != empthAddress || quotaUsed != 58336 || err != nil {
		t.Fatalf("send create fail, %v %v %v", addr, quotaUsed, err)
	}
}

func TestVM_CreateReceive(t *testing.T) {
	inputdata, _ := hex.DecodeString("608060405260008055348015601357600080fd5b5060358060216000396000f3006080604052600080fd00a165627a7a723058207c31c74808fe0f95820eb3c48eac8e3e10ef27058dc6ca159b547fccde9290790029")
	vm := NewVM(Transaction{Depth: 1, TxType: 2, TokenTypeId: types.CreateTokenTypeId(), Amount: big.NewInt(0), Data: inputdata})
	vm.StateDb = &testDatabase{}
	vm.Debug = true
	addr, quotaUsed, _, _, err := vm.Create()
	empthAddress := types.Address{}
	if addr == empthAddress || quotaUsed != 74008 || err != nil {
		t.Fatalf("send create fail, %v %v %v", addr, quotaUsed, err)
	}
}
