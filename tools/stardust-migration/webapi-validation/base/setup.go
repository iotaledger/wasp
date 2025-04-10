package base

import (
	"context"
	"strconv"

	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
)

const MainnetChainID = "iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"

type ValidationContext struct {
	Ctx     context.Context
	SClient *stardust_client.APIClient
	RClient *rebased_client.APIClient
}

func NewValidationContext(ctx context.Context, sClient *stardust_client.APIClient, rClient *rebased_client.APIClient) ValidationContext {
	return ValidationContext{
		Ctx:     ctx,
		SClient: sClient,
		RClient: rClient,
	}
}

func Uint32ToString(uint32 uint32) string {
	return strconv.FormatUint(uint64(uint32), 10)
}
