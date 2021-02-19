package wasmhost

const (
	KeyAddress         = int32(-1)
	KeyBalances        = int32(-2)
	KeyBase58Bytes     = int32(-3)
	KeyBase58String    = int32(-4)
	KeyBlsAddress      = int32(-5)
	KeyBlsAggregate    = int32(-6)
	KeyBlsValid        = int32(-7)
	KeyCall            = int32(-8)
	KeyCaller          = int32(-9)
	KeyChainOwnerId    = int32(-10)
	KeyColor           = int32(-11)
	KeyContractCreator = int32(-12)
	KeyContractId      = int32(-13)
	KeyDeploy          = int32(-14)
	KeyEd25519Address  = int32(-15)
	KeyEd25519Valid    = int32(-16)
	KeyEvent           = int32(-17)
	KeyExports         = int32(-18)
	KeyHashBlake2b     = int32(-19)
	KeyHashSha3        = int32(-20)
	KeyHname           = int32(-21)
	KeyIncoming        = int32(-22)
	KeyLength          = int32(-23)
	KeyLog             = int32(-24)
	KeyMaps            = int32(-25)
	KeyName            = int32(-26)
	KeyPanic           = int32(-27)
	KeyParams          = int32(-28)
	KeyPost            = int32(-29)
	KeyRandom          = int32(-30)
	KeyResults         = int32(-31)
	KeyReturn          = int32(-32)
	KeyState           = int32(-33)
	KeyTimestamp       = int32(-34)
	KeyTrace           = int32(-35)
	KeyTransfers       = int32(-36)
	KeyUtility         = int32(-37)
	KeyValid           = int32(-38)

	// Treat this one like a version number. When anything changes
	// to the keys give this one a different value and make sure
	// the client side in wasplib is updated accordingly
	KeyZzzzzzz = int32(-39)
)

var keyMap = map[string]int32{
	"address":         KeyAddress,
	"balances":        KeyBalances,
	"base58Bytes":     KeyBase58Bytes,
	"base58String":    KeyBase58String,
	"blsAddress":        KeyBlsAddress,
	"blsAggregate":    KeyBlsAggregate,
	"blsValid":        KeyBlsValid,
	"call":            KeyCall,
	"caller":          KeyCaller,
	"chainOwnerId":    KeyChainOwnerId,
	"color":           KeyColor,
	"contractCreator": KeyContractCreator,
	"contractId":      KeyContractId,
	"deploy":          KeyDeploy,
	"ed25519Address":    KeyEd25519Address,
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
}
