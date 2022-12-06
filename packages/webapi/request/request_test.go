package request

import (
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	util "github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

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

func newMockedAPI() *offLedgerReqAPI {
	return &offLedgerReqAPI{
		getChain: func(chainID *isc.ChainID) chain.Chain {
			return &testutil.MockChain{}
		},
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

func TestNewRequestBase64(t *testing.T) {
	instance := newMockedAPI()
	chainID := isc.RandomChainID()
	body := model.OffLedgerRequestBody{Request: model.NewBytes(util.DummyOffledgerRequest(chainID).Bytes())}
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestNewRequestBinary(t *testing.T) {
	instance := newMockedAPI()
	chainID := isc.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusAccepted)
}

func TestRequestAlreadyProcessed(t *testing.T) {
	instance := newMockedAPI()
	instance.hasRequestBeenProcessed = hasRequestBeenProcessedMocked(true)

	chainID := isc.RandomChainID()
	body := util.DummyOffledgerRequest(chainID).Bytes()
	testRequest(t, instance, chainID, body, http.StatusBadRequest)
}

func TestWrongChainID(t *testing.T) {
	instance := newMockedAPI()
	body := util.DummyOffledgerRequest(isc.RandomChainID()).Bytes()
	testRequest(t, instance, isc.RandomChainID(), body, http.StatusBadRequest)
}
