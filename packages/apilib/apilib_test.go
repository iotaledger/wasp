package apilib

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/magiconair/properties/assert"
	"testing"
)

const (
	scAddrStr = "behcN8An9CQDU3rcHSUucamZreVWfJepmiuBoSpomfi1"
	dscr      = "test test sc"
)

var nodeLocations = []string{"127.0.0.1:4000", "127.0.0.1:4001", "127.0.0.1:4002", "127.0.0.1:4003"}

func TestReadWrite(t *testing.T) {
	scAddr, err := address.FromBase58(scAddrStr)
	assert.Equal(t, err, nil)

	tx, _ := CreateOriginData(&NewOriginParams{
		Address:      scAddr,
		OwnerAddress: utxodb.GetAddress(1),
		ProgramHash:  *hashing.HashStrings(dscr),
	}, dscr, nodeLocations)
	t.Logf("created transaction txid = %s", tx.ID().String())

	data := tx.Bytes()
	vtx, _, err := valuetransaction.FromBytes(data)
	assert.Equal(t, err, nil)

	txback, err := sctransaction.ParseValueTransaction(vtx)
	assert.Equal(t, err, nil)

	assert.Equal(t, tx.ID(), txback.ID())
}
