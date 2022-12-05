package reqstatus

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

type mockedChain struct {
	// TODO mock chaincore is deprecated, what should be used in its place?
	//*testchain.MockedChainCore
}

// chain.ChainCore implementation

func (*mockedChain) ID() *isc.ChainID {
	panic("unimplemented")
}

func (*mockedChain) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("unimplemented")
}

func (*mockedChain) Processors() *processors.Cache {
	panic("unimplemented")
}

func (*mockedChain) GetStateReader() state.Store {
	panic("unimplemented")
}

func (*mockedChain) GetChainNodes() []peering.PeerStatusProvider {
	panic("unimplemented")
}

func (*mockedChain) GetCandidateNodes() []*governance.AccessNodeInfo {
	panic("unimplemented")
}

func (*mockedChain) Log() *logger.Logger {
	panic("unimplemented")
}

// chain.ChainRequests implementation

func (*mockedChain) ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey) {
	panic("unimplemented")
}

func (*mockedChain) AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID) <-chan *blocklog.RequestReceipt {
	panic("unimplemented")
}

// chain.Chain implementation

func (m *mockedChain) GetConsensusPipeMetrics() chain.ConsensusPipeMetrics {
	panic("unimplemented")
}

func (m *mockedChain) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	panic("unimplemented")
}

func (m *mockedChain) GetConsensusWorkflowStatus() chain.ConsensusWorkflowStatus {
	panic("unimplemented")
}

var _ chain.Chain = &mockedChain{}

/*
const foo = "foo"

// TODO: still needed?
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
*/

func TestRequestReceipt(t *testing.T) {
	r := &reqstatusWebAPI{func(chainID *isc.ChainID) chain.Chain {
		return &mockedChain{}
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
