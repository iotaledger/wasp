package stateapi

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/labstack/echo"
	"net/http"
)

type ReqStateRequest struct {
	Address    string   `json:"address"`
	RequestIds []string `json:"requests"`
}

type ReqStateResponse struct {
	Address  string          `json:"address"`
	Requests map[string]bool `json:"processed"` // false - backlog, true - processed, not present in the map - unknown
	Error    string          `json:"error"`
}

func HandlerQueryRequestState(c echo.Context) error {
	var req ReqStateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &ReqStateResponse{Error: err.Error()})
	}
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ReqStateResponse{Error: err.Error()})
	}
	cmt := committees.CommitteeByAddress(addr)
	if cmt == nil {
		return c.JSON(http.StatusBadRequest, &ReqStateResponse{Error: fmt.Sprintf("smart contract unknown. Address: %s", addr.String())})
	}
	reqIds := make([]*sctransaction.RequestId, len(req.RequestIds))
	for i, str := range req.RequestIds {
		if reqIds[i], err = sctransaction.RequestIdFromBase58(str); err != nil {
			return c.JSON(http.StatusBadRequest, &ReqStateResponse{Error: err.Error()})
		}
	}
	resp := ReqStateResponse{
		Address:  req.Address,
		Requests: make(map[string]bool, len(req.RequestIds)),
	}
	for _, reqid := range reqIds {
		switch cmt.GetRequestProcessingStatus(reqid) {
		case committee.RequestProcessingStatusCompleted:
			resp.Requests[reqid.ToBase58()] = true
		case committee.RequestProcessingStatusBacklog:
			resp.Requests[reqid.ToBase58()] = false
		}
	}
	return c.JSON(http.StatusOK, &resp)
}
