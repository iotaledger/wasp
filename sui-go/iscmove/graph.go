package iscmove

import (
	"context"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/iotaledger/wasp/sui-go/iscmove/sui_graph"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

type SuiGraph struct {
	graphqlUrl string
}

func NewGraph(api *sui.ImplSuiAPI, graphqlUrl string) *SuiGraph {
	return &SuiGraph{
		graphqlUrl,
	}
}

func (g *SuiGraph) GetAssetBag(ctx context.Context, assetBagId sui_types.ObjectID) (
	*sui_graph.GetAssetsBagResponse,
	error,
) {
	httpClient := http.Client{}
	graphqlClient := graphql.NewClient(g.graphqlUrl, &httpClient)
	return sui_graph.GetAssetsBag(ctx, graphqlClient, assetBagId.String())
}
