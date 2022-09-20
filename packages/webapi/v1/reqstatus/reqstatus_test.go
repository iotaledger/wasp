package reqstatus

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"github.com/iotaledger/wasp/packages/webapi/v1/testutil"
)

const foo = "foo"

func mockedGetReceiptFromBlocklog(_ chain.Chain, id isc.RequestID) (*blocklog.RequestReceipt, error) {
	req := isc.NewOffLedgerRequest(
		isc.ChainID{123},
		isc.Hn("some contract"),
		isc.Hn("some entrypoint"),
		dict.Dict{foo: []byte("bar")},
		42,
	).Sign(cryptolib.NewKeyPair())

	return &blocklog.RequestReceipt{
		Request: req,
		Error: &isc.UnresolvedVMError{
			ErrorCode: isc.VMErrorCode{
				ContractID: isc.Hn("error contract"),
				ID:         3,
			},
			Params: []interface{}{},
			Hash:   0,
		},
		GasBudget:     123,
		GasBurned:     10,
		GasFeeCharged: 100,
		BlockIndex:    111,
		RequestIndex:  222,
	}, nil
}

func mockedResolveReceipt(c echo.Context, ch chain.Chain, rec *blocklog.RequestReceipt) error {
	iscReceipt := rec.ToISCReceipt(nil)
	receiptJSON, _ := json.Marshal(iscReceipt)
	return c.JSON(http.StatusOK,
		&model.RequestReceiptResponse{
			Receipt: string(receiptJSON),
		},
	)
}

func TestRequestReceipt(t *testing.T) {
	r := &reqstatusWebAPI{
		getChain: func(chainID isc.ChainID) chain.Chain {
			return &testutil.MockChain{}
		},
		getReceiptFromBlocklog: mockedGetReceiptFromBlocklog,
		resolveReceipt:         mockedResolveReceipt,
	}

	chainID := isc.RandomChainID()
	reqID := isc.NewRequestID(iotago.TransactionID{}, 0)

	var res model.RequestReceiptResponse
	testutil.CallWebAPIRequestHandler(
		t,
		r.handleRequestReceipt,
		http.MethodGet,
		routes.RequestReceipt(":chainID", ":reqID"),
		map[string]string{
			"chainID": chainID.String(),
			"reqID":   reqID.String(),
		},
		nil,
		&res,
		http.StatusOK,
	)

	require.NotEmpty(t, res)
}
