package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

//go:embed templates/chain.tmpl
var tplChain string

func chainBreadcrumb(e *echo.Echo, chainID isc.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("chain"),
		Title: fmt.Sprintf("Chain %.8s…", chainID.String()),
		Href:  e.Reverse("chain", chainID.String()),
	}
}

func (d *Dashboard) initChain(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid", d.handleChain)
	route.Name = "chain"
	r[route.Path] = d.makeTemplate(e, tplChain)
}

func (d *Dashboard) handleChain(c echo.Context) error {
	chainID, err := isc.ChainIDFromString(c.Param("chainid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	tab := chainBreadcrumb(c.Echo(), chainID)

	result := &ChainTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		ChainID:            chainID,
	}

	result.Record, err = d.wasp.GetChainRecord(chainID)
	if err != nil {
		return err
	}

	if result.Record != nil && result.Record.Active {
		result.LatestBlock, err = d.getLatestBlock(chainID)
		if err != nil {
			return err
		}

		result.Committee, err = d.wasp.GetChainCommitteeInfo(chainID)
		if err != nil {
			return err
		}

		result.ChainInfo, err = d.fetchChainInfo(chainID)
		if err != nil {
			return err
		}

		result.Accounts, err = d.fetchAccounts(chainID)
		if err != nil {
			return err
		}

		result.TotalAssets, err = d.fetchTotalAssets(chainID)
		if err != nil {
			return err
		}

		result.Blobs, err = d.fetchBlobs(chainID)
		if err != nil {
			return err
		}

		result.EVMChainID, err = d.fetchEVMChainID(chainID)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func (d *Dashboard) getLatestBlock(chainID isc.ChainID) (*LatestBlock, error) {
	ret, err := d.wasp.CallView(chainID, blocklog.Contract.Name, blocklog.ViewGetBlockInfo.Name, nil)
	if err != nil {
		return nil, err
	}
	index, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex), 0)
	if err != nil {
		return nil, err
	}
	block, err := blocklog.BlockInfoFromBytes(index, ret.MustGet(blocklog.ParamBlockInfo))
	if err != nil {
		return nil, err
	}
	return &LatestBlock{Index: index, Info: block}, nil
}

func (d *Dashboard) fetchAccounts(chainID isc.ChainID) ([]isc.AgentID, error) {
	accs, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.ViewAccounts.Name, nil)
	if err != nil {
		return nil, err
	}

	ret := make([]isc.AgentID, 0)
	for k := range accs {
		agentid, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, agentid)
	}
	return ret, nil
}

func (d *Dashboard) fetchTotalAssets(chainID isc.ChainID) (*isc.FungibleTokens, error) {
	bal, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.ViewTotalAssets.Name, nil)
	if err != nil {
		return nil, err
	}
	return isc.FungibleTokensFromDict(bal)
}

func (d *Dashboard) fetchBlobs(chainID isc.ChainID) (map[hashing.HashValue]uint32, error) {
	ret, err := d.wasp.CallView(chainID, blob.Contract.Name, blob.ViewListBlobs.Name, nil)
	if err != nil {
		return nil, err
	}
	return blob.DecodeDirectory(ret)
}

func (d *Dashboard) fetchEVMChainID(chainID isc.ChainID) (uint16, error) {
	ret, err := d.wasp.CallView(chainID, evm.Contract.Name, evm.FuncGetChainID.Name, nil)
	if err != nil {
		return 0, err
	}
	return codec.DecodeUint16(ret.MustGet(evm.FieldResult))
}

type LatestBlock struct {
	Index uint32
	Info  *blocklog.BlockInfo
}

type ChainTemplateParams struct {
	BaseTemplateParams

	ChainID isc.ChainID

	EVMChainID  uint16
	Record      *registry.ChainRecord
	LatestBlock *LatestBlock
	ChainInfo   *ChainInfo
	Accounts    []isc.AgentID
	TotalAssets *isc.FungibleTokens
	Blobs       map[hashing.HashValue]uint32
	Committee   *chain.CommitteeInfo
}
