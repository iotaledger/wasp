package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chainblock.tmpl
var tplChainBlock string

func (d *Dashboard) initChainBlock(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/block/:index", d.handleChainBlock)
	route.Name = "chainBlock"
	r[route.Path] = d.makeTemplate(e, tplChainBlock, tplWebSocket)
}

func (d *Dashboard) handleChainBlock(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	result := &ChainBlockTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Block #%s", c.Param("index")),
			Href:  "#",
		}),
		ChainID: chainID,
		Index:   uint32(index),
	}

	latestBlock, err := d.getLatestBlock(chainID)
	if err != nil {
		return err
	}
	result.LatestBlockIndex = latestBlock.Index

	if uint32(index) == result.LatestBlockIndex {
		result.Block = latestBlock.Info
	} else {
		ret, err := d.wasp.CallView(chainID, blocklog.Contract.Name, blocklog.ViewGetBlockInfo.Name, dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(uint32(index)),
		})
		if err != nil {
			return err
		}
		result.Block, err = blocklog.BlockInfoFromBytes(uint32(index), ret.MustGet(blocklog.ParamBlockInfo))
		if err != nil {
			return err
		}
		if result.Block.L1Commitment == nil {
			// to please the template.
			result.Block.L1Commitment = state.L1CommitmentNil
		}
	}

	{
		ret, err := d.wasp.CallView(chainID, blocklog.Contract.Name, blocklog.ViewGetRequestReceiptsForBlock.Name, dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(uint32(index)),
		})
		if err != nil {
			return err
		}
		arr := collections.NewArray16ReadOnly(ret, blocklog.ParamRequestRecord)
		result.Receipts = make([]*blocklog.RequestReceipt, arr.MustLen())
		result.ResolvedErrors = make([]string, arr.MustLen())
		for i := uint16(0); i < arr.MustLen(); i++ {
			receipt, err := blocklog.RequestReceiptFromBytes(arr.MustGetAt(i))
			if err != nil {
				return err
			}
			result.Receipts[i] = receipt
			if receipt.Error != nil {
				resolved, err := errors.Resolve(receipt.Error, func(c string, f string, params dict.Dict) (dict.Dict, error) {
					return d.wasp.CallView(chainID, c, f, params)
				})
				if err != nil {
					return err
				}
				result.ResolvedErrors[i] = resolved.Error()
			}
		}
	}

	{
		ret, err := d.wasp.CallView(chainID, blocklog.Contract.Name, blocklog.ViewGetEventsForBlock.Name, dict.Dict{
			blocklog.ParamBlockIndex: codec.EncodeUint32(uint32(index)),
		})
		if err != nil {
			return err
		}
		arr := collections.NewArray16ReadOnly(ret, blocklog.ParamEvent)
		result.Events = make([]string, arr.MustLen())
		for i := uint16(0); i < arr.MustLen(); i++ {
			result.Events[i] = string(arr.MustGetAt(i))
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainBlockTemplateParams struct {
	BaseTemplateParams
	ChainID          *isc.ChainID
	Index            uint32
	LatestBlockIndex uint32
	Block            *blocklog.BlockInfo
	Receipts         []*blocklog.RequestReceipt
	ResolvedErrors   []string
	Events           []string
}
