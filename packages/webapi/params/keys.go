package params

const (
	ParamAgentID       = "agentID"
	ParamBlobHash      = "blobHash"
	ParamBlockIndex    = "blockIndex"
	ParamChainID       = "chainID"
	ParamContractHName = "contractHname"
	ParamFieldKey      = "fieldKey"
	ParamNFTID         = "nftID"
	ParamPeer          = "peer"
	ParamPublicKey     = "publicKey"
	ParamRequestID     = "requestID"
	ParamSharedAddress = "sharedAddress"
	ParamStateKey      = "stateKey"
	ParamTxHash        = "txHash"
	ParamUsername      = "username"
)

const (
	DescriptionAgentID       = "AgentID (Bech32 for WasmVM | Hex for EVM)"
	DescriptionBlobHash      = "BlobHash (Hex)"
	DescriptionChainID       = "ChainID (Bech32)"
	DescriptionContractHName = "The contract hname (Hex)"
	DescriptionFieldKey      = "FieldKey (String)"
	DescriptionNFTID         = "NFT ID (Hex)"
	DescriptionPeer          = "Name or PubKey (hex) of the trusted peer"
	DescriptionRequestID     = "RequestID (Hex)"
	DescriptionSharedAddress = "SharedAddress (Bech32)"
	DescriptionStateKey      = "State Key (Hex)"
	DescriptionTxHash        = "Transaction hash (Hex)"
	DescriptionUsername      = "The username"
)
