package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*sctransaction.Transaction, error) {
	return c.ChainClient.PostRequest(c.ContractHname, coretypes.Hn(fname), params...)
}
