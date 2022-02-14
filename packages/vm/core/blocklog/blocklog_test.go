package blocklog

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/stretchr/testify/require"
)

/*func TestSimpleErrorSerialization(t *testing.T) {
	mu := marshalutil.New()

	// Initial error
	blockError := FailedToLoadError.Create("placeBet", "destroy", "setAdmin")

	// Serialize error to bytes
	err := blockError.Serialize(mu)
	require.NoError(t, err)

	// Recreate error from bytes
	newError, err := errors.ErrorFromBytes(mu, nil)
	require.NoError(t, err)

	// Validate that properties are the same
	require.EqualValues(t, blockError.Hash(), newError.Hash())
	require.EqualValues(t, blockError.params, newError.params)
	require.EqualValues(t, blockError.messageFormat, newError.messageFormat)

	// Validate that error returns a proper error type
	require.Error(t, newError)

	t.Log(newError.Error())
}*/

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := iscp.NewOffLedgerRequest(iscp.RandomChainID(), iscp.Hn("0"), iscp.Hn("0"), nil, nonce)
	//testError, err := GeneralErrorCollection.Create(1, 1)

	rec := &RequestReceipt{
		Request: req,
		//Error:   testError,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}
