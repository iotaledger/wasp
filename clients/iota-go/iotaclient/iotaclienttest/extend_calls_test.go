package iotaclienttest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestMintToken(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	tokenPackageID, treasuryCap := DeployCoinPackage(
		t,
		client.IotaClient(),
		signer,
		contracts.Testcoin(),
	)
	mintAmount := uint64(1000000)
	_ = MintCoins(
		t,
		client.IotaClient(),
		signer,
		tokenPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		mintAmount,
	)
	coinType := fmt.Sprintf(
		"%s::%s::%s",
		tokenPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	)

	// all the minted tokens were sent to the signer, so we should find a single object contains all the minted token
	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner:    signer.Address(),
			CoinType: &coinType,
			Limit:    10,
		},
	)
	require.NoError(t, err)
	require.Equal(t, mintAmount, coins.Data[0].Balance.Uint64())
}

func TestBatchGetObjectsOwnedByAddress(t *testing.T) {
	api := l1starter.Instance().L1Client()

	options := iotajsonrpc.IotaObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
	}
	coinType := fmt.Sprintf("0x2::coin::Coin<%v>", iotajsonrpc.IotaCoinType)
	address := iotago.MustAddressFromHex(testcommon.TestAddress)
	filterObject, err := api.BatchGetObjectsOwnedByAddress(context.Background(), address, &options, coinType)
	require.NoError(t, err)
	t.Log(filterObject)
}
