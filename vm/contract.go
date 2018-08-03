package vm

import (
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
)

type contract struct {
	caller    types.Address
	address   types.Address
	jumpdests destinations
	code      []byte
	codeHash  types.Hash
	codeAddr  types.Address
	data      []byte
	tokenId   types.TokenTypeId
	amount    *big.Int
}

func newContract(caller types.Address, address types.Address, tokenId types.TokenTypeId, amount *big.Int, data []byte) *contract {
	return &contract{caller: caller, address: address, tokenId: tokenId, amount: amount, data: data, jumpdests: make(destinations)}
}

func (c *contract) getOp(n uint64) opCode {
	return opCode(c.getByte(n))
}

func (c *contract) getByte(n uint64) byte {
	if n < uint64(len(c.code)) {
		return c.code[n]
	}

	return 0
}

func (c *contract) setCallCode(addr types.Address, hash types.Hash, code []byte) {
	c.code = code
	c.codeHash = hash
	c.codeAddr = addr
}
