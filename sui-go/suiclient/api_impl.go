package suiclient

import (
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

type Client struct {
	http      *suiconn.HTTPClient
	websocket *suiconn.WebsocketClient
}

func New(url string) *Client {
	return &Client{
		http: suiconn.NewHTTPClient(url),
	}
}

// test only. If localnet is used then iota network will be connect
func (i *Client) WithSignerAndFund(seed []byte, index int) (*Client, suisigner.Signer) {
	keySchemeFlag := suisigner.KeySchemeFlagEd25519
	// special case if localnet is used, then
	if i.http.URL() == suiconn.LocalnetEndpointURL {
		keySchemeFlag = suisigner.KeySchemeFlagIotaEd25519
	}
	// there are only 256 different signers can be generated
	signer := suisigner.NewSignerByIndex(seed, keySchemeFlag, index)
	var faucetURL string
	switch i.http.URL() {
	case suiconn.TestnetEndpointURL:
		faucetURL = suiconn.TestnetFaucetURL
	case suiconn.DevnetEndpointURL:
		faucetURL = suiconn.DevnetFaucetURL
	case suiconn.LocalnetEndpointURL:
		faucetURL = suiconn.LocalnetFaucetURL
	default:
		panic("not supported network")
	}
	err := RequestFundsFromFaucet(signer.Address(), faucetURL)
	if err != nil {
		panic(err)
	}
	return i, signer
}

func (i *Client) WithWebsocket(url string) {
	i.websocket = suiconn.NewWebsocketClient(url)
}

func NewWebsocket(url string) *Client {
	return &Client{
		websocket: suiconn.NewWebsocketClient(url),
	}
}

const (
	DefaultGasBudget uint64 = 10000000
	DefaultGasPrice  uint64 = 1000
)
