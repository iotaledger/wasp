package scclient

import (
	"time"

	"github.com/iotaledger/wasp/packages/kv/dict"
)

func (c *SCClient) CallView(functionName string, args dict.Dict, optimisticReadTimeout ...time.Duration) (dict.Dict, error) {
	return c.ChainClient.CallView(c.ContractHname, functionName, args, optimisticReadTimeout...)
}
