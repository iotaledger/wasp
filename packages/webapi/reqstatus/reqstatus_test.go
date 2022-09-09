package reqstatus

import (
	"net/http"
	"testing"

	"github.com/iotaledger/hive.go/events"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
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

func (m *mockChain) GetRequestReceipt(id isc.RequestID) (*blocklog.RequestReceipt, error) {
	req := isc.NewOffLedgerRequest(
		&isc.ChainID{123},
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

func (m *mockChain) GetRequestProcessingStatus(id isc.RequestID) chain.RequestProcessingStatus {
	return chain.RequestProcessingStatusCompleted
}

func (m *mockChain) ResolveError(e *isc.UnresolvedVMError) (*isc.VMError, error) {
	return nil, nil
}

func (m *mockChain) AttachToRequestProcessed(func(isc.RequestID)) (attachID *events.Closure) {
	panic("not implemented")
}

func (m *mockChain) DetachFromRequestProcessed(attachID *events.Closure) {
	panic("not implemented")
}

func (m *mockChain) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	panic("not implemented")
}

func TestRequestReceipt(t *testing.T) {
	r := &reqstatusWebAPI{func(chainID *isc.ChainID) chain.ChainRequests {
		return &mockChain{}
	}}

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
