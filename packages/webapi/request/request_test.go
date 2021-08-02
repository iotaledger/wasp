package request

import (
	"net/http"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

type mockedChain struct {
	*testchain.MockedChainCore
}

func (m *mockedChain) GetRequestProcessingStatus(_ iscp.RequestID) chain.RequestProcessingStatus {
	panic("implement me")
}

func (m *mockedChain) EventRequestProcessed() *events.Event {
	panic("implement me")
}

func (m *mockedChain) ReceiveTransaction(_ *ledgerstate.Transaction) {
	panic("implement me")
}

func (m *mockedChain) ReceiveInclusionState(_ ledgerstate.TransactionID, _ ledgerstate.InclusionState) {
	panic("implement me")
}

func (m *mockedChain) ReceiveState(_ *ledgerstate.AliasOutput, _ time.Time) {
	panic("implement me")
}

func (m *mockedChain) ReceiveOutput(_ ledgerstate.Output) {
	panic("implement me")
}

func (m *mockedChain) Dismiss(_ string) {
	panic("implement me")
}

func (m *mockedChain) IsDismissed() bool {
	panic("implement me")
}

func createMockedGetChain(t *testing.T) chains.ChainProvider {
	return func(chainID *iscp.ChainID) chain.Chain {
		return &mockedChain{
			testchain.NewMockedChainCore(t, *chainID, testlogger.NewLogger(t)),
		}
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

func dummyOffledgerRequest() *request.OffLedger {
	contract := iscp.Hn("somecontract")
	entrypoint := iscp.Hn("someentrypoint")
	args := requestargs.New(dict.Dict{})
	req := request.NewOffLedger(contract, entrypoint, args)
	keys, _ := testkey.GenKeyAddr()
	req.Sign(keys)
	return req
}

func TestNewRequestBase64(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain:                createMockedGetChain(t),
		getAccountBalance:       getAccountBalanceMocked,
		hasRequestBeenProcessed: hasRequestBeenProcessedMocked(false),
		requestsCache:           expiringcache.New(10 * time.Second),
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": iscp.RandomChainID().Base58()},
		model.OffLedgerRequestBody{Request: model.NewBytes(dummyOffledgerRequest().Bytes())},
		nil,
		http.StatusAccepted,
	)
}

func TestNewRequestBinary(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain:                createMockedGetChain(t),
		getAccountBalance:       getAccountBalanceMocked,
		hasRequestBeenProcessed: hasRequestBeenProcessedMocked(false),
		requestsCache:           expiringcache.New(10 * time.Second),
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": iscp.RandomChainID().Base58()},
		dummyOffledgerRequest().Bytes(),
		nil,
		http.StatusAccepted,
	)
}

func TestRequestAlreadyProcessed(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain:                createMockedGetChain(t),
		getAccountBalance:       getAccountBalanceMocked,
		hasRequestBeenProcessed: hasRequestBeenProcessedMocked(true),
		requestsCache:           expiringcache.New(10 * time.Second),
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": iscp.RandomChainID().Base58()},
		dummyOffledgerRequest().Bytes(),
		nil,
		http.StatusBadRequest,
	)
}
