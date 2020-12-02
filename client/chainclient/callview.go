package chainclient

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (c *Client) CallView(contractHname coretypes.Hname, fname string, arguments dict.Dict) (dict.Dict, error) {
	return c.WaspClient.CallView(coretypes.NewContractID(c.ChainID, contractHname), fname, arguments)
}
