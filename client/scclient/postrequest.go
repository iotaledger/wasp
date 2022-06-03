package scclient

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*iotago.Transaction, error) {
	return c.ChainClient.Post1Request(c.ContractHname, iscp.Hn(fname), params...)
}

func (c *SCClient) PostOffLedgerRequest(fname string, params ...chainclient.PostRequestParams) (iscp.OffLedgerRequest, error) {
	return c.ChainClient.PostOffLedgerRequest(c.ContractHname, iscp.Hn(fname), params...)
}
