package chainclient

import "github.com/iotaledger/wasp/packages/registry_pkg/chain_record"

// GetChainRecord fetches the chain's Record
func (c *Client) GetChainRecord() (*chain_record.ChainRecord, error) {
	return c.WaspClient.GetChainRecord(c.ChainID)
}
