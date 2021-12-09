package gas

const (
	// Gas burned per 1 stored byte
	StoreByte = 2
	LogByte   = 1 // logging (issuing events) can be cheaper than state storage, as it can be pruned without breaking things

	// for the iscp.Sandbox.Utils interface
	UtilsHashingBlake2b              = 200
	UtilsHashingSha3                 = 300
	UtilsHashingHname                = 210
	UtilsBase58Encode                = 50
	UtilsBase58Decode                = 50
	UtilsED25519ValidSignature       = 500
	UtilsED25519AddressFromPublicKey = UtilsHashingBlake2b
	UtilsBLSValidSignature           = UtilsED25519ValidSignature * 40
	UtilsBLSAddressFromPublicKey     = UtilsHashingBlake2b
	UtilsBLSAggregateBLSSignature1   = 500

	// fixed gas budged to process NFT in the VM
	FixedGasBudgedNFT = 100

	// Sandbox calls
	ReadTheState        = 2 // (should we charge per byte read ?)
	GetRequestData      = 10
	GetCallerData       = 10
	GetIncomingTransfer = 10
	GetEntropy          = 10
	GetStateAnchorInfo  = 10
	LogEvent            = 10 // + byte cost of LogByte * payload size
	CallContract        = 100
	SendL1Request       = 1000

	// Constant initial cas cost to call Core Contracts entrypoints
	CoreRootDeployContract    = 5000
	CoreRootChangePermissions = 50   // grant/revoke/require permissions
	CoreAccounts              = 100  // withdrawal/deposit/harvest/sendTo
	CoreBlobStore             = 5000 // + byte cost of StoreByte * blob size
	CoreGovernance            = 1000 // all governance operations

	// Tokenization stuff yet to be implemented
	CreateTokenFoundry = 50000
	IssueNFT           = 2000
)
