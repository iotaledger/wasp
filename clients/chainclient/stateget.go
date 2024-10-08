package chainclient

import (
	"context"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

// ContractStateGet fetches the raw value associated with the given key in the chain state
func (c *Client) ContractStateGet(ctx context.Context, contract isc.Hname, key string) ([]byte, error) {
	return c.StateGet(ctx, string(contract.Bytes())+key)
}

// StateGet fetches the raw value associated with the given key in the chain state
func (c *Client) StateGet(ctx context.Context, key string) ([]byte, error) {
	stateResponse, _, err := c.WaspClient.ChainsApi.GetStateValue(ctx, c.ChainID.String(), iotago.EncodeHex([]byte(key))).Execute()
	if err != nil {
		return nil, err
	}

	return cryptolib.DecodeHex(stateResponse.State)
}
