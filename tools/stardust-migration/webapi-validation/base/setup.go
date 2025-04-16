package base

import (
	"context"
	"strconv"

	stardust_client "github.com/nnikolash/wasp-types-exported/clients/apiclient"
	"github.com/samber/lo"

	rebased_client "github.com/iotaledger/wasp/clients/apiclient"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
)

const MainnetChainID = "iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5"

var ChainID old_isc.ChainID

type ValidationContext struct {
	Ctx     context.Context
	SClient *stardust_client.APIClient
	RClient *rebased_client.APIClient
}

func NewValidationContext(ctx context.Context, sClient *stardust_client.APIClient, rClient *rebased_client.APIClient) ValidationContext {
	ChainID = lo.Must(old_isc.ChainIDFromString(MainnetChainID))

	return ValidationContext{
		Ctx:     ctx,
		SClient: sClient,
		RClient: rClient,
	}
}

func Uint32ToString(uint32 uint32) string {
	return strconv.FormatUint(uint64(uint32), 10)
}
