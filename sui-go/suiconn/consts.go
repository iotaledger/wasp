package suiconn

const (
	DevnetEndpointURL   = "https://fullnode.devnet.sui.io"
	TestnetEndpointURL  = "https://fullnode.testnet.sui.io"
	MainnetEndpointURL  = "https://fullnode.mainnet.sui.io"
	LocalnetEndpointURL = "http://localhost:9000"

	DevnetWebsocketEndpointURL   = "wss://rpc.devnet.sui.io:443"
	TestnetWebsocketEndpointURL  = "wss://rpc.testnet.sui.io:443"
	MainnetWebsocketEndpointURL  = "wss://rpc.mainnet.sui.io:443"
	LocalnetWebsocketEndpointURL = "ws://localhost:9000"

	DevnetFaucetURL   = "https://faucet.devnet.sui.io/v1/gas"
	TestnetFaucetURL  = "https://faucet.testnet.sui.io/v1/gas"
	LocalnetFaucetURL = "http://localhost:9123/gas"
)

const (
	ChainIdentifierSuiDevnet  = "f4950f51"
	ChainIdentifierSuiTestnet = "4c78adac"
	ChainIdentifierSuiMainnet = "35834a8a"

	// localnet doesn't have a fixed ChainIdentifier
	// ChainIdentifierLocalnet
)

func FaucetURL(apiURL string) string {
	switch apiURL {
	case TestnetEndpointURL:
		return TestnetFaucetURL
	case DevnetEndpointURL:
		return DevnetFaucetURL
	case LocalnetEndpointURL:
		return LocalnetFaucetURL
	default:
		panic("unspecified FaucetURL")
	}
}