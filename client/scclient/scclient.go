package scclient

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coret"
)

type SCClient struct {
	ChainClient   *chainclient.Client
	ContractHname coret.Hname
}

func New(
	chainClient *chainclient.Client,
	contractHname coret.Hname,
) *SCClient {
	return &SCClient{
		ChainClient:   chainClient,
		ContractHname: contractHname,
	}
}
