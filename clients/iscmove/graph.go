package iscmove

import (
	"context"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/iotaledger/wasp/clients/iscmove/sui_graph"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type SuiGraph struct {
	graphqlURL string
}

func NewGraph(graphqlURL string) *SuiGraph {
	return &SuiGraph{
		graphqlURL,
	}
}

func (g *SuiGraph) GetAssetsBag(ctx context.Context, assetBagID sui.ObjectID) (
	*sui_graph.GetAssetsBagResponse,
	error,
) {
	httpClient := http.Client{}
	graphqlClient := graphql.NewClient(g.graphqlURL, &httpClient)
	return sui_graph.GetAssetsBag(ctx, graphqlClient, assetBagID.String())
}
