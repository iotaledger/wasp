package chainclient

import (
	"github.com/iotaledger/wasp/packages/registry"
)

func (c *Client) GetChainRecord() (*registry.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
