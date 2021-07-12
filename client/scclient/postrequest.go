package scclient

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*ledgerstate.Transaction, error) {
	return c.ChainClient.Post1Request(c.ContractHname, coretypes.Hn(fname), params...)
}

func (c *SCClient) PostOffLedgerRequest(fname string, params ...chainclient.PostRequestParams) (*request.RequestOffLedger, error) {
	return c.ChainClient.PostOffLedgerRequest(c.ContractHname, coretypes.Hn(fname), params...)
}
