package wasmhost

// @formatter:off
const (
	KeyAddress     = int32(-1)
	KeyAgent       = int32(-2)
	KeyBalances    = int32(-3)
	KeyBase58      = int32(-4)
	KeyCaller      = int32(-5)
	KeyCalls       = int32(-6)
	KeyChain       = int32(-7)
	KeyChainOwner  = int32(-8)
	KeyColor       = int32(-9)
	KeyContract    = int32(-10)
	KeyCreator     = int32(-11)
	KeyData        = int32(-12)
	KeyDelay       = int32(-13)
	KeyDeploys     = int32(-14)
	KeyDescription = int32(-15)
	KeyEvent       = int32(-16)
	KeyExports     = int32(-17)
	KeyFunction    = int32(-18)
	KeyHash        = int32(-19)
	KeyId          = int32(-20)
	KeyIncoming    = int32(-21)
	KeyLength      = int32(-22)
	KeyLog         = int32(-23)
	KeyLogs        = int32(-24)
	KeyMaps        = int32(-25)
	KeyName        = int32(-26)
	KeyPanic       = int32(-27)
	KeyParams      = int32(-28)
	KeyRandom      = int32(-29)
	KeyResults     = int32(-30)
	KeyState       = int32(-31)
	KeyTimestamp   = int32(-32)
	KeyTrace       = int32(-33)
	KeyTransfers   = int32(-34)
	KeyUtility     = int32(-35)
	// treat this like a version number when anything changes to the keys
	// make sure the client side in wasplib is updated accordingly
	KeyZzzzzzz = int32(-98)
)

// @formatter:on

var keyMap = map[string]int32{
	"address":     KeyAddress,
	"agent":       KeyAgent,
	"balances":    KeyBalances,
	"base58":      KeyBase58,
	"caller":      KeyCaller,
	"calls":       KeyCalls,
	"chain":       KeyChain,
	"chain_owner": KeyChainOwner,
	"color":       KeyColor,
	"contract":    KeyContract,
	"creator":     KeyCreator,
	"data":        KeyData,
	"delay":       KeyDelay,
	"deploys":     KeyDeploys,
	"description": KeyDescription,
	"event":       KeyEvent,
	"exports":     KeyExports,
	"function":    KeyFunction,
	"hash":        KeyHash,
	"id":          KeyId,
	"incoming":    KeyIncoming,
	"length":      KeyLength,
	"log":         KeyLog,
	"logs":        KeyLogs,
	"maps":        KeyMaps,
	"name":        KeyName,
	"panic":       KeyPanic,
	"params":      KeyParams,
	"random":      KeyRandom,
	"results":     KeyResults,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"trace":       KeyTrace,
	"transfers":   KeyTransfers,
	"utility":     KeyUtility,
}
