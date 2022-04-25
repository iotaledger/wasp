package blocklog

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/stretchr/testify/require"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := iscp.NewOffLedgerRequest(iscp.RandomChainID(), iscp.Hn("0"), iscp.Hn("0"), nil, nonce)

	rec := &RequestReceipt{
		Request: req,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}
