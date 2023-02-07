package chainclient

import (
	"context"

	iotago "github.com/iotaledger/iota.go/v3"
)

// StateGet fetches the raw value associated with the given key in the chain state
func (c *Client) StateGet(ctx context.Context, key string) ([]byte, error) {
	stateResponse, _, err := c.WaspClient.ChainsApi.GetStateValue(ctx, c.ChainID.String(), iotago.EncodeHex([]byte(key))).Execute()

	if err != nil {
		return nil, err
	}

	hexBytes, err := iotago.DecodeHex(stateResponse.State)

	return hexBytes, err
}
