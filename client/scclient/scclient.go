package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
)

// SCClient allows to send webapi requests targeted to a specific contract
type SCClient struct {
	ChainClient   *chainclient.Client
	ContractHname coretypes.Hname
}

// New creates a new SCClient
func New(
	chainClient *chainclient.Client,
	contractHname coretypes.Hname,
) *SCClient {
	return &SCClient{
		ChainClient:   chainClient,
		ContractHname: contractHname,
	}
}
