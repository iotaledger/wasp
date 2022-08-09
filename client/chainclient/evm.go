package chainclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/wasp/packages/isc"
)

func (c *Client) RequestIDByEVMTransactionHash(txHash common.Hash) (isc.RequestID, error) {
	return c.WaspClient.RequestIDByEVMTransactionHash(c.ChainID, txHash)
}
