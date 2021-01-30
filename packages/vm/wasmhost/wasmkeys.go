package wasmhost

const (
	KeyAddress      = int32(-1)
	KeyAgent        = int32(-2)
	KeyBalances     = int32(-3)
	KeyBase58Bytes  = int32(-4)
	KeyBase58String = int32(-5)
	KeyCall         = int32(-6)
	KeyCaller       = int32(-7)
	KeyChain        = int32(-8)
	KeyChainOwner   = int32(-9)
	KeyColor        = int32(-10)
	KeyContract     = int32(-11)
	KeyCreator      = int32(-12)
	KeyData         = int32(-13)
	KeyDelay        = int32(-14)
	KeyDeploy       = int32(-15)
	KeyDescription  = int32(-16)
	KeyEvent        = int32(-17)
	KeyExports      = int32(-18)
	KeyFunction     = int32(-19)
	KeyHash         = int32(-20)
	KeyHname        = int32(-21)
	KeyId           = int32(-22)
	KeyIncoming     = int32(-23)
	KeyLength       = int32(-24)
	KeyLog          = int32(-25)
	KeyLogs         = int32(-26)
	KeyMaps         = int32(-27)
	KeyName         = int32(-28)
	KeyPanic        = int32(-29)
	KeyParams       = int32(-30)
	KeyPost         = int32(-31)
	KeyRandom       = int32(-32)
	KeyResults      = int32(-33)
	KeyReturn       = int32(-34)
	KeyState        = int32(-35)
	KeyTimestamp    = int32(-36)
	KeyTrace        = int32(-37)
	KeyTransfers    = int32(-38)
	KeyUtility      = int32(-39)

	// Treat this one like a version number. When anything changes
	// to the keys give this one a different value and make sure
	// the client side in wasplib is updated accordingly
	KeyZzzzzzz = int32(-99)
)

var keyMap = map[string]int32{
	"address":      KeyAddress,
	"agent":        KeyAgent,
	"balances":     KeyBalances,
	"base58decode": KeyBase58Bytes,
	"base58encode": KeyBase58String,
	"call":         KeyCall,
	"caller":       KeyCaller,
	"chain":        KeyChain,
	"chain_owner":  KeyChainOwner,
	"color":        KeyColor,
	"contract":     KeyContract,
	"creator":      KeyCreator,
	"data":         KeyData,
	"delay":        KeyDelay,
	"deploy":       KeyDeploy,
	"description":  KeyDescription,
	"event":        KeyEvent,
	"exports":      KeyExports,
	"function":     KeyFunction,
	"hash":         KeyHash,
	"hname":        KeyHname,
	"id":           KeyId,
	"incoming":     KeyIncoming,
	"length":       KeyLength,
	"log":          KeyLog,
	"logs":         KeyLogs,
	"maps":         KeyMaps,
	"name":         KeyName,
	"panic":        KeyPanic,
	"params":       KeyParams,
	"post":         KeyPost,
	"random":       KeyRandom,
	"results":      KeyResults,
	"return":       KeyReturn,
	"state":        KeyState,
	"timestamp":    KeyTimestamp,
	"trace":        KeyTrace,
	"transfers":    KeyTransfers,
	"utility":      KeyUtility,
}
