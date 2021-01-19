package wasmhost

// @formatter:off
const (
	KeyAgent       = int32(-1)
	KeyBalances    = int32(-2)
	KeyBase58      = int32(-3)
	KeyCaller      = int32(-4)
	KeyCalls       = int32(-5)
	KeyChain       = int32(-6)
	KeyChainOwner  = int32(-7)
	KeyColor       = int32(-8)
	KeyContract    = int32(-9)
	KeyCreator     = int32(-10)
	KeyData        = int32(-11)
	KeyDelay       = int32(-12)
	KeyDeploys     = int32(-13)
	KeyDescription = int32(-14)
	KeyEvent       = int32(-15)
	KeyExports     = int32(-16)
	KeyFunction    = int32(-17)
	KeyHash        = int32(-18)
	KeyId          = int32(-19)
	KeyIncoming    = int32(-20)
	KeyLength      = int32(-21)
	KeyLog         = int32(-22)
	KeyLogs        = int32(-23)
	KeyMaps        = int32(-24)
	KeyName        = int32(-25)
	KeyPanic       = int32(-26)
	KeyParams      = int32(-27)
	KeyPosts       = int32(-28)
	KeyRandom      = int32(-29)
	KeyResults     = int32(-30)
	KeyState       = int32(-31)
	KeyTimestamp   = int32(-32)
	KeyTrace       = int32(-33)
	KeyTransfers   = int32(-34)
	KeyUtility     = int32(-35)
	// treat this like a version number when anything changes to the keys
	// make sure the client side in wasplib is updated accordingly
	KeyZzzzzzz     = int32(-100)
)
// @formatter:on

var keyMap = map[string]int32{
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
	"posts":       KeyPosts,
	"random":      KeyRandom,
	"results":     KeyResults,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"trace":       KeyTrace,
	"transfers":   KeyTransfers,
	"utility":     KeyUtility,
}
