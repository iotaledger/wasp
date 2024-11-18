package iscmoveclienttest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

func StartNewChain(
	t *testing.T,
	client clients.L1Client,
	signer cryptolib.Signer,
	iscPackage iotago.PackageID,
	txFeePerReq uint64,
) (*iscmove.RefWithObject[iscmove.Anchor], *iotago.ObjectRef) {
	getCoinsRes, err := client.GetCoins(context.Background(), iotaclient.GetCoinsRequest{Owner: signer.Address().AsIotaAddress()})
	require.NoError(t, err)
	gasObj := getCoinsRes.Data[2]
	anchor, err := client.L2().StartNewChain(
		context.Background(),
		signer,
		iscPackage,
		[]byte{1, 2, 3, 4},
		getCoinsRes.Data[1].Ref(),
		gasObj.CoinObjectID,
		txFeePerReq,
		[]*iotago.ObjectRef{getCoinsRes.Data[0].Ref()},
		iotaclient.DefaultGasPrice,
		iotaclient.DefaultGasBudget,
	)
	require.NoError(t, err)
	return anchor, gasObj.Ref()
}
