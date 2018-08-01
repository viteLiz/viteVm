package vm

import "errors"

var (
	ErrOutOfQuota                  = errors.New("out of quota")
	ErrDepth                       = errors.New("max call Depth exceeded")
	ErrInsufficientBalance         = errors.New("insufficient balance for transfer")
	ErrContractAddressCreationFail = errors.New("contract address collision")
)

var (
	errGasUintOverflow   = errors.New("gas uint64 overflow")
	errExecutionReverted = errors.New("execution reverted")
)
