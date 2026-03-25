package clients_test

import (
	"context"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/v2/packages/testutil/testpeers"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestExecuteTransactionBlockDeduplication(t *testing.T) {
	client := l1starter.Instance().L1Client()
	ctx := context.Background()

	// Set up a DSS signer — each call to SignTransactionBlock produces a different
	// (but valid) signature for the same data, because DSS uses randomness internally.
	log := testlogger.NewLogger(t)
	defer log.Shutdown()
	n := 4
	f := 1
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	nodeIDs := gpa.MakeTestNodeIDs(n)
	committeeAddr, dkRegs := testpeers.SetupDistributedKeyGenerationTrivial(t, n, f, peerIdentities, nil)
	dssSigner := testpeers.NewTestDistributedSignatureSigner(committeeAddr, dkRegs, nodeIDs, peerIdentities, log)
	signer := cryptolib.SignerToIotaSigner(dssSigner)

	// Fund the DSS signer address
	addr := signer.Address()
	err := client.RequestFundsFromFaucet(ctx, addr)
	require.NoError(t, err)

	// Get coins for gas payment
	coins, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: addr,
		Limit: 1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coins.Address.Coins.Nodes)

	coinRef, err := iotagraphql.Coins(coins.Address.Coins.Nodes)[0].ObjectRef()
	require.NoError(t, err)

	// Build a simple PayIota transaction (send 1000 NANOS back to self)
	ptb := iotago.NewProgrammableTransactionBuilder()
	err = ptb.PayIota([]*iotago.Address{&addr}, []uint64{1000})
	require.NoError(t, err)

	tx := iotago.NewProgrammable(&addr, ptb.Finish(), []*iotago.ObjectRef{coinRef}, iotagraphql.DefaultGasBudget, iotagraphql.DefaultGasPrice)
	txBytes, err := bcs.Marshal(&tx)
	require.NoError(t, err)

	// Sign and execute — first call succeeds
	sig1, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)

	resp1, err := client.ExecuteTransactionBlock(ctx, txBytes, []*iotasigner.Signature{sig1})
	require.NoError(t, err)
	require.True(t, resp1.IsSuccess())

	// Execute again with the same signature — should return the same result
	resp2, err := client.ExecuteTransactionBlock(ctx, txBytes, []*iotasigner.Signature{sig1})
	require.NoError(t, err)
	require.Equal(t, resp1, resp2)

	// Sign again — DSS produces a different valid signature for the same data
	sig2, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)
	require.NotEqual(t, sig1.Bytes(), sig2.Bytes())

	_, err = client.ExecuteTransactionBlock(ctx, txBytes, []*iotasigner.Signature{sig2})
	require.Error(t, err)
	require.Contains(t, err.Error(), "The transaction is already finalized but with different user signatures")
}
