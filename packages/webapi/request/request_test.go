package request

import (
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	util "github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

type mockedChain struct {
	*testchain.MockedChainCore
}

var (
	_ chain.Chain         = &mockedChain{}
	_ chain.ChainCore     = &mockedChain{} // from testchain.MockedChainCore
	_ chain.ChainEntry    = &mockedChain{}
	_ chain.ChainRequests = &mockedChain{}
	_ chain.ChainMetrics  = &mockedChain{}
)

// chain.ChainRequests implementation

func (m *mockedChain) GetRequestProcessingStatus(_ iscp.RequestID) chain.RequestProcessingStatus {
	panic("implement me")
}

func (m *mockedChain) AttachToRequestProcessed(func(iscp.RequestID)) (attachID *events.Closure) {
	panic("implement me")
}

func (m *mockedChain) DetachFromRequestProcessed(attachID *events.Closure) {
	panic("implement me")
}

// chain.ChainEntry implementation

func (m *mockedChain) ReceiveTransaction(_ *ledgerstate.Transaction) {
	panic("implement me")
}

func (m *mockedChain) ReceiveState(_ *ledgerstate.AliasOutput, _ time.Time) {
	panic("implement me")
}

func (m *mockedChain) Dismiss(_ string) {
	panic("implement me")
}

func (m *mockedChain) IsDismissed() bool {
	panic("implement me")
}

// chain.ChainMetrics implementation

func (m *mockedChain) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics {
	panic("implement me")
}

// private methods

func createMockedGetChain(t *testing.T) chains.ChainProvider {
	return func(chainID *iscp.ChainID) chain.Chain {
		chainCore := testchain.NewMockedChainCore(t, chainID, testlogger.NewLogger(t))
		chainCore.OnOffLedgerRequest(func(msg *messages.OffLedgerRequestMsgIn) {
			t.Logf("Offledger request %v received", msg)
		})
		return &mockedChain{chainCore}
	}
}

func getAccountBalanceMocked(_ chain.Chain, _ *iscp.AgentID) (colored.Balances, error) {
	return colored.NewBalancesForIotas(100), nil
}

func hasRequestBeenProcessedMocked(ret bool) hasRequestBeenProcessedFn {
	return func(_ chain.Chain, _ iscp.RequestID) (bool, error) {
		return ret, nil
	}
}

func newMockedAPI(t *testing.T) *offLedgerReqAPI {
	return &offLedgerReqAPI{
		getChain:                createMockedGetChain(t),
		getAccountBalance:       getAccountBalanceMocked,
		hasRequestBeenProcessed: hasRequestBeenProcessedMocked(false),
		requestsCache:           expiringcache.New(10 * time.Second),
	}
}

func testRequest(t *testing.T, instance *offLedgerReqAPI, chainID *iscp.ChainID, body interface{}, expectedStatus int) {
	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": chainID.Base58()},
		body,
		nil,
		expectedStatus,
	)
}

// Tests

func TestNewRequestBase64(t *testing.T) {
	instance := newMockedAPI(t)
	chainID := iscp.RandomChainID()
	body := model.OffLedgerRequestBody{Request: model.NewBytes(util.DummyOffledgerRequest(chainID).Bytes())}
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestNewRequestBinary(t *testing.T) {
	instance := newMockedAPI(t)
	chainID := iscp.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestRequestAlreadyProcessed(t *testing.T) {
	instance := newMockedAPI(t)
	instance.hasRequestBeenProcessed = hasRequestBeenProcessedMocked(true)

	chainID := iscp.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusBadRequest)
}

func TestWrongChainID(t *testing.T) {
	instance := newMockedAPI(t)
	body := util.DummyOffledgerRequest(iscp.RandomChainID()).Bytes()
	testRequest(t, instance, iscp.RandomChainID(), body, http.StatusBadRequest)
}
