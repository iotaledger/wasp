package gas

var (
	// Gas burned per 1 stored byte
	StoreByte  = uint64(2)
	StoreBytes = func(length int) uint64 { return uint64(length) * StoreByte }
	LogByte    = uint64(1) // logging (issuing events) can be cheaper than state storage, as it can be pruned without breaking things

	// for the iscp.Sandbox.Utils interface
	UtilsHashingBlake2b              = uint64(200)
	UtilsHashingSha3                 = uint64(300)
	UtilsHashingHname                = uint64(210)
	UtilsBase58Encode                = uint64(50)
	UtilsBase58Decode                = uint64(50)
	UtilsED25519ValidSignature       = uint64(500)
	UtilsED25519AddressFromPublicKey = UtilsHashingBlake2b
	UtilsBLSValidSignature           = UtilsED25519ValidSignature * 40
	UtilsBLSAddressFromPublicKey     = UtilsHashingBlake2b
	UtilsBLSAggregateBLSSignature1   = uint64(500)

	// fixed gas budged to process NFT in the VM
	FixedGasBudgedNFT = uint64(100)

	// Sandbox calls
	ReadTheState        = func(length int) uint64 { return 10 + uint64(length/10) } // 2 + 0.1 per byte
	GetRequestData      = uint64(10)
	GetContractContext  = uint64(10) // contract accountID, balances, current timestamp, etc
	GetCallerData       = uint64(10)
	GetIncomingTransfer = uint64(10)
	GetEntropy          = uint64(10)
	GetStateAnchorInfo  = uint64(10)
	LogEvent            = func(msg string) uint64 { return 10 + uint64(len([]byte(msg)))*LogByte }
	CallContract        = uint64(100)
	SendL1Request       = uint64(1000)

	// Constant initial cas cost to call Core Contracts entrypoints
	CoreRootDeployContract    = uint64(5000)
	CoreRootChangePermissions = uint64(50)   // grant/revoke/require permissions
	CoreAccounts              = uint64(100)  // withdrawal/deposit/harvest/sendTo
	CoreBlobStore             = uint64(5000) // + byte cost of StoreByte * blob size
	CoreGovernance            = uint64(1000) // all governance operations

	// Tokenization stuff yet to be implemented
	CreateTokenFoundry = uint64(50000)
	IssueNFT           = uint64(2000)
)
