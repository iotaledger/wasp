package webapi_validation

import (
	"context"

	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

const MainnetChainID = "iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"

type ValidationContext struct {
	ctx     context.Context
	sClient *stardust_client.APIClient
	rClient *rebased_client.APIClient
}

func NewValidationContext(ctx context.Context, sClient *stardust_client.APIClient, rClient *rebased_client.APIClient) ValidationContext {
	return ValidationContext{
		ctx:     ctx,
		sClient: sClient,
		rClient: rClient,
	}
}
