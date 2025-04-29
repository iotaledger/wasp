// Package params defines parameters for webapi keys
package params

const (
	ParamAgentID              = "agentID"
	ParamChainID              = "chainID"
	ParamPeer                 = "peer"
	ParamRequestID            = "requestID"
	ParamSharedAddress        = "sharedAddress"
	ParamStateKey             = "stateKey"
	ParamUsername             = "username"
	ParamBlockIndexOrTrieRoot = "block"
	ParamBlockIndex           = "blockIndex"
	ParamPublicKey            = "publicKey"
)

const (
	DescriptionAgentID              = "AgentID (Hex Address for L1 accounts | Hex for EVM)"
	DescriptionChainID              = "ChainID (Hex Address)"
	DescriptionPeer                 = "Name or PubKey (hex) of the trusted peer"
	DescriptionRequestID            = "RequestID (Hex)"
	DescriptionSharedAddress        = "SharedAddress (Hex Address)"
	DescriptionStateKey             = "State Key (Hex)"
	DescriptionUsername             = "The username"
	DescriptionBlockIndexOrTrieRoot = "Block index or trie root"
)
