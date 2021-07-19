package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
)

// SCClient allows to send webapi requests targeted to a specific contract
type SCClient struct {
	ChainClient   *chainclient.Client
	ContractHname iscp.Hname
}

// New creates a new SCClient
func New(
	chainClient *chainclient.Client,
	contractHname iscp.Hname,
) *SCClient {
	return &SCClient{
		ChainClient:   chainClient,
		ContractHname: contractHname,
	}
}
