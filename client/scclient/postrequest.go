package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*sctransaction.Transaction, error) {
	return c.ChainClient.PostRequest(c.ContractHname, coret.Hn(fname), params...)
}
