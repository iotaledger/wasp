package wasmhost

// all predefined key id values should exactly match their counterpart values on the client!
// note that predefined key ids are negative values to distinguish them from indexes

const (
	KeyAccountId       = int32(-1)
	KeyAddress         = int32(-2)
	KeyBalances        = int32(-3)
	KeyBase58Bytes     = int32(-4)
	KeyBase58String    = int32(-5)
	KeyBlsAddress      = int32(-6)
	KeyBlsAggregate    = int32(-7)
	KeyBlsValid        = int32(-8)
	KeyCall            = int32(-9)
	KeyCaller          = int32(-10)
	KeyChainId         = int32(-11)
	KeyChainOwnerId    = int32(-12)
	KeyColor           = int32(-13)
	KeyContract        = int32(-14)
	KeyContractCreator = int32(-15)
	KeyDeploy          = int32(-16)
	KeyEd25519Address  = int32(-17)
	KeyEd25519Valid    = int32(-18)
	KeyEvent           = int32(-19)
	KeyExports         = int32(-20)
	KeyHashBlake2b     = int32(-21)
	KeyHashSha3        = int32(-22)
	KeyHname           = int32(-23)
	KeyIncoming        = int32(-24)
	KeyLength          = int32(-25)
	KeyLog             = int32(-26)
	KeyMaps            = int32(-27)
	KeyMinted          = int32(-28)
	KeyName            = int32(-29)
	KeyPanic           = int32(-30)
	KeyParams          = int32(-31)
	KeyPost            = int32(-32)
	KeyRandom          = int32(-33)
	KeyRequestId       = int32(-34)
	KeyResults         = int32(-35)
	KeyReturn          = int32(-36)
	KeyState           = int32(-37)
	KeyTimestamp       = int32(-38)
	KeyTrace           = int32(-39)
	KeyTransfers       = int32(-40)
	KeyUtility         = int32(-41)
	KeyValid           = int32(-42)

	// Treat this one like a version number. When anything changes
	// to the keys give this one a different value and make sure
	// that the client side is updated accordingly
	KeyZzzzzzz = int32(-43)
)

// associate names with predefined key ids
var keyMap = map[string]int32{
	"accountId":       KeyAccountId,
	"address":         KeyAddress,
	"balances":        KeyBalances,
	"base58Bytes":     KeyBase58Bytes,
	"base58String":    KeyBase58String,
	"blsAddress":      KeyBlsAddress,
	"blsAggregate":    KeyBlsAggregate,
	"blsValid":        KeyBlsValid,
	"call":            KeyCall,
	"caller":          KeyCaller,
	"chainId":         KeyChainId,
	"chainOwnerId":    KeyChainOwnerId,
	"color":           KeyColor,
	"contract":        KeyContract,
	"contractCreator": KeyContractCreator,
	"deploy":          KeyDeploy,
	"ed25519Address":  KeyEd25519Address,
	"ed25519Valid":    KeyEd25519Valid,
	"event":           KeyEvent,
	"exports":         KeyExports,
	"hashBlake2b":     KeyHashBlake2b,
	"hashSha3":        KeyHashSha3,
	"hname":           KeyHname,
	"incoming":        KeyIncoming,
	"length":          KeyLength,
	"log":             KeyLog,
	"maps":            KeyMaps,
	"minted":          KeyMinted,
	"name":            KeyName,
	"panic":           KeyPanic,
	"params":          KeyParams,
	"post":            KeyPost,
	"random":          KeyRandom,
	"requestId":       KeyRequestId,
	"results":         KeyResults,
	"return":          KeyReturn,
	"state":           KeyState,
	"timestamp":       KeyTimestamp,
	"trace":           KeyTrace,
	"transfers":       KeyTransfers,
	"utility":         KeyUtility,
	"valid":           KeyValid,
}
