package reqstatus

import (
	"net/http"
	"testing"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

type mockChain struct{}

var _ chain.ChainRequests = &mockChain{}

func (m *mockChain) GetRequestProcessingStatus(id iscp.RequestID) chain.RequestProcessingStatus {
	return chain.RequestProcessingStatusCompleted
}

func (m *mockChain) AttachToRequestProcessed(func(iscp.RequestID)) (attachID *events.Closure) {
	panic("not implemented")
}

func (m *mockChain) DetachFromRequestProcessed(attachID *events.Closure) {
	panic("not implemented")
}

func TestRequestStatus(t *testing.T) {
	r := &reqstatusWebAPI{func(chainID *iscp.ChainID) chain.ChainRequests {
		return &mockChain{}
	}}

	chainID := iscp.RandomChainID()
	reqID := iscp.RequestID(ledgerstate.OutputID{})

	var res model.RequestStatusResponse
	testutil.CallWebAPIRequestHandler(
		t,
		r.handleRequestStatus,
		http.MethodGet,
		routes.RequestStatus(":chainID", ":reqID"),
		map[string]string{
			"chainID": chainID.Hex(),
			"reqID":   reqID.Base58(),
		},
		nil,
		&res,
		http.StatusOK,
	)

	require.True(t, res.IsProcessed)
}
