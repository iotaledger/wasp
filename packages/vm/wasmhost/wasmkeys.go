package wasmhost

// all predefined key id values should exactly match their counterpart values on the client!
// note that predefined key ids are negative values to distinguish them from indexes
// this allows us to use them to initiate special functionality
// for example array[KeyLength] returns the array length

const (
	KeyAccountID       = int32(-1)
	KeyAddress         = int32(-2)
	KeyBalances        = int32(-3)
	KeyBase58Decode    = int32(-4)
	KeyBase58Encode    = int32(-5)
	KeyBlsAddress      = int32(-6)
	KeyBlsAggregate    = int32(-7)
	KeyBlsValid        = int32(-8)
	KeyCall            = int32(-9)
	KeyCaller          = int32(-10)
	KeyChainID         = int32(-11)
	KeyChainOwnerID    = int32(-12)
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
	KeyPanic           = int32(-29)
	KeyParams          = int32(-30)
	KeyPost            = int32(-31)
	KeyRandom          = int32(-32)
	KeyRequestID       = int32(-33)
	KeyResults         = int32(-34)
	KeyReturn          = int32(-35)
	KeyState           = int32(-36)
	KeyTimestamp       = int32(-37)
	KeyTrace           = int32(-38)
	KeyTransfers       = int32(-39)
	KeyUtility         = int32(-40)

	// KeyZzzzzzz is treated like a version number.
	// When anything changes to the keys give this one a different value
	// and make sure that the client side is updated accordingly
	KeyZzzzzzz = int32(-41)
)

// associate names with predefined key ids
var predefinedKeyMap = map[string]int32{
	"$accountID":       KeyAccountID,
	"$address":         KeyAddress,
	"$balances":        KeyBalances,
	"$base58Bytes":     KeyBase58Decode,
	"$base58String":    KeyBase58Encode,
	"$blsAddress":      KeyBlsAddress,
	"$blsAggregate":    KeyBlsAggregate,
	"$blsValid":        KeyBlsValid,
	"$call":            KeyCall,
	"$caller":          KeyCaller,
	"$chainID":         KeyChainID,
	"$chainOwnerID":    KeyChainOwnerID,
	"$color":           KeyColor,
	"$contract":        KeyContract,
	"$contractCreator": KeyContractCreator,
	"$deploy":          KeyDeploy,
	"$ed25519Address":  KeyEd25519Address,
	"$ed25519Valid":    KeyEd25519Valid,
	"$event":           KeyEvent,
	"$exports":         KeyExports,
	"$hashBlake2b":     KeyHashBlake2b,
	"$hashSha3":        KeyHashSha3,
	"$hname":           KeyHname,
	"$incoming":        KeyIncoming,
	"$length":          KeyLength,
	"$log":             KeyLog,
	"$maps":            KeyMaps,
	"$minted":          KeyMinted,
	"$panic":           KeyPanic,
	"$params":          KeyParams,
	"$post":            KeyPost,
	"$random":          KeyRandom,
	"$requestID":       KeyRequestID,
	"$results":         KeyResults,
	"$return":          KeyReturn,
	"$state":           KeyState,
	"$timestamp":       KeyTimestamp,
	"$trace":           KeyTrace,
	"$transfers":       KeyTransfers,
	"$utility":         KeyUtility,
}

var predefinedKeys = initKeyMap()

func initKeyMap() [][]byte {
	keys := make([][]byte, len(predefinedKeyMap)+1)
	for k, v := range predefinedKeyMap {
		keys[-v] = []byte(k)
	}
	return keys
}
