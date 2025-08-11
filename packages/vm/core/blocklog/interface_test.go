package blocklog_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestOutputRequestReceiptCodec(t *testing.T) {
	v := &blocklog.RequestReceipt{
		Request: isc.NewOffLedgerRequest(isc.NewMessage(isc.Hn("0"),
			isc.Hn("0")), 123, 456).Sign(cryptolib.NewKeyPair()),
		Error: &isc.UnresolvedVMError{
			ErrorCode: blocklog.ErrBlockNotFound.Code(),
			Params:    []isc.VMErrorParam{int16(1), uint64(2), "string"},
		},
		GasBudget:     123,
		GasBurned:     456,
		GasFeeCharged: 789,
		BlockIndex:    123,
		RequestIndex:  456,
		GasBurnLog: &gas.BurnLog{
			Records: []gas.BurnRecord{
				{
					Code:      1,
					GasBurned: 2,
				},
			},
		},
	}

	vEnc := blocklog.OutputRequestReceipt{}.Encode(v)
	vDec, err := blocklog.OutputRequestReceipt{}.Decode(vEnc)
	require.NoError(t, err)
	require.Equal(t, v, vDec)
}

func TestOutputRequestReceiptsCodec(t *testing.T) {
	v := &blocklog.RequestReceiptsResponse{
		BlockIndex: 123,
		Receipts: []*blocklog.RequestReceipt{
			{
				Request: isc.NewOffLedgerRequest(isc.NewMessage(isc.Hn("0"),
					isc.Hn("0")), 123, 456).Sign(cryptolib.NewKeyPair()),
				Error: &isc.UnresolvedVMError{
					ErrorCode: blocklog.ErrBlockNotFound.Code(),
					Params:    []isc.VMErrorParam{int16(1), uint64(2), "string"},
				},
				GasBudget:     123,
				GasBurned:     456,
				GasFeeCharged: 789,
				BlockIndex:    123,
				RequestIndex:  0,
				GasBurnLog: &gas.BurnLog{
					Records: []gas.BurnRecord{
						{
							Code:      1,
							GasBurned: 2,
						},
					},
				},
			},
			{
				Request: isc.NewOffLedgerRequest(isc.NewMessage(isc.Hn("0"),
					isc.Hn("0")), 123, 456).Sign(cryptolib.NewKeyPair()),
				Error: &isc.UnresolvedVMError{
					ErrorCode: blocklog.ErrBlockNotFound.Code(),
					Params:    []isc.VMErrorParam{int16(1), uint64(2), "string"},
				},
				GasBudget:     123,
				GasBurned:     456,
				GasFeeCharged: 789,
				BlockIndex:    123,
				RequestIndex:  1,
				GasBurnLog: &gas.BurnLog{
					Records: []gas.BurnRecord{
						{
							Code:      1,
							GasBurned: 2,
						},
					},
				},
			},
		},
	}

	vEnc := blocklog.OutputRequestReceipts{}.Encode(v)
	vDec, err := blocklog.OutputRequestReceipts{}.Decode(vEnc)
	require.NoError(t, err)
	require.Equal(t, v, vDec)
}
