package chainclient

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// CallView sends a request to call a view function of a given contract, and returns the result of the call
func (c *Client) CallView(contractHname coretypes.Hname, fname string, arguments dict.Dict) (dict.Dict, error) {
	return c.WaspClient.CallView(coretypes.NewContractID(c.ChainID, contractHname), fname, arguments)
}
