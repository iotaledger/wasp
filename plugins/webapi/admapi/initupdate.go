package admapi

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/iotaledger/wasp/plugins/webapi/misc"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
)

type InitScRequest struct {
	Address   string `json:"address"`
	BatchData string `json:"batch_data"`
}

func HandlerInitSC(c echo.Context) error {
	var req InitScRequest

	if err := c.Bind(&req); err != nil {
		return misc.OkJson(c, &misc.SimpleResponse{
			Error: err.Error(),
		})
	}

	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return misc.OkJson(c, &misc.SimpleResponse{
			Error: err.Error(),
		})
	}

	data, err := base58.Decode(req.BatchData)
	if err != nil {
		return misc.OkJson(c, &misc.SimpleResponse{
			Error: err.Error(),
		})
	}

	batch, err := state.BatchFromBytes(data)
	if err != nil {
		return misc.OkJson(c, &misc.SimpleResponse{
			Error: err.Error(),
		})
	}

	cmt := committees.CommitteeByAddress(addr)
	if cmt == nil {
		return misc.OkJson(c, &misc.SimpleResponse{
			Error: "committee not found",
		})
	}

	cmt.ReceiveMessage(committee.PendingBatchMsg{
		Batch: batch,
	})

	return misc.OkJson(c, &misc.SimpleResponse{})
}
