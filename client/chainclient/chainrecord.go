package chainclient

import "github.com/iotaledger/wasp/packages/coretypes"

// GetChainRecord fetches the chain's Record
func (c *Client) GetChainRecord() (*coretypes.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
