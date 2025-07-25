package iotaclienttest

import (
	"context"
	"errors"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func GetValidatorAddress(ctx context.Context) (iotago.Address, error) {
	client := l1starter.Instance().L1Client()
	apy, err := client.GetValidatorsApy(ctx)
	if err != nil {
		return iotago.Address{}, err
	}
	validator1 := apy.Apys[0].Address
	address, err := iotago.AddressFromHex(validator1)
	if err != nil {
		return iotago.Address{}, err
	}

	return *address, nil
}

func GetValidatorAddressWithCoins(ctx context.Context) (iotago.Address, error) {
	client := l1starter.Instance().L1Client()
	apy, err := client.GetValidatorsApy(ctx)
	if err != nil {
		return iotago.Address{}, err
	}

	for _, apy := range apy.Apys {
		coins, err := client.GetCoins(
			ctx, iotaclient.GetCoinsRequest{
				Owner: iotago.MustAddressFromHex(apy.Address),
				Limit: 10,
			},
		)
		if err != nil {
			return iotago.Address{}, err
		}
		if len(coins.Data) > 0 {
			return *iotago.MustAddressFromHex(apy.Address), nil
		}
	}

	return iotago.Address{}, errors.New("validator with coins not found")
}
