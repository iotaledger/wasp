package request

import (
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/core/events"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	util "github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

type mockedChain struct {
	*testchain.MockedChainCore
}

var _ chain.Chain = &mockedChain{}

// chain.ChainRequests implementation

func (m *mockedChain) ResolveError(e *isc.UnresolvedVMError) (*isc.VMError, error) {
	panic("implement me")
}

func (m *mockedChain) GetRequestReceipt(reqID isc.RequestID) (*blocklog.RequestReceipt, error) {
	panic("implement me")
}

func (m *mockedChain) AttachToRequestProcessed(func(isc.RequestID)) (attachID *events.Closure) {
	panic("implement me")
}

func (m *mockedChain) DetachFromRequestProcessed(attachID *events.Closure) {
	panic("implement me")
}

// chain.ChainEntry implementation

func (m *mockedChain) ReceiveTransaction(_ *iotago.Transaction) {
	panic("implement me")
}

func (m *mockedChain) ReceiveState(_ *iotago.AliasOutput, _ time.Time) {
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

func (m *mockedChain) GetConsensusWorkflowStatus() chain.ConsensusWorkflowStatus {
	panic("implement me")
}

func (m *mockedChain) GetConsensusPipeMetrics() chain.ConsensusPipeMetrics {
	panic("implement me")
}

// chain.ChainRunner implementation

func (*mockedChain) GetBranch() (chainstore.Branch, bool, error) {
	panic("unimplemented")
}

func (*mockedChain) GetTimeData() time.Time {
	panic("unimplemented")
}

// private methods

func createMockedGetChain(t *testing.T) chains.ChainProvider {
	return func(chainID *isc.ChainID) chain.Chain {
		chainCore := testchain.NewMockedChainCore(t, chainID, testlogger.NewLogger(t))
		chainCore.OnOffLedgerRequest(func(msg *messages.OffLedgerRequestMsgIn) {
			t.Logf("Offledger request %v received", msg)
		})
		return &mockedChain{chainCore}
	}
}

func getAccountBalanceMocked(_ chain.ChainCore, _ isc.AgentID) (*isc.FungibleTokens, error) {
	return isc.NewFungibleBaseTokens(100), nil
}

func hasRequestBeenProcessedMocked(ret bool) hasRequestBeenProcessedFn {
	return func(_ chain.ChainCore, _ isc.RequestID) (bool, error) {
		return ret, nil
	}
}

func checkNonceMocked(ch chain.ChainCore, req isc.OffLedgerRequest) error {
	return nil
}

func newMockedAPI(t *testing.T) *offLedgerReqAPI {
	return &offLedgerReqAPI{
		getChain:                createMockedGetChain(t),
		getAccountAssets:        getAccountBalanceMocked,
		hasRequestBeenProcessed: hasRequestBeenProcessedMocked(false),
		checkNonce:              checkNonceMocked,
		requestsCache:           expiringcache.New(10 * time.Second),
	}
}

func testRequest(t *testing.T, instance *offLedgerReqAPI, chainID *isc.ChainID, body interface{}, expectedStatus int) {
	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": chainID.String()},
		body,
		nil,
		expectedStatus,
	)
}

// Tests

func TestNewRequestBase64(t *testing.T) {
	instance := newMockedAPI(t)
	chainID := isc.RandomChainID()
	body := model.OffLedgerRequestBody{Request: model.NewBytes(util.DummyOffledgerRequest(chainID).Bytes())}
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestNewRequestBinary(t *testing.T) {
	instance := newMockedAPI(t)
	chainID := isc.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestRequestAlreadyProcessed(t *testing.T) {
	instance := newMockedAPI(t)
	instance.hasRequestBeenProcessed = hasRequestBeenProcessedMocked(true)

	chainID := isc.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusBadRequest)
}

func TestWrongChainID(t *testing.T) {
	instance := newMockedAPI(t)
	body := util.DummyOffledgerRequest(isc.RandomChainID()).Bytes()
	testRequest(t, instance, isc.RandomChainID(), body, http.StatusBadRequest)
}
