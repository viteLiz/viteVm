package vm

import (
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
)

type MethodSelector [4]byte

type Contract struct {
	caller     types.Address
	self       types.Address
	jumpdests  destinations
	code       []byte
	codeHash   types.Hash
	codeAddr   types.Address
	data       []byte
	quotaLeft  uint64
	tokenId    types.TokenTypeId
	amount     *big.Int
	quotaLimit map[MethodSelector]uint64
}

func NewContract(caller types.Address, object types.Address, tokenId types.TokenTypeId, amount *big.Int, quota uint64, parentJumpDests destinations) *Contract {
	c := &Contract{caller: caller, self: object, tokenId: tokenId, amount: amount, quotaLeft: quota}

	if parentJumpDests != nil {
		c.jumpdests = parentJumpDests
	} else {
		c.jumpdests = make(destinations)
	}

	return c
}

func (c *Contract) GetOp(n uint64) OpCode {
	return OpCode(c.GetByte(n))
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

func (c *Contract) UseQuota(gas uint64) (ok bool) {
	if c.quotaLeft < gas {
		return false
	}
	c.quotaLeft -= gas
	return true
}
