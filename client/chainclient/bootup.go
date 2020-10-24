package chainclient

import (
	"github.com/iotaledger/wasp/packages/registry"
)

func (c *Client) GetBootupData() (*registry.BootupData, error) {
	return c.WaspClient.GetBootupData(c.ChainID)
}
