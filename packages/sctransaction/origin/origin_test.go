package origin

import (
	"testing"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/utxodb"
	"github.com/iotaledger/wasp/packages/sctransaction"
	_ "github.com/iotaledger/wasp/packages/sctransaction/properties"
	"github.com/stretchr/testify/assert"
)

const (
	scAddrStr = "behcN8An9CQDU3rcHSUucamZreVWfJepmiuBoSpomfi1"
	dscr      = "test test sc"
)

func TestReadWrite(t *testing.T) {
	u := utxodb.New()

	scAddr, err := address.FromBase58(scAddrStr)
	assert.NoError(t, err)

	ownerSigScheme := signaturescheme.RandBLS()

	_, err = u.RequestFunds(ownerSigScheme.Address())
	assert.NoError(t, err)

	tx, err := NewOriginTransaction(NewOriginTransactionParams{
		OriginAddress:             scAddr,
		OriginatorSignatureScheme: ownerSigScheme,
		AllInputs:                 u.GetAddressOutputs(ownerSigScheme.Address()),
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
