package scclient

import (
	"context"

	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (c *SCClient) CallView(context context.Context, functionName string, args dict.Dict) (dict.Dict, error) {
	return c.ChainClient.CallView(context, c.ContractHname, functionName, args)
}
