package scclient

import (
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
)

// SCClient allows to send webapi requests targeted to a specific contract
type SCClient struct {
	ChainClient   *chainclient.Client
	ContractHname isc.Hname
}

// New creates a new SCClient
func New(
	chainClient *chainclient.Client,
	contractHname isc.Hname,
) *SCClient {
	return &SCClient{
		ChainClient:   chainClient,
		ContractHname: contractHname,
	}
}
