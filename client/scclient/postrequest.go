package scclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*ledgerstate.Transaction, error) {
	return c.ChainClient.PostRequest(c.ContractHname, coretypes.Hn(fname), params...)
}
