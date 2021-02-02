package wasmhost

const (
	KeyAddress         = int32(-1)
	KeyBalances        = int32(-2)
	KeyBase58Bytes     = int32(-3)
	KeyBase58String    = int32(-4)
	KeyCall            = int32(-5)
	KeyCaller          = int32(-6)
	KeyChainOwnerId    = int32(-7)
	KeyColor           = int32(-8)
	KeyContractCreator = int32(-9)
	KeyContractId      = int32(-10)
	KeyData            = int32(-11)
	KeyDeploy          = int32(-12)
	KeyEvent           = int32(-13)
	KeyExports         = int32(-14)
	KeyHashBlake2b     = int32(-15)
	KeyHashSha3        = int32(-16)
	KeyHname           = int32(-17)
	KeyIncoming        = int32(-18)
	KeyLength          = int32(-19)
	KeyLog             = int32(-20)
	KeyLogs            = int32(-21)
	KeyMaps            = int32(-22)
	KeyName            = int32(-23)
	KeyPanic           = int32(-24)
	KeyParams          = int32(-25)
	KeyPost            = int32(-26)
	KeyRandom          = int32(-27)
	KeyResults         = int32(-28)
	KeyReturn          = int32(-29)
	KeyState           = int32(-30)
	KeyTimestamp       = int32(-31)
	KeyTrace           = int32(-32)
	KeyTransfers       = int32(-33)
	KeyUtility         = int32(-34)
	KeyValid           = int32(-35)
	KeyValidEd25519    = int32(-36)

	// Treat this one like a version number. When anything changes
	// to the keys give this one a different value and make sure
	// the client side in wasplib is updated accordingly
	KeyZzzzzzz = int32(-95)
)

var keyMap = map[string]int32{
	"address":         KeyAddress,
	"balances":        KeyBalances,
	"base58Bytes":     KeyBase58Bytes,
	"base58String":    KeyBase58String,
	"call":            KeyCall,
	"caller":          KeyCaller,
	"chainOwnerId":    KeyChainOwnerId,
	"color":           KeyColor,
	"contractCreator": KeyContractCreator,
	"contractId":      KeyContractId,
	"data":            KeyData,
	"deploy":          KeyDeploy,
	"event":           KeyEvent,
	"exports":         KeyExports,
	"hashBlake2b":     KeyHashBlake2b,
	"hashSha3":        KeyHashSha3,
	"hname":           KeyHname,
	"incoming":        KeyIncoming,
	"length":          KeyLength,
	"log":             KeyLog,
	"logs":            KeyLogs,
	"maps":            KeyMaps,
	"name":            KeyName,
	"panic":           KeyPanic,
	"params":          KeyParams,
	"post":            KeyPost,
	"random":          KeyRandom,
	"results":         KeyResults,
	"return":          KeyReturn,
	"state":           KeyState,
	"timestamp":       KeyTimestamp,
	"trace":           KeyTrace,
	"transfers":       KeyTransfers,
	"utility":         KeyUtility,
	"valid":           KeyValid,
	"validEd25519":    KeyValidEd25519,
}
