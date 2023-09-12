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
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/legacymigration"
	"github.com/iotaledger/wasp/packages/vm/core/evm/iscmagic"
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
	ethPvtKey, _ := EVMEnv.newEthereumAccountWithL2Funds()

	// deploy the "governance" contract
	govContractABI, err := abi.JSON(strings.NewReader(evmtest.LegacyMigrationGovernanceABI))
	require.NoError(t, err)
	_, _, govContractAddr := EVMEnv.DeployEVMContract(ethPvtKey, govContractABI, evmtest.LegacyMigrationGovernanceBytecode)
	govContractAgentID := isc.NewEthereumAddressAgentID(EVMChain.ChainID, govContractAddr)

	// deposit funds to the gov contract
	depositToAgentID(t, EVMEnv.Chain, 1*isc.Million, govContractAgentID, EVMChain.Cluster.OriginatorKeyPair)

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
	migrationChain, err := migrationCluster.DeployChainWithDKG(migrationCluster.AllNodes(), migrationCluster.AllNodes(), 3, govContractAgentID)
	require.NoError(t, err)
	migrationEnv := newChainEnv(t, migrationCluster, migrationChain)

	// fill the migration contract with funds
	someWallet, _, err := migrationCluster.NewKeyPairWithFunds()
	require.NoError(t, err)
	migrationContractAgentID := isc.NewContractAgentID(migrationChain.ChainID, legacymigration.Contract.Hname())
	depositToAgentID(t, migrationChain, 100*isc.Million, migrationContractAgentID, someWallet)
	migBalance := migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
	require.Positive(t, migBalance)

	// assert valid migrations work
	{
		bundleHex, err2 := os.ReadFile("../../../packages/legacymigration/valid_bundle_example.hex")
		require.NoError(t, err2)
		bundleBytes, err2 := iotago.DecodeHex(string(bundleHex))
		require.NoError(t, err2)

		migrationReq, err2 := migrationChain.Client(cryptolib.NewKeyPair(), 0).
			PostOffLedgerRequest(
				context.Background(),
				legacymigration.Contract.Hname(),
				legacymigration.FuncMigrate.Hname(),
				chainclient.PostRequestParams{
					Args: map[kv.Key][]byte{
						legacymigration.ParamBundle: bundleBytes,
					},
				},
			)
		require.NoError(t, err2)
		_, err2 = migrationCluster.MultiClient().WaitUntilRequestProcessedSuccessfully(migrationChain.ChainID, migrationReq.ID(), true, 20*time.Second)
		require.NoError(t, err2)

		newMigBalance := migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
		require.Less(t, newMigBalance, migBalance)

		// we know that the pre-built bundle targets this address
		migrationTarget := iotago.MustParseEd25519AddressFromHexString("0x7ad1aee6262b8823aa74177692d917f2603c30587df6916f666eeb692f22b38d")
		require.EqualValues(t, migrationCluster.AddressBalances(migrationTarget).BaseTokens, migBalance-newMigBalance)
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

	// assert radom off-ledger requests are not processed by the migration chain
	{
		_, err2 := migrationChain.Client(cryptolib.NewKeyPair(), 0).
			PostOffLedgerRequest(
				context.Background(),
				isc.Hn("dummycontract"),
				isc.Hn("dummyfunc"),
			)
		require.Error(t, err2)
	}

	migBlockIndex, err := migrationChain.BlockIndex()
	require.NoError(t, err)

	// issue the BURN
	burnTxData, err := govContractABI.Pack("burn", iscmagic.WrapL1Address(migrationChain.ChainID.AsAddress()))
	require.NoError(t, err)
	burnTx, err := types.SignTx(
		types.NewTransaction(1, govContractAddr, big.NewInt(0), math.MaxUint64, big.NewInt(10000000000000000), burnTxData),
		EVMEnv.Signer(),
		ethPvtKey,
	)
	require.NoError(t, err)
	rec, err := EVMEnv.SendTransactionAndWait(burnTx)
	require.NoError(t, err)
	require.EqualValues(t, 1, rec.Status)

	// wait for migration chain to process the cross-chain gov-burn request
	for {
		time.Sleep(100 * time.Millisecond)
		nextMigBlockIndex, err := migrationChain.BlockIndex()
		require.NoError(t, err)
		if nextMigBlockIndex == migBlockIndex+1 {
			break
		}
	}

	migBalance = migrationEnv.getBalanceOnChain(migrationContractAgentID, isc.BaseTokenID)
	require.Zero(t, migBalance)
}
