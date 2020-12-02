package chainclient

import (
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (c *Client) CallView(contractHname coret.Hname, fname string, arguments dict.Dict) (dict.Dict, error) {
	return c.WaspClient.CallView(coret.NewContractID(c.ChainID, contractHname), fname, arguments)
}
