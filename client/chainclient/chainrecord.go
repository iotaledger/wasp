package chainclient

import "github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"

// GetChainRecord fetches the chain's Record
func (c *Client) GetChainRecord() (*chainrecord.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
