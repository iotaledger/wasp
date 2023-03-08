package origin_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func TestOrigin(t *testing.T) {
	store := origin.InitChain(state.NewStore(mapdb.NewMapDB()), nil, 0)
	l1commitment := origin.L1Commitment(nil, 0)
	block, err := store.LatestBlock()
	require.NoError(t, err)
	require.True(t, l1commitment.Equals(block.L1Commitment()))
}

func TestCreateOrigin(t *testing.T) {
	var u *utxodb.UtxoDB
	var originTx *iotago.Transaction
	var userKey *cryptolib.KeyPair
	var userAddr, stateAddr *iotago.Ed25519Address
	var err error
	var chainID isc.ChainID
	var originTxID iotago.TransactionID

	initTest := func() {
		u = utxodb.New()
		userKey = cryptolib.NewKeyPair()
		userAddr = userKey.GetPublicKey().AsEd25519Address()
		_, err2 := u.GetFundsFromFaucet(userAddr)
		require.NoError(t, err2)

		stateKey := cryptolib.NewKeyPair()
		stateAddr = stateKey.GetPublicKey().AsEd25519Address()

		require.EqualValues(t, utxodb.FundsFromFaucetAmount, u.GetAddressBalanceBaseTokens(userAddr))
		require.EqualValues(t, 0, u.GetAddressBalanceBaseTokens(stateAddr))
	}
	createOrigin := func() {
		allOutputs, ids := u.GetUnspentOutputs(userAddr)

		originTx, _, chainID, err = origin.NewChainOriginTransaction(
			userKey,
			stateAddr,
			stateAddr,
			1000,
			nil,
			allOutputs,
			ids,
		)
		require.NoError(t, err)

		err = u.AddToLedger(originTx)
		require.NoError(t, err)

		originTxID, err = originTx.ID()
		require.NoError(t, err)

		txBack, ok := u.GetTransaction(originTxID)
		require.True(t, ok)
		txidBack, err2 := txBack.ID()
		require.NoError(t, err2)
		require.EqualValues(t, originTxID, txidBack)

		t.Logf("New chain ID: %s", chainID.String())
	}

	t.Run("create origin", func(t *testing.T) {
		initTest()
		createOrigin()

		anchor, _, err := transaction.GetAnchorFromTransaction(originTx)
		require.NoError(t, err)
		require.True(t, anchor.IsOrigin)
		require.EqualValues(t, chainID, anchor.ChainID)
		require.EqualValues(t, 0, anchor.StateIndex)
		require.True(t, stateAddr.Equal(anchor.StateController))
		require.True(t, stateAddr.Equal(anchor.GovernanceController))
		require.True(t,
			bytes.Equal(
				origin.L1Commitment(
					dict.Dict{origin.ParamChainOwner: isc.NewAgentID(anchor.GovernanceController).Bytes()},
					accounts.MinimumBaseTokensOnCommonAccount,
				).Bytes(),
				anchor.StateData),
		)

		// only one output is expected in the ledger under the address of chainID
		outs, ids := u.GetUnspentOutputs(chainID.AsAddress())
		require.EqualValues(t, 1, len(outs))
		require.EqualValues(t, 1, len(ids))

		out := u.GetOutput(anchor.OutputID)
		require.NotNil(t, out)
	})
	t.Run("create init chain originTx", func(t *testing.T) {
		initTest()
		createOrigin()

		chainBaseTokens := originTx.Essence.Outputs[0].Deposit()

		t.Logf("chainBaseTokens: %d", chainBaseTokens)

		require.EqualValues(t, utxodb.FundsFromFaucetAmount-chainBaseTokens, int(u.GetAddressBalanceBaseTokens(userAddr)))
		require.EqualValues(t, 0, u.GetAddressBalanceBaseTokens(stateAddr))
		allOutputs, ids := u.GetUnspentOutputs(chainID.AsAddress())
		require.EqualValues(t, 1, len(allOutputs))
		require.EqualValues(t, 1, len(ids))
	})
}
