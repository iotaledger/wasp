package reqstatus

import (
	"net/http"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
	"github.com/stretchr/testify/require"
)

type mockChain struct{}

func (m *mockChain) GetRequestProcessingStatus(id iscp.RequestID) chain.RequestProcessingStatus {
	return chain.RequestProcessingStatusCompleted
}

func (m *mockChain) EventRequestProcessed() *events.Event {
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
