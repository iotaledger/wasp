package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
)

type SCClient struct {
	ChainClient   *chainclient.Client
	ContractHname coretypes.Hname
}

func New(
	chainClient *chainclient.Client,
	contractHname coretypes.Hname,
) *SCClient {
	return &SCClient{
		ChainClient:   chainClient,
		ContractHname: contractHname,
	}
}
