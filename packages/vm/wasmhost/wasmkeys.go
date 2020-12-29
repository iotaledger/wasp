package wasmhost

// @formatter:off
const (
	KeyAgent       = int32(-1)
	KeyBalances    = KeyAgent       - 1
	KeyBase58      = KeyBalances    - 1
	KeyCaller      = KeyBase58      - 1
	KeyCalls       = KeyCaller      - 1
	KeyChain       = KeyCalls       - 1
	KeyChainOwner  = KeyChain       - 1
	KeyColor       = KeyChainOwner  - 1
	KeyContract    = KeyColor       - 1
	KeyCreator     = KeyContract    - 1
	KeyData        = KeyCreator     - 1
	KeyDelay       = KeyData        - 1
	KeyDescription = KeyDelay       - 1
	KeyEvent       = KeyDescription - 1
	KeyExports     = KeyEvent       - 1
	KeyFunction    = KeyExports     - 1
	KeyHash        = KeyFunction    - 1
	KeyId          = KeyHash        - 1
	KeyIncoming    = KeyId          - 1
	KeyLength      = KeyIncoming    - 1
	KeyLog         = KeyLength      - 1
	KeyLogs        = KeyLog         - 1
	KeyName        = KeyLogs        - 1
	KeyPanic       = KeyName        - 1
	KeyParams      = KeyPanic       - 1
	KeyPosts       = KeyParams      - 1
	KeyRandom      = KeyPosts       - 1
	KeyResults     = KeyRandom      - 1
	KeyState       = KeyResults     - 1
	KeyTimestamp   = KeyState       - 1
	KeyTrace       = KeyTimestamp   - 1
	KeyTransfers   = KeyTrace       - 1
	KeyUtility     = KeyTransfers   - 1
	KeyViews       = KeyUtility     - 1
	KeyZzzzzzz     = KeyViews       - 1
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
	"views":       KeyViews,
}
