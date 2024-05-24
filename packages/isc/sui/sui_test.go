package sui

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/isc"
	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/sui/conn"
	"github.com/iotaledger/wasp/sui-go/sui_signer"
	"github.com/iotaledger/wasp/sui-go/utils"
)

func setupAndDeploy(t *testing.T) {
	suiClient, signer := sui.NewTestSuiClientWithSignerAndFund(conn.LocalnetEndpointUrl, sui_signer.TEST_MNEMONIC)
	client := isc.NewIscClient(suiClient)

	modules, err := utils.MoveBuild(utils.GetGitRoot() + "/../kinesis/dapps/isc/")
	require.NoError(t, err)

	fmt.Printf("%s", signer.Address.String())

	txnBytes, err := client.Publish(
		context.Background(),
		signer.Address,
		modules.Modules,
		modules.Dependencies,
		nil,
		models.NewSafeSuiBigInt(uint64(100000000)),
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID := txnResponse.GetPublishedPackageID()
	t.Log("packageID: ", packageID)

	startNewChainRes, err := client.StartNewChain(
		context.Background(),
		signer,
		packageID,
		sui.DefaultGasBudget,
		&models.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, startNewChainRes.Effects.Data.IsSuccess())
	t.Logf("StartNewChain response: %#v\n", startNewChainRes)
}

func TestMinimalClient(t *testing.T) {
	setupAndDeploy(t)
}
