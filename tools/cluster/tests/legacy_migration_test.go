package tests

import (
	"context"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/legacymigration"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster/templates"
)

func init() {
	os.Setenv("GO_TESTING", "1")
}

func TestMigrationAndBurn(t *testing.T) {
	EVMCluster := newCluster(t)

	// create an EVM chain
	EVMChain, err := EVMCluster.DeployDefaultChain()
	require.NoError(t, err)
	EVMEnv := newClusterTestEnv(t, &ChainEnv{
		t:     t,
		Clu:   EVMCluster,
		Chain: EVMChain,
	}, 0)

	// create an EVM AgentID with funds to deploy/interact with the "governance" contract
	ethPvtKey, ethUserAddr := EVMEnv.newEthereumAccountWithL2Funds()

	// deploy the "governance" contract
	govContractABI, err := abi.JSON(strings.NewReader(evmtest.LegacyMigrationGovernanceABI))
	require.NoError(t, err)
	_, _, govContractAddr := EVMEnv.DeployEVMContract(ethPvtKey, govContractABI, evmtest.LegacyMigrationGovernanceBytecode)
	govContractAgentID := isc.NewEthereumAddressAgentID(EVMChain.ChainID, govContractAddr)

	// deposit funds to the gov contract
	depositToAgentID(t, EVMEnv.Chain, 10*isc.Million, govContractAgentID, EVMChain.Cluster.OriginatorKeyPair)

	// deploy the migration chain
	migrationCluster := newCluster(t, waspClusterOpts{
		nNodes:  4,
		dirName: "wasp-cluster-legacymigration",
		modifyConfig: func(nodeIndex int, configParams templates.WaspConfigParams) templates.WaspConfigParams {
			// avoid port conflicts when running everything on localhost
			configParams.APIPort += 100
			configParams.MetricsPort += 100
			configParams.PeeringPort += 100
			configParams.ProfilingPort += 100
			return configParams
		},
	})
	migrationChain, err := migrationCluster.DeployChainWithDKG(migrationCluster.AllNodes(), migrationCluster.AllNodes(), 3, true)
	require.NoError(t, err)
	migrationEnv := newChainEnv(t, migrationCluster, migrationChain)

	// set gas fee to 0x000000
	migrationOriginatorClient := migrationChain.Client(migrationEnv.Clu.OriginatorKeyPair)
	req, err := migrationOriginatorClient.PostOffLedgerRequest(
		context.Background(),
		governance.Contract.Hname(),
		governance.FuncSetFeePolicy.Hname(),
		chainclient.PostRequestParams{
			Args: map[kv.Key][]byte{
				governance.ParamFeePolicyBytes: lo.Must(iotago.DecodeHex("0x0000000000000000000000000000000000")),
			},
		},
	)
	require.NoError(t, err)
	_, err = migrationChain.AllNodesMultiClient().WaitUntilRequestProcessedSuccessfully(migrationChain.ChainID, req.ID(), false, 30*time.Second)
	require.NoError(t, err)

	// fill the migration contract with funds
	someWallet, _, err := migrationCluster.NewKeyPairWithFunds()
	require.NoError(t, err)
	// deposit with transfer == allowance
	someWalletMigratorChainClient := migrationChain.Client(someWallet)
	transfer := isc.NewAssetsBaseTokens(100 * isc.Million)
	migrationContractAgentID := isc.NewContractAgentID(migrationChain.ChainID, legacymigration.Contract.Hname())
	tx, err := someWalletMigratorChainClient.Post1Request(
		accounts.Contract.Hname(),
		accounts.FuncTransferAllowanceTo.Hname(),
		chainclient.PostRequestParams{
			Transfer: transfer,
			Args: map[kv.Key][]byte{
				accounts.ParamAgentID: codec.EncodeAgentID(migrationContractAgentID),
			},
			Allowance: transfer,
		},
	)
	require.NoError(t, err)
	_, err = migrationChain.AllNodesMultiClient().WaitUntilAllRequestsProcessedSuccessfully(migrationChain.ChainID, tx, false, 30*time.Second)
	require.NoError(t, err)

	migBalance := migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
	require.Positive(t, migBalance)

	someWalletWithoutFunds := cryptolib.NewKeyPair()

	// assert valid migrations work
	{
		bundleHex, err2 := os.ReadFile("../../../packages/legacymigration/valid_bundle_example.hex")
		require.NoError(t, err2)
		bundleBytes, err2 := iotago.DecodeHex(string(bundleHex))
		require.NoError(t, err2)

		migrationReq, err2 := migrationChain.Client(someWalletWithoutFunds, 0).
			PostOffLedgerRequest(
				context.Background(),
				legacymigration.Contract.Hname(),
				legacymigration.FuncMigrate.Hname(),
				chainclient.PostRequestParams{
					Args: map[kv.Key][]byte{
						legacymigration.ParamBundle: bundleBytes,
					},
					Nonce: 10, // bad nonce, should be ignored
				},
			)
		require.NoError(t, err2)
		rec, err2 := migrationCluster.MultiClient().WaitUntilRequestProcessedSuccessfully(migrationChain.ChainID, migrationReq.ID(), true, 20*time.Second)
		require.NoError(t, err2)
		require.NotNil(t, rec)

		newMigBalance := migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
		require.Less(t, newMigBalance, migBalance)

		// we know that the pre-built bundle targets this address
		migrationTarget := iotago.MustParseEd25519AddressFromHexString("0x7ad1aee6262b8823aa74177692d917f2603c30587df6916f666eeb692f22b38d")
		require.EqualValues(t, migrationCluster.AddressBalances(migrationTarget).BaseTokens, migBalance-newMigBalance)
		migBalance = newMigBalance
	}

	// assert a second valid migration works
	{
		bundleHex, err2 := os.ReadFile("../../../packages/legacymigration/valid_bundle_example2.hex")
		require.NoError(t, err2)
		bundleBytes, err2 := iotago.DecodeHex(string(bundleHex))
		require.NoError(t, err2)

		migrationReq, err2 := migrationChain.Client(someWalletWithoutFunds, 0).
			PostOffLedgerRequest(
				context.Background(),
				legacymigration.Contract.Hname(),
				legacymigration.FuncMigrate.Hname(),
				chainclient.PostRequestParams{
					Args: map[kv.Key][]byte{
						legacymigration.ParamBundle: bundleBytes,
					},
					Nonce: 10, // reuse nonce 10, it should still work
				},
			)
		require.NoError(t, err2)
		rec, err2 := migrationCluster.MultiClient().WaitUntilRequestProcessedSuccessfully(migrationChain.ChainID, migrationReq.ID(), true, 20*time.Second)
		require.NoError(t, err2)
		require.NotNil(t, rec)

		newMigBalance := migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
		require.Less(t, newMigBalance, migBalance)

		// we know that the pre-built bundle targets this address
		migrationTarget := iotago.MustParseEd25519AddressFromHexString("0xcb1e2db315cafdb365fdc0dd71876cf633c61942c78f3690af9bc12c95ddbf30")
		require.EqualValues(t, migBalance-newMigBalance, migrationCluster.AddressBalances(migrationTarget).BaseTokens)
	}

	// assert invalid migrations are not processed by the migration chain
	{
		_, err2 := migrationChain.Client(cryptolib.NewKeyPair(), 0).
			PostOffLedgerRequest(
				context.Background(),
				legacymigration.Contract.Hname(),
				legacymigration.FuncMigrate.Hname(),
				chainclient.PostRequestParams{
					Args: map[kv.Key][]byte{
						legacymigration.ParamBundle: []byte("foobar"),
					},
				},
			)
		require.Error(t, err2)
	}

	//  assert random on-ledger requests are not processed by the migration chain
	{
		kp, _, err2 := EVMCluster.NewKeyPairWithFunds()
		require.NoError(t, err2)
		tx, err2 := migrationChain.Client(kp, 0).
			Post1Request(
				isc.Hn("dummycontract"),
				isc.Hn("dummyfunc"),
			)
		require.NoError(t, err2)
		_, err2 = migrationCluster.MultiClient().WaitUntilAllRequestsProcessed(migrationChain.ChainID, tx, false, 5*time.Second)
		require.Error(t, err2)
	}

	// assert random off-ledger requests are not processed by the migration chain
	{
		_, err2 := migrationChain.Client(cryptolib.NewKeyPair()).
			PostOffLedgerRequest(
				context.Background(),
				isc.Hn("dummycontract"),
				isc.Hn("dummyfunc"),
			)
		require.Error(t, err2)
	}

	issueGovTx := func(functionName string, args ...interface{}) {
		txArgs := []interface{}{
			iscmagic.WrapL1Address(migrationChain.ChainID.AsAddress()),
		}
		txArgs = append(txArgs, args...)
		burnTxData, err2 := govContractABI.Pack(functionName, txArgs...)
		require.NoError(t, err2)
		burnTx, err2 := types.SignTx(
			types.NewTransaction(EVMEnv.NonceAt(ethUserAddr), govContractAddr, big.NewInt(0), math.MaxUint64, big.NewInt(10000000000000000), burnTxData),
			EVMEnv.Signer(),
			ethPvtKey,
		)
		require.NoError(t, err2)
		rec, err2 := EVMEnv.SendTransactionAndWait(burnTx)
		require.NoError(t, err2)
		require.EqualValues(t, 1, rec.Status)
	}

	currentMigBlock, err := migrationChain.BlockIndex()
	require.NoError(t, err)

	waitForMigrationBlock := func(block uint32) {
		for {
			time.Sleep(100 * time.Millisecond)
			nextMigBlockIndex, err2 := migrationChain.BlockIndex()
			require.NoError(t, err2)
			if nextMigBlockIndex == block {
				break
			}
			if nextMigBlockIndex > block {
				t.Fatalf("error, expected block %d, got %d", block, nextMigBlockIndex)
			}
		}
	}

	// issue the BURN
	burnAddr := &iotago.Ed25519Address{0}
	burnAddrWrapped := iscmagic.WrapL1Address(burnAddr)
	issueGovTx("withdraw", burnAddrWrapped)
	// nothing should happen (admin is not the gov contract yet), chain will ignore the request
	time.Sleep(5 * time.Second)
	require.EqualValues(t, currentMigBlock, lo.Must(migrationChain.BlockIndex()))

	// set the EVM chain gov contract as the migration admin
	setNextAdminReq, err := migrationChain.Client(migrationChain.OriginatorKeyPair).PostOffLedgerRequest(
		context.Background(),
		legacymigration.Contract.Hname(),
		legacymigration.FuncSetNextAdmin.Hname(),
		chainclient.PostRequestParams{
			Args: map[kv.Key][]byte{
				legacymigration.ParamNextAdminAgentID: codec.Encode(govContractAgentID),
			},
		},
	)
	require.NoError(t, err)
	_, err = migrationCluster.MultiClient().WaitUntilRequestProcessedSuccessfully(migrationChain.ChainID, setNextAdminReq.ID(), true, 20*time.Second)
	require.NoError(t, err)
	currentMigBlock++

	// issue the BURN again
	issueGovTx("withdraw", burnAddrWrapped)
	// now the request should fail with an error (chain accepts requests from the next admin, but it has to claim ownership to do important things)
	waitForMigrationBlock(currentMigBlock + 1)
	currentMigBlock++
	latestReceipts, err := migrationChain.GetRequestReceiptsForBlock(nil)
	require.NoError(t, err)
	require.Len(t, latestReceipts, 1)
	require.EqualValues(t, latestReceipts[0].Request.SenderAccount, govContractAgentID.String())
	require.NotNil(t, latestReceipts[0].ErrorMessage)

	// call claim ownership, and wait for the migration chain to process it
	issueGovTx("claimOwnership")
	waitForMigrationBlock(currentMigBlock + 1)
	currentMigBlock++
	latestReceipts, err = migrationChain.GetRequestReceiptsForBlock(nil)
	require.NoError(t, err)
	require.Len(t, latestReceipts, 1)
	require.EqualValues(t, latestReceipts[0].Request.SenderAccount, govContractAgentID.String())
	require.Nil(t, latestReceipts[0].ErrorMessage)

	// re-issue the BURN, and wait for the migration chain to process it
	issueGovTx("withdraw", burnAddrWrapped)
	waitForMigrationBlock(currentMigBlock + 1)

	// assert balance was burned
	migBalance = migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
	require.Zero(t, migBalance)
	outs, err := migrationCluster.L1Client().OutputMap(burnAddr)
	require.NoError(t, err)
	require.Len(t, outs, 1)
}
