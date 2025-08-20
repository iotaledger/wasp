package iscmoveclient_test

import (
	"fmt"
	"testing"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func buildDeployMintTestcoin(
	t *testing.T,
	client *iscmoveclient.Client,
	signer cryptolib.Signer,
) (
	*iotago.ObjectRef,
	*iotago.ResourceType,
) {
	tokenPackageID, treasuryCap := iotaclienttest.DeployCoinPackage(
		t,
		client.Client,
		cryptolib.SignerToIotaSigner(signer),
		contracts.Testcoin(),
	)
	mintAmount := uint64(1000000)
	coinRef := iotaclienttest.MintCoins(
		t,
		client.Client,
		cryptolib.SignerToIotaSigner(signer),
		tokenPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		mintAmount,
	)
	coinType := lo.Must(iotago.NewResourceType(fmt.Sprintf(
		"%s::%s::%s",
		tokenPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	)))
	return coinRef, coinType
}
