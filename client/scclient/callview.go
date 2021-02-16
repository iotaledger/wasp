package scclient

import "github.com/iotaledger/wasp/packages/kv/dict"

func (c *SCClient) CallView(fname string, args dict.Dict) (dict.Dict, error) {
	return c.ChainClient.CallView(c.ContractHname, fname, args)
}
