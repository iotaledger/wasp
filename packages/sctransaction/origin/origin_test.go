package origin

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/stretchr/testify/assert"
)

const (
	scAddrStr = "behcN8An9CQDU3rcHSUucamZreVWfJepmiuBoSpomfi1"
	dscr      = "test test sc"
)

var nodeLocations = []string{"127.0.0.1:4000", "127.0.0.1:4001", "127.0.0.1:4002", "127.0.0.1:4003"}

func TestReadWrite(t *testing.T) {
	u := utxodb.New()

	scAddr, err := address.FromBase58(scAddrStr)
	assert.NoError(t, err)

	sigScheme := utxodb.NewSigScheme("C6hPhCS2E2dKUGS3qj4264itKXohwgL3Lm2fNxayAKr", 0)

	u.RequestFunds(sigScheme.Address())

	tx, err := NewOriginTransaction(NewOriginTransactionParams{
		Address:              scAddr,
		OwnerSignatureScheme: sigScheme,
		AllInputs:            u.GetAddressOutputs(sigScheme.Address()),
		InputColor:           balance.ColorIOTA,
		ProgramHash:          *hashing.HashStrings(dscr),
	})
	assert.NoError(t, err)
	t.Logf("created transaction txid = %s", tx.ID().String())

	data := tx.Bytes()
	vtx, _, err := valuetransaction.FromBytes(data)
	assert.NoError(t, err)

	txback, err := sctransaction.ParseValueTransaction(vtx)
	assert.NoError(t, err)

	assert.EqualValues(t, tx.ID(), txback.ID())
}
