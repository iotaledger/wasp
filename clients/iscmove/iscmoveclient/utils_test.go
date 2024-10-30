package iscmoveclient_test

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove/iscmoveclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
)

func TestMain(m *testing.M) {
	flag.Parse()
	iotaNode := l1starter.Start(context.Background(), l1starter.DefaultConfig)
	defer iotaNode.Stop()
	m.Run()
}

func buildDeployMintTestcoin(
	t *testing.T,
	client *iscmoveclient.Client,
	signer cryptolib.Signer,
) (
	coinRef *iotago.ObjectRef,
	coinType *iotago.ResourceType,
) {
	tokenPackageID, treasuryCap := iotaclienttest.DeployCoinPackage(
		t,
		client.Client,
		cryptolib.SignerToIotaSigner(signer),
		contracts.Testcoin(),
	)
	mintAmount := uint64(1000000)
	coinRef = iotaclienttest.MintCoins(
		t,
		client.Client,
		cryptolib.SignerToIotaSigner(signer),
		tokenPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap.ObjectID,
		mintAmount,
	)
	coinType = lo.Must(iotago.NewResourceType(fmt.Sprintf(
		"%s::%s::%s",
		tokenPackageID.String(),
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
	)))
	return
}
