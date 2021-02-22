package statetxbuilder

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/hashing"
	_ "github.com/iotaledger/wasp/packages/sctransaction/properties"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasic(t *testing.T) {
	chSig := signaturescheme.ED25519(ed25519.GenerateKeyPair())
	chAddr := chSig.Address()
	col1, _, err := balance.ColorFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)
	txid1, _, err := transaction.IDFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)
	txid2, _, err := transaction.IDFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)

	inps := map[transaction.ID][]*balance.Balance{
		txid1: {
			balance.New(col1, 1),
			balance.New(balance.ColorIOTA, 3),
		},
		txid2: {
			balance.New(balance.ColorIOTA, 5),
		},
	}
	b, err := New(chAddr, col1, inps)
	require.NoError(t, err)

	b.MustValidate()

	tx, err := b.Build()
	require.NoError(t, err)
	tx.Sign(chSig)

	prop, err := tx.Properties()
	require.NoError(t, err)
	t.Logf("properties: %s\ntx: %s\nvalue: %s", prop.String(), tx.String(), tx.Transaction.String())
}

func TestClone(t *testing.T) {
	chAddr := signaturescheme.ED25519(ed25519.GenerateKeyPair()).Address()
	col1, _, err := balance.ColorFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)
	txid1, _, err := transaction.IDFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)
	txid2, _, err := transaction.IDFromBytes(hashing.RandomHash(nil).Bytes())
	require.NoError(t, err)

	inps := map[transaction.ID][]*balance.Balance{
		txid1: {
			balance.New(col1, 1),
			balance.New(balance.ColorIOTA, 3),
		},
		txid2: {
			balance.New(balance.ColorIOTA, 5),
		},
	}
	b, err := New(chAddr, col1, inps)
	require.NoError(t, err)

	b.MustValidate()

	b1 := b.Clone()

	b1.MustValidate()

	tx, err := b.Build()
	require.NoError(t, err)

	tx1, err := b1.Build()
	require.NoError(t, err)

	require.EqualValues(t, tx.ID(), tx1.ID())
}
