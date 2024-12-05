package iotaconn

const (
	LocalnetEndpointURL = "http://localhost:9000"
	AlphanetEndpointURL = "https://api.iota-rebased-alphanet.iota.cafe"
	TestnetEndpointURL  = "https://api.testnet.iota.cafe"
	DevnetEndpointURL   = "https://api.devnet.iota.cafe"

	LocalnetWebsocketEndpointURL = "ws://localhost:9000"
	AlphanetWebsocketEndpointURL = "wss://api.iota-rebased-alphanet.iota.cafe"
	TestnetWebsocketEndpointURL  = "wss://api.testnet.iota.cafe"
	DevnetWebsocketEndpointURL   = "wss://api.devnet.iota.cafe"

	LocalnetFaucetURL = "http://localhost:9123/gas"
	AlphanetFaucetURL = "https://faucet.iota-rebased-alphanet.iota.cafe/gas"
	TestnetFaucetURL  = "https://faucet.testnet.iota.cafe/gas"
	DevnetFaucetURL   = "https://faucet.devnet.iota.cafe/gas"
)

type Host string

var (
	Localnet Host = "localhost"
	AlphaNet Host = "alphanet"
	DevNet   Host = "devnet"
	TestNet  Host = "testnet"
)

const (
	ChainIdentifierAlphanet = "c065f131"
	// localnet doesn't have a fixed ChainIdentifier
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
