package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*sctransaction_old.TransactionEssence, error) {
	return c.ChainClient.PostRequest(c.ContractHname, coretypes.Hn(fname), params...)
}
