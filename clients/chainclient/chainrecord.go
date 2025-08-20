package chainclient

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
)

// GetChainRecord fetches the chain's Record
func (c *Client) GetChainRecord(ctx context.Context) (*apiclient.ChainInfoResponse, error) {
	chainInfo, _, err := c.WaspClient.ChainsAPI.GetChainInfo(ctx).Execute()
	return chainInfo, err
}
