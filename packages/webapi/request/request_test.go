package request

import (
	"net/http"
	"testing"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

func createMockedGetChain(t *testing.T) getChainFn {
	return func(chainID *chainid.ChainID) chain.ChainCore {
		return testchain.NewMockedChainCore(t, *chainID, testlogger.NewLogger(t))
	}
}

const foo = "foo"

func dummyOffledgerRequest() *request.RequestOffLedger {
	contract := coretypes.Hn("somecontract")
	entrypoint := coretypes.Hn("someentrypoint")
	args := requestargs.New(
		dict.Dict{foo: []byte("bar")},
	)
	return request.NewRequestOffLedger(contract, entrypoint, args)
}

func TestNewRequestBase64(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain: createMockedGetChain(t),
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": chainid.RandomChainID().Base58()},
		OffLedgerRequestBody{Request: dummyOffledgerRequest().Base64()},
		nil,
		http.StatusAccepted,
	)
}

func TestNewRequestBinary(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain: createMockedGetChain(t),
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": chainid.RandomChainID().Base58()},
		dummyOffledgerRequest().Bytes(),
		nil,
		http.StatusAccepted,
	)
}
