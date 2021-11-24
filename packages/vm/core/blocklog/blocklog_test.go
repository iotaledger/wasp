package blocklog

import (
	"math/rand"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/stretchr/testify/require"
)

func TestSerdeRequestLogRecord(t *testing.T) {
	var txid ledgerstate.TransactionID
	rand.Read(txid[:])
	req := iscp.NewOffLedgerRequest(iscp.RandomChainID(), iscp.Hn("0"), iscp.Hn("0"))

	rec := &RequestReceipt{
		Request: req,
		Error:   "some log data",
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}
