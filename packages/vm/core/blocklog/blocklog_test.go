package blocklog

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/stretchr/testify/require"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("0"), isc.Hn("0"), nil, nonce)
	signedReq := req.Sign(cryptolib.NewKeyPair())
	rec := &RequestReceipt{
		Request: signedReq,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}
