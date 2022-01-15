package gas

// Gas burned per 1 stored byte
const (
	StoreByte = uint64(1)
	LogByte   = uint64(1) // logging (issuing events) can be cheaper than state storage, as it can be pruned without breaking things
)

func StoreBytes(length int) uint64 { return uint64(length) * StoreByte }

// for the iscp.Sandbox.Utils interface
const (
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
)

// Sandbox calls
const (
	GetRequest         = uint64(10)
	GetContractContext = uint64(10)
	GetCallerData      = uint64(10)
	GetAllowance       = uint64(10)
	GetStateAnchorInfo = uint64(10)
	GetBalance         = uint64(20)
)

func ReadTheState(length int) uint64 { return 10 + uint64(length/10) } // 2 + 0.1 per byte
func LogEvent(msg string) uint64     { return 10 + uint64(len([]byte(msg)))*LogByte }

const (
	CallContract      = uint64(100)
	EmitEventFixed    = uint64(500)
	TransferAllowance = uint64(500) // must be parametrized
	NotFoundTarget    = uint64(100)
	SendL1Request     = uint64(10_000)
	MinGasPerBlob     = uint64(1000)
)

// Constant initial cas cost to call Core Contracts entry points
const (
	CoreRootDeployContract    = uint64(5000)
	CoreRootChangePermissions = uint64(50)   // grant/revoke/require permissions
	CoreAccounts              = uint64(100)  // withdrawal/deposit/harvest/sendTo
	CoreBlobStore             = uint64(5000) // + byte cost of StoreByte * blob size
	CoreGovernance            = uint64(1000) // all governance operations
)

// Tokenization stuff yet to be implemented
const (
	CreateTokenFoundry = uint64(50000)
	IssueNFT           = uint64(2000)
)
