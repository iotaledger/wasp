package reqstatus

import (
	"net/http"
	"testing"

	"github.com/iotaledger/hive.go/events"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

type mockChain struct{}

var _ chain.ChainRequests = &mockChain{}

const foo = "foo"

func (m *mockChain) GetRequestReceipt(id iscp.RequestID) (*blocklog.RequestReceipt, error) {
	req := iscp.NewOffLedgerRequest(
		&iscp.ChainID{123}, iscp.Hn("some contract"), iscp.Hn("some entrypoint"), dict.Dict{foo: []byte("bar")}, 42,
	)
	return &blocklog.RequestReceipt{
		Request: req,
		Error: &iscp.UnresolvedVMError{
			ErrorCode: iscp.VMErrorCode{
				ContractID: iscp.Hn("error contract"),
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

func (m *mockChain) GetRequestProcessingStatus(id iscp.RequestID) chain.RequestProcessingStatus {
	return chain.RequestProcessingStatusCompleted
}

func (m *mockChain) TranslateError(e *iscp.UnresolvedVMError) (string, error) {
	return "translated", nil
}

func (m *mockChain) AttachToRequestProcessed(func(iscp.RequestID)) (attachID *events.Closure) {
	panic("not implemented")
}

func (m *mockChain) DetachFromRequestProcessed(attachID *events.Closure) {
	panic("not implemented")
}

func (m *mockChain) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	panic("not implemented")
}

func TestRequestReceipt(t *testing.T) {
	r := &reqstatusWebAPI{func(chainID *iscp.ChainID) chain.ChainRequests {
		return &mockChain{}
	}}

	chainID := iscp.RandomChainID()
	reqID := iscp.NewRequestID(iotago.TransactionID{}, 0)

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
