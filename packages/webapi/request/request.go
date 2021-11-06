package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type (
	getAccountBalanceFn       func(ch chain.Chain, agentID *iscp.AgentID) (colored.Balances, error)
	hasRequestBeenProcessedFn func(ch chain.Chain, reqID iscp.RequestID) (bool, error)
)

func AddEndpoints(
	server echoswagger.ApiRouter,
	getChain chains.ChainProvider,
	getChainBalance getAccountBalanceFn,
	hasRequestBeenProcessed hasRequestBeenProcessedFn,
	cacheTTL time.Duration,
	log *logger.Logger,
) {
	instance := &offLedgerReqAPI{
		getChain:                getChain,
		getAccountBalance:       getChainBalance,
		hasRequestBeenProcessed: hasRequestBeenProcessed,
		requestsCache:           expiringcache.New(cacheTTL),
		log:                     log,
	}
	server.POST(routes.NewRequest(":chainID"), instance.handleNewRequest).
		SetSummary("New off-ledger request").
		AddParamPath("", "chainID", "chainID represented in base58").
		AddParamBody(
			model.OffLedgerRequestBody{Request: "base64 string"},
			"Request",
			"Offledger Request encoded in base64. Optionally, the body can be the binary representation of the offledger request, but mime-type must be specified to \"application/octet-stream\"",
			false).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil)
}

type offLedgerReqAPI struct {
	getChain                chains.ChainProvider
	getAccountBalance       getAccountBalanceFn
	hasRequestBeenProcessed hasRequestBeenProcessedFn
	requestsCache           *expiringcache.ExpiringCache
	log                     *logger.Logger
}

func (o *offLedgerReqAPI) handleNewRequest(c echo.Context) error {
	chainID, offLedgerReq, err := parseParams(c)
	if err != nil {
		return err
	}

	// check req signature
	if !offLedgerReq.VerifySignature() {
		return httperrors.BadRequest("Invalid signature.")
	}

	// check chain exists
	ch := o.getChain(chainID)
	if ch == nil {
		return httperrors.NotFound(fmt.Sprintf("Unknown chain: %s", chainID.Base58()))
	}

	reqID := offLedgerReq.ID()

	if o.requestsCache.Get(reqID) != nil {
		return httperrors.BadRequest("request already processed")
	}

	alreadyProcessed, err := o.hasRequestBeenProcessed(ch, reqID)
	if err != nil {
		o.log.Errorf("webapi.offledger - check if already processed: %w", err)
		return httperrors.ServerError("internal error")
	}

	if alreadyProcessed {
		return httperrors.BadRequest("request already processed")
	}

	// check user has on-chain balance
	balances, err := o.getAccountBalance(ch, offLedgerReq.SenderAccount())
	if err != nil {
		o.log.Errorf("webapi.offledger - account balance: %w", err)
		return httperrors.ServerError("Unable to get account balance")
	}

	o.requestsCache.Set(reqID, true)

	if len(balances) == 0 {
		return httperrors.BadRequest(fmt.Sprintf("No balance on account %s", offLedgerReq.SenderAccount().Base58()))
	}
	ch.ReceiveOffLedgerRequest(offLedgerReq, "")

	return c.NoContent(http.StatusAccepted)
}

func parseParams(c echo.Context) (chainID *iscp.ChainID, req *request.OffLedger, err error) {
	chainID, err = iscp.ChainIDFromBase58(c.Param("chainID"))
	if err != nil {
		return nil, nil, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "json") {
		r := new(model.OffLedgerRequestBody)
		if err = c.Bind(r); err != nil {
			return nil, nil, httperrors.BadRequest("Error parsing request from payload")
		}
		rGeneric, err := request.FromMarshalUtil(marshalutil.New(r.Request.Bytes()))
		if err != nil {
			return nil, nil, httperrors.BadRequest(fmt.Sprintf("Error constructing off-ledger request from base64 string: %q", r.Request))
		}
		var ok bool
		if req, ok = rGeneric.(*request.OffLedger); !ok {
			return nil, nil, httperrors.BadRequest("Error parsing request: off-ledger request is expected")
		}
		return chainID, req, err
	}

	// binary format
	reqBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error parsing request from payload")
	}
	rGeneric, err := request.FromMarshalUtil(marshalutil.New(reqBytes))
	if err != nil {
		return nil, nil, httperrors.BadRequest("Error parsing request from payload")
	}
	req, ok := rGeneric.(*request.OffLedger)
	if !ok {
		return nil, nil, httperrors.BadRequest("Error parsing request: off-ledger request expected")
	}
	return chainID, req, err
}
