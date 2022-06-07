package chainclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *Client) EVMRequestIDByTransactionHash(txHash common.Hash) (iscp.RequestID, error) {
	return c.WaspClient.EVMRequestIDByTransactionHash(c.ChainID, txHash)
}
