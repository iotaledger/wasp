// Package stardustwasp provides legacy keys for the stardust network
package stardustwasp

const (
	// KeyAllAccounts stores a map of <agentID> => true
	// Covered in: TestFoundries
	KeyAllAccounts = "a"

	// PrefixBaseTokens | <accountID> stores the amount of base tokens (big.Int)
	// Covered in: TestFoundries
	PrefixBaseTokens = "b"
	// PrefixBaseTokens | <accountID> stores a map of <nativeTokenID> => big.Int
	// Covered in: TestFoundries
	PrefixNativeTokens = "t"

	// L2TotalsAccount is the special <accountID> storing the total fungible tokens
	// controlled by the chain
	// Covered in: TestFoundries
	L2TotalsAccount = "*"

	// PrefixNFTs | <agentID> stores a map of <NFTID> => true
	// Covered in: TestDepositNFTWithMinStorageDeposit
	PrefixNFTs = "n"
	// PrefixNFTsByCollection | <agentID> | <collectionID> stores a map of <nftID> => true
	// Covered in: TestNFTMint
	// Covered in: TestDepositNFTWithMinStorageDeposit
	PrefixNFTsByCollection = "c"
	// PrefixNewlyMintedNFTs stores a map of <position in minted list> => <newly minted NFT> to be updated when the outputID is known
	// Covered in: TestNFTMint
	PrefixNewlyMintedNFTs = "N"
	// PrefixMintIDMap stores a map of <internal NFTID> => <NFTID> it is updated when the NFTID of newly minted nfts is known
	// Covered in: TestNFTMint
	PrefixMintIDMap = "M"
	// PrefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	// Covered in: TestFoundries
	PrefixFoundries = "f"

	// noCollection is the special <collectionID> used for storing NFTs that do not belong in a collection
	// Covered in: TestNFTMint
	NoCollection = "-"

	// KeyNonce stores a map of <agentID> => nonce (uint64)
	// Covered in: TestNFTMint
	KeyNonce = "m"

	// KeyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	// Covered in: TestFoundries
	KeyNativeTokenOutputMap = "TO"
	// KeyFoundryOutputRecords stores a map of <foundrySN> => foundryOutputRec
	// Covered in: TestFoundries
	KeyFoundryOutputRecords = "FO"
	// KeyNFTOutputRecords stores a map of <NFTID> => NFTOutputRec
	// Covered in: TestDepositNFTWithMinStorageDeposit
	KeyNFTOutputRecords = "NO"
	// KeyNFTOwner stores a map of <NFTID> => isc.AgentID
	// Covered in: TestDepositNFTWithMinStorageDeposit
	KeyNFTOwner = "NW"

	// KeyNewNativeTokens stores an array of <nativeTokenID>, containing the newly created native tokens that need filling out the OutputID
	// Covered in: TestFoundries
	KeyNewNativeTokens = "TN"
	// KeyNewFoundries stores an array of <foundrySN>, containing the newly created foundries that need filling out the OutputID
	// Covered in: TestFoundries
	KeyNewFoundries = "FN"
	// KeyNewNFTs stores an array of <NFTID>, containing the newly created NFTs that need filling out the OutputID
	// Covered in: TestDepositNFTWithMinStorageDeposit
	KeyNewNFTs = "NN"
)

var AccountsKeys = map[string]string{
	"a":  "keyAllAccounts",
	"b":  "prefixBaseTokens",
	"t":  "PrefixNativeTokens",
	"*":  "L2TotalsAccount",
	"n":  "PrefixNFTs",
	"c":  "PrefixNFTsByCollection",
	"N":  "prefixNewlyMintedNFTs",
	"M":  "prefixMintIDMap",
	"f":  "PrefixFoundries",
	"-":  "noCollection",
	"m":  "keyNonce",
	"TO": "keyNativeTokenOutputMap",
	"FO": "keyFoundryOutputRecords",
	"NO": "keyNFTOutputRecords",
	"NW": "keyNFTOwner",
	"TN": "keyNewNativeTokens",
	"FN": "keyNewFoundries",
	"NN": "keyNewNFTs",
}

var BlocklogKeys = map[string]string{
	"a": "PrefixBlockRegistry",
	"b": "prefixRequestLookupIndex",
	"c": "prefixRequestReceipts",
	"d": "prefixRequestEvents",
	"u": "prefixUnprocessableRequests",
	"U": "prefixNewUnprocessableRequests",
}

var ErrorKeys = map[string]string{
	"a": "prefixErrorTemplateMap",
}

var EvmKeys = map[string]string{
	"s": "keyEmulatorState",
	"m": "keyISCMagic",
}

var GovernanceKeys = map[string]string{
	"a":  "VarAllowedStateControllerAddresses",
	"r":  "VarRotateToAddress",
	"pa": "VarPayoutAgentID",
	"vs": "VarMinBaseTokensOnCommonAccount",
	"o":  "VarChainOwnerID",
	"n":  "VarChainOwnerIDDelegated",
	"g":  "VarGasFeePolicyBytes",
	"l":  "VarGasLimitsBytes",
	"an": "VarAccessNodes",
	"ac": "VarAccessNodeCandidates",
	"m":  "VarMaintenanceStatus",
	"md": "VarMetadata",
	"x":  "VarPublicURL",
	"b":  "VarBlockKeepAmount",
}

var RootKeys = map[string]string{
	"v": "VarSchemaVersion",
	"r": "VarContractRegistry",
	"a": "VarDeployPermissionsEnabled",
	"p": "VarDeployPermissions",
}
