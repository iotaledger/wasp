package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chaincontract.tmpl
var tplChainContract string

func initChainContract(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/contract/:hname", handleChainContract)
	route.Name = "chainContract"
	r[route.Path] = makeTemplate(e, tplChainContract, tplWs)
}

func handleChainContract(c echo.Context) error {
	chainID, err := coretypes.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	hname, err := coretypes.HnameFromString(c.Param("hname"))
	if err != nil {
		return err
	}

	result := &ChainContractTemplateParams{
		BaseTemplateParams: BaseParams(c, chainBreadcrumb(c.Echo(), *chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Contract %d", hname),
			Href:  "#",
		}),
		ChainID: *chainID,
		Hname:   hname,
	}

	chain := chains.AllChains().Get(chainID)
	if chain != nil {
		r, err := callView(chain, root.Interface.Hname(), root.FuncFindContract, codec.MakeDict(map[string]interface{}{
			root.ParamHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		result.ContractRecord, err = root.DecodeContractRecord(r[root.ParamData])
		if err != nil {
			return err
		}

		r, err = callView(chain, eventlog.Interface.Hname(), eventlog.FuncGetRecords, codec.MakeDict(map[string]interface{}{
			eventlog.ParamContractHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		records := collections.NewArrayReadOnly(r, eventlog.ParamRecords)
		result.Log = make([]*collections.TimestampedLogRecord, records.MustLen())
		for i := uint16(0); i < records.MustLen(); i++ {
			b := records.MustGetAt(i)
			result.Log[i], err = collections.ParseRawLogRecord(b)
			if err != nil {
				return err
			}
		}

		result.RootInfo, err = fetchRootInfo(chain)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainContractTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID
	Hname   coretypes.Hname

	ContractRecord *root.ContractRecord
	Log            []*collections.TimestampedLogRecord
	RootInfo       RootInfo
}
