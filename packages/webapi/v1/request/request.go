package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/hive.go/core/marshalutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/expiringcache"
	"github.com/iotaledger/wasp/packages/webapi/v1/httperrors"
	"github.com/iotaledger/wasp/packages/webapi/v1/model"
	"github.com/iotaledger/wasp/packages/webapi/v1/routes"
	"github.com/iotaledger/wasp/packages/webapi/v2/services"
)

type (
	shouldBeProcessedFn func(ch chain.ChainCore, req isc.OffLedgerRequest) error
)

func AddEndpoints(
	server echoswagger.ApiRouter,
	getChain chains.ChainProvider,
	nodePubKey *cryptolib.PublicKey,
	cacheTTL time.Duration,
) {
	instance := &offLedgerReqAPI{
		getChain:          getChain,
		shouldBeProcessed: services.ShouldBeProcessed,
		requestsCache:     expiringcache.New(cacheTTL),
		nodePubKey:        nodePubKey,
	}
	server.POST(routes.NewRequest(":chainID"), instance.handleNewRequest).
		SetDeprecated().
		SetSummary("Post an off-ledger request").
		AddParamPath("", "chainID", "chainID").
		AddParamBody(
			model.OffLedgerRequestBody{Request: "base64 string"},
			"Request",
			"Offledger Request encoded in base64. Optionally, the body can be the binary representation of the offledger request, but mime-type must be specified to \"application/octet-stream\"",
			false).
		AddResponse(http.StatusAccepted, "Request submitted", nil, nil)
}

type offLedgerReqAPI struct {
	getChain          chains.ChainProvider
	shouldBeProcessed shouldBeProcessedFn
	requestsCache     *expiringcache.ExpiringCache
	nodePubKey        *cryptolib.PublicKey
}

func (o *offLedgerReqAPI) handleNewRequest(c echo.Context) error {
	chainID, req, err := parseParams(c)
	if err != nil {
		return err
	}

	reqID := req.ID()

	if o.requestsCache.Get(reqID) != nil {
		return httperrors.BadRequest("request already processed")
	}
	o.requestsCache.Set(reqID, true)

	// check req signature
	if err := req.VerifySignature(); err != nil {
		o.requestsCache.Set(reqID, true)
		return httperrors.BadRequest(fmt.Sprintf("could not verify: %s", err.Error()))
	}

	// check req is for the correct chain
	if !req.ChainID().Equals(chainID) {
		// do not add to cache, it can still be sent to the correct chain
		return httperrors.BadRequest("Request is for a different chain")
	}

	// check chain exists
	ch := o.getChain(chainID)
	if ch == nil {
		return httperrors.NotFound(fmt.Sprintf("Unknown chain: %s", chainID.String()))
	}

	err = o.shouldBeProcessed(ch, req)
	if err != nil {
		return err
	}

	ch.ReceiveOffLedgerRequest(req, o.nodePubKey)

	return c.NoContent(http.StatusAccepted)
}

func parseParams(c echo.Context) (chainID isc.ChainID, req isc.OffLedgerRequest, err error) {
	chainID, err = isc.ChainIDFromString(c.Param("chainID"))
	if err != nil {
		return isc.ChainID{}, nil, httperrors.BadRequest(fmt.Sprintf("Invalid Chain ID %+v: %s", c.Param("chainID"), err.Error()))
	}

	contentType := c.Request().Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "json") {
		r := new(model.OffLedgerRequestBody)
		if err = c.Bind(r); err != nil {
			return isc.ChainID{}, nil, httperrors.BadRequest("error parsing request from payload")
		}
		rGeneric, err := isc.NewRequestFromMarshalUtil(marshalutil.New(r.Request.Bytes()))
		if err != nil {
			return isc.ChainID{}, nil, httperrors.BadRequest(fmt.Sprintf("cannot decode off-ledger request: %v", err))
		}
		var ok bool
		if req, ok = rGeneric.(isc.OffLedgerRequest); !ok {
			return isc.ChainID{}, nil, httperrors.BadRequest("error parsing request: off-ledger request is expected")
		}
		return chainID, req, err
	}

	// binary format
	reqBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return isc.ChainID{}, nil, httperrors.BadRequest("error parsing request from payload")
	}
	rGeneric, err := isc.NewRequestFromMarshalUtil(marshalutil.New(reqBytes))
	if err != nil {
		return isc.ChainID{}, nil, httperrors.BadRequest("error parsing request from payload")
	}
	req, ok := rGeneric.(isc.OffLedgerRequest)
	if !ok {
		return isc.ChainID{}, nil, httperrors.BadRequest("error parsing request: off-ledger request expected")
	}
	return chainID, req, err
}
