package iotaconn

const (
	AlphanetEndpointURL = "https://api.iota-rebased-alphanet.iota.cafe"
	LocalnetEndpointURL = "http://localhost:9000"

	// Can be uncommented once a public Test/Mainnet is online
	//DevnetEndpointURL   = "https://api.iota-rebased-alphanet.iota.cafe"
	//MainnetEndpointURL  = "https://api.iota-rebased-alphanet.iota.cafe"

	AlphanetWebsocketEndpointURL = "wss://api.iotago-rebased-alphanet.iotago.cafe/websocket"
	LocalnetWebsocketEndpointURL = "ws://localhost:9000"

	// Can be uncommented once a public Test/Mainnet is online
	//DevnetWebsocketEndpointURL   = "wss://api.iotago-rebased-alphanet.iotago.cafe/websocket"
	//MainnetWebsocketEndpointURL  = "wss://api.iotago-rebased-alphanet.iotago.cafe/websocket"

	AlphanetFaucetURL = "https://faucet.iota-rebased-alphanet.iota.cafe/gas"
	LocalnetFaucetURL = "http://localhost:9123/gas"

	// Can be uncommented once a public Test/Mainnet is online
	//DevnetFaucetURL   = "https://faucet.iota-rebased-alphanet.iota.cafe/gas"
)

const (
	ChainIdentifierAlphanet = "5d9dbb07"

	// Can be uncommented once a public Test/Mainnet is online
	ChainIdentifierSuiDevnet  = "f4950f51"
	ChainIdentifierSuiTestnet = "4c78adac"
	ChainIdentifierSuiMainnet = "35834a8a"

	// localnet doesn't have a fixed ChainIdentifier
	// ChainIdentifierLocalnet
)

func FaucetURL(apiURL string) string {
	switch apiURL {
	case AlphanetEndpointURL:
		return AlphanetFaucetURL
	case LocalnetEndpointURL:
		return LocalnetFaucetURL
	default:
		panic("unspecified FaucetURL")
	}
}
