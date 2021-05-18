package chainclient

import (
	"github.com/iotaledger/wasp/packages/registry_pkg"
)

// GetChainRecord fetches the chain's Record
func (c *Client) GetChainRecord() (*registry_pkg.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
