package chainclient

import (
	"github.com/iotaledger/wasp/packages/registry"
)

// GetChainRecord fetches the chain's ChainRecord
func (c *Client) GetChainRecord() (*registry.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
