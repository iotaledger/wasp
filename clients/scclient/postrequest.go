package scclient

import (
	"context"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
)

func (c *SCClient) PostRequest(fname string, params ...chainclient.PostRequestParams) (*iotago.Transaction, error) {
	return c.ChainClient.Post1Request(c.ContractHname, isc.Hn(fname), params...)
}

func (c *SCClient) PostNRequests(fname string, n int, params ...chainclient.PostRequestParams) ([]*iotago.Transaction, error) {
	return c.ChainClient.PostNRequests(c.ContractHname, isc.Hn(fname), n, params...)
}

func (c *SCClient) PostOffLedgerRequest(fname string, params ...chainclient.PostRequestParams) (isc.OffLedgerRequest, error) {
	return c.ChainClient.PostOffLedgerRequest(context.Background(), c.ContractHname, isc.Hn(fname), params...)
}
