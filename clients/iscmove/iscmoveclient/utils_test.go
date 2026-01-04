package iscmoveclient_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql/iotaclienttest"
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
		client,
		cryptolib.SignerToIotaSigner(signer),
		contracts.Testcoin(),
	)
	mintAmount := uint64(1000000)
	time.Sleep(1 * time.Second) // FIXME tmp for graphql
	coinRef := iotaclienttest.MintCoins(
		t,
		client,
		cryptolib.SignerToIotaSigner(signer),
		tokenPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap,
		mintAmount,
	)
	time.Sleep(1 * time.Second) // FIXME tmp for graphql
	coinType := lo.Must(iotago.NewResourceType(fmt.Sprintf(
		"%s::%s::%s",
		tokenPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	)))
	return coinRef, coinType
}
