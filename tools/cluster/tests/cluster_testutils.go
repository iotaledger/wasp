package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
)

const incCounterSCName = "inccounter1"

var incCounterSCHname = iscp.Hn(incCounterSCName)

func deployIncCounterSC(t *testing.T, chain *cluster.Chain, counter *cluster.MessageCounter) *ledgerstate.Transaction {
	description := "testing contract deployment with inccounter" //nolint:goconst
	programHash := inccounter.Contract.ProgramHash
	check(err, t)

	tx, err := chain.DeployContract(incCounterSCName, programHash.String(), description, map[string]interface{}{
		inccounter.VarCounter: 42,
		root.ParamName:        incCounterSCName,
	})
	check(err, t)

	if counter != nil && !counter.WaitUntilExpectationsMet() {
		t.Fail()
	}

	checkCoreContracts(t, chain)

	for i := range chain.CommitteeNodes {
		blockIndex, err := chain.BlockIndex(i)
		require.NoError(t, err)
		require.EqualValues(t, 2, blockIndex)

		contractRegistry, err := chain.ContractRegistry(i)
		require.NoError(t, err)

		cr := contractRegistry[incCounterSCHname]

		require.EqualValues(t, programHash, cr.ProgramHash)
		require.EqualValues(t, description, cr.Description)
		require.EqualValues(t, 0, cr.OwnerFee)
		require.EqualValues(t, cr.Name, incCounterSCName)

		counterValue, err := chain.GetCounterValue(incCounterSCHname, i)
		require.NoError(t, err)
		require.EqualValues(t, 42, counterValue)
	}

	return tx
}

var addressIndex uint64 = 1

func createNewClient(t *testing.T, clu1 *cluster.Cluster, chain1 *cluster.Chain) *scclient.SCClient {
	keyPair, _ := getOrCreateAddress(t, clu1)

	client := chain1.SCClient(iscp.Hn(incCounterSCName), keyPair)

	return client
}

func getOrCreateAddress(t *testing.T, clu1 *cluster.Cluster) (*ed25519.KeyPair, *ledgerstate.ED25519Address) {
	const minTokenAmountBeforeRequestingNewFunds uint64 = 1000

	randomAddress := rand.NewSource(time.Now().UnixNano())

	keyPair := wallet.KeyPair(addressIndex)
	myAddress := ledgerstate.NewED25519Address(keyPair.PublicKey)

	funds, err := clu1.GoshimmerClient().BalanceIOTA(myAddress)

	require.NoError(t, err)

	if funds <= minTokenAmountBeforeRequestingNewFunds {
		// Requesting new token requires a new address

		addressIndex = rand.New(randomAddress).Uint64()
		t.Logf("Generating new address: %v", addressIndex)

		keyPair = wallet.KeyPair(addressIndex)
		myAddress = ledgerstate.NewED25519Address(keyPair.PublicKey)

		err = requestFunds(clu1, myAddress, "myAddress")
		t.Logf("Funds: %v, addressIndex: %v", funds, addressIndex)
		require.NoError(t, err)
	}

	return keyPair, myAddress
}
