package suiclient

import (
	"context"

	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

type Client struct {
	http *suiconn.HTTPClient
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
	faucetURL := suiconn.FaucetURL(i.http.URL())
	err := RequestFundsFromFaucet(context.Background(), signer.Address(), faucetURL)
	if err != nil {
		panic(err)
	}
	return i, signer
}

type WebsocketClient struct {
	*Client // TODO: should use websocket connection for all methods
	ws      *suiconn.WebsocketClient
}

func NewWebsocket(ctx context.Context, apiURL, wsURL string) *WebsocketClient {
	return &WebsocketClient{
		Client: New(apiURL),
		ws:     suiconn.NewWebsocketClient(ctx, wsURL),
	}
}

const (
	DefaultGasBudget uint64 = 10000000
	DefaultGasPrice  uint64 = 1000
)
