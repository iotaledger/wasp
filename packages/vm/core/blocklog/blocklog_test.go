package blocklog

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestSerdeRequestLogRecord(t *testing.T) {
	var txid ledgerstate.TransactionID
	rand.Read(txid[:])
	rid := coretypes.RequestID(ledgerstate.NewOutputID(txid, 0))
	rec := &RequestLogRecord{
		RequestID: rid,
		OffLedger: true,
		LogData:   []byte("some log data"),
	}
	forward := rec.Bytes()
	back, err := RequestLogRecordFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}
