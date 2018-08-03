package vm

import (
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
)

type Contract struct {
	caller    types.Address
	self      types.Address
	jumpdests destinations
	code      []byte
	codeHash  types.Hash
	codeAddr  types.Address
	data      []byte
	tokenId   types.TokenTypeId
	amount    *big.Int
}

func NewContract(caller types.Address, object types.Address, tokenId types.TokenTypeId, amount *big.Int, data []byte) *Contract {
	return &Contract{caller: caller, self: object, tokenId: tokenId, amount: amount, data: data, jumpdests: make(destinations)}
}

func (c *Contract) GetOp(n uint64) opCode {
	return opCode(c.GetByte(n))
}

func (c *Contract) GetByte(n uint64) byte {
	if n < uint64(len(c.code)) {
		return c.code[n]
	}

	return 0
}

func (c *Contract) Caller() types.Address {
	return c.caller
}

func (c *Contract) Address() types.Address {
	return c.self
}

func (c *Contract) SetCallCode(addr types.Address, hash types.Hash, code []byte) {
	c.code = code
	c.codeHash = hash
	c.codeAddr = addr
}
