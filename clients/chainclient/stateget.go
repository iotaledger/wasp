// Package chainclient provides client functionality for interacting with the blockchain state.
package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

// ContractStateGet fetches the raw value associated with the given key in the chain state
func (c *Client) ContractStateGet(ctx context.Context, contract isc.Hname, key string) ([]byte, error) {
	return c.StateGet(ctx, string(contract.Bytes())+key)
}

// StateGet fetches the raw value associated with the given key in the chain state
func (c *Client) StateGet(ctx context.Context, key string) ([]byte, error) {
	stateResponse, _, err := c.WaspClient.ChainsAPI.GetStateValue(ctx, cryptolib.EncodeHex([]byte(key))).Execute()
	if err != nil {
		return nil, err
	}

	return cryptolib.DecodeHex(stateResponse.State)
}
