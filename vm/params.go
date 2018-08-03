package vm

const (
	quickStepGas   uint64 = 2
	fastestStepGas uint64 = 3
	fastStepGas    uint64 = 5
	midStepGas     uint64 = 8
	slowStepGas    uint64 = 10
	extStepGas     uint64 = 20
	balanceGas     uint64 = 20
	sLoadGas       uint64 = 50
	expByteGas     uint64 = 10
	//GasLimitBoundDivisor uint64 = 1024    // The bound divisor of the gas limit, used in update calculations.
	//MinGasLimit          uint64 = 5000    // Minimum the gas limit may ever be.
	//GenesisGasLimit      uint64 = 4712388 // Gas limit of the Genesis block.
	//
	//MaximumExtraDataSize  uint64 = 32    // Maximum size extra data may be after Genesis.
	//CallValueTransferGas  uint64 = 9000  // Paid for CALL when the amount transfer is non-zero.
	//CallNewAccountGas     uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	txGas                 uint64 = 21000 // Per transaction not creating a contract.
	txGasContractCreation uint64 = 53000 // Per transaction that creates a contract.
	txDataZeroGas         uint64 = 4     // Per byte of data attached to a transaction that equals zero.
	txDataNonZeroGas      uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero.
	quadCoeffDiv          uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	logGas                uint64 = 375   // Per LOG* operation.
	logTopicGas           uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicGas, LOG4 incurs 4 * c_txLogTopicGas.
	logDataGas            uint64 = 8     // Per byte in a LOG* operation's data.
	//CallStipend           uint64 = 2300  // Free gas given at beginning of call.
	blake2bGas      uint64 = 30    // Once per Blake2b operation.
	blake2bWordGas  uint64 = 6     // Once per word of the Blake2b operation's data.
	sstoreSetGas    uint64 = 20000 // Once per SSTORE operation
	sstoreResetGas  uint64 = 5000  // Once per SSTORE operation if the zeroness changes from zero.
	sstoreClearGas  uint64 = 5000  // Once per SSTORE operation if the zeroness doesn't change.
	sstoreRefundGas uint64 = 15000 // Once per SSTORE operation if the zeroness changes to zero.
	jumpdestGas     uint64 = 1     // Jumpdest gas cost.
	//EpochDuration    uint64 = 30000 // Duration between proof-of-work epochs.
	//CallGas          uint64 = 40    // Once per CALL operation & message call transaction.
	contractCodeGas uint64 = 200  // Per byte in contract code
	callCreateDepth uint64 = 1024 // Maximum Depth of call/create stack.
	copyGas         uint64 = 3    //
	stackLimit      uint64 = 1024 // Maximum size of VM stack allowed.
	//TierStepGas      uint64 = 0     // Once per operation, for a selection of them.
	//SuicideRefundGas uint64 = 24000 // Refunded following a suicide operation.
	memoryGas uint64 = 3 // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.
	//
	//MaxCodeSize = 24576 // Maximum bytecode to permit for a contract
	//
	//// Precompiled contract gas prices
	//
	//EcrecoverGas            uint64 = 3000   // Elliptic curve sender recovery gas price
	//Sha256BaseGas           uint64 = 60     // Base price for a SHA256 operation
	//Sha256PerWordGas        uint64 = 12     // Per-word price for a SHA256 operation
	//Ripemd160BaseGas        uint64 = 600    // Base price for a RIPEMD160 operation
	//Ripemd160PerWordGas     uint64 = 120    // Per-word price for a RIPEMD160 operation
	//IdentityBaseGas         uint64 = 15     // Base price for a data copy operation
	//IdentityPerWordGas      uint64 = 3      // Per-work price for a data copy operation
	//ModExpQuadCoeffDiv      uint64 = 20     // Divisor for the quadratic particle of the big int modular exponentiation
	//Bn256AddGas             uint64 = 500    // Gas needed for an elliptic curve addition
	//Bn256ScalarMulGas       uint64 = 40000  // Gas needed for an elliptic curve scalar multiplication
	//Bn256PairingBaseGas     uint64 = 100000 // Base price for an elliptic curve pairing check
	//Bn256PairingPerPointGas uint64 = 80000  // Per-point price for an elliptic curve pairing check
)
