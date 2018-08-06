package vm

import (
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
)

type Database interface {
	GetBalance(addr types.Address, tokenTypeId types.TokenTypeId) *big.Int
	SubBalance(addr types.Address, tokenTypeId types.TokenTypeId, amount *big.Int)
	AddBalance(addr types.Address, tokenTypeId types.TokenTypeId, amount *big.Int)

	Snapshot() int
	RevertToSnapShot(revertId int)

	IsExistAddress(addr types.Address) bool

	CreateAccount(addr types.Address)
	DeleteAccount(addr types.Address)

	SetContractCode(addr types.Address, code []byte)
	GetContractCode(addr types.Address) []byte
	GetContractCodeSize(addr types.Address) int
	GetContractCodeHash(addr types.Address) types.Hash

	GetState(addr types.Address, loc types.Hash) types.Hash
	SetState(addr types.Address, loc types.Hash, value types.Hash)
	GetStatesString(addr types.Address) string

	GetHash(num uint64) types.Hash
}

type testDatabase struct{}

func (db *testDatabase) GetBalance(addr types.Address, tokenTypeId types.TokenTypeId) *big.Int {
	return big.NewInt(1000)
}
func (db *testDatabase) SubBalance(addr types.Address, tokenTypeId types.TokenTypeId, amount *big.Int) {
}
func (db *testDatabase) AddBalance(addr types.Address, tokenTypeId types.TokenTypeId, amount *big.Int) {
}
func (db *testDatabase) Snapshot() int                                                 { return 0 }
func (db *testDatabase) RevertToSnapShot(revertId int)                                 {}
func (db *testDatabase) IsExistAddress(addr types.Address) bool                        { return false }
func (db *testDatabase) CreateAccount(addr types.Address)                              {}
func (db *testDatabase) DeleteAccount(addr types.Address)                              {}
func (db *testDatabase) SetContractCode(addr types.Address, code []byte)               {}
func (db *testDatabase) GetContractCode(addr types.Address) []byte                     { return nil }
func (db *testDatabase) GetContractCodeSize(addr types.Address) int                    { return 0 }
func (db *testDatabase) GetContractCodeHash(addr types.Address) types.Hash             { return types.Hash{} }
func (db *testDatabase) GetState(addr types.Address, loc types.Hash) types.Hash        { return types.Hash{} }
func (db *testDatabase) SetState(addr types.Address, loc types.Hash, value types.Hash) {}
func (db *testDatabase) GetStatesString(addr types.Address) string                     { return "" }
func (db *testDatabase) GetHash(num uint64) types.Hash                                 { return types.Hash{} }
