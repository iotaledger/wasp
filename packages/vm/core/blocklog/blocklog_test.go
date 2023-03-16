package blocklog

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestSerdeRequestReceipt(t *testing.T) {
	nonce := uint64(time.Now().UnixNano())
	req := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("0"), isc.Hn("0"), nil, nonce, gas.LimitsDefault.MaxGasPerRequest)
	signedReq := req.Sign(cryptolib.NewKeyPair())
	rec := &RequestReceipt{
		Request: signedReq,
	}
	forward := rec.Bytes()
	back, err := RequestReceiptFromBytes(forward)
	require.NoError(t, err)
	require.EqualValues(t, forward, back.Bytes())
}

func TestBlockInfoOldSchema(t *testing.T) {
	oldSchemaBytes, _ := hex.DecodeString("01d3d73b00000000010001000000fe0afbc988e009f6290b666b7a1b52ed0f7c8a9354c95d88df50b1b4ea1f4fde5df1c4f838b5cf47a11af95629c25d03f34bf84621dc3f3589c9515b0e1e037d4e8aa8365a90cad78abfbd502db06c072b50377f9a412948a8d2c43052bb74b84858e6323b5ce38e0131fad7c5ad6d9409ee1e6ca22c875e580985cdddbd474130171046a3378611a3febfc1d317ba8aa1f8564c0000000000ea4c0f000000000010270000000000006400000000000000")
	bi, err := BlockInfoFromBytes(2, oldSchemaBytes)
	require.NoError(t, err)
	require.EqualValues(t, 0, bi.SchemaVersion)
	require.Nil(t, bi.AliasOutput)
	require.NotNil(t, bi.L1Commitment())
}
