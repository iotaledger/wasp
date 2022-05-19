package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chain.tmpl
var tplChain string

func chainBreadcrumb(e *echo.Echo, chainID *iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("chain"),
		Title: fmt.Sprintf("Chain %.8s…", chainID.String()),
		Href:  e.Reverse("chain", chainID.String()),
	}
}

func (d *Dashboard) initChain(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid", d.handleChain)
	route.Name = "chain"
	r[route.Path] = d.makeTemplate(e, tplChain, tplWebSocket)
}

func (d *Dashboard) handleChain(c echo.Context) error {
	chainID, err := iscp.ChainIDFromString(c.Param("chainid"))
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
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func (d *Dashboard) getLatestBlock(chainID *iscp.ChainID) (*LatestBlock, error) {
	ret, err := d.wasp.CallView(chainID, blocklog.Contract.Name, blocklog.ViewGetLatestBlockInfo.Name, nil)
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

func (d *Dashboard) fetchAccounts(chainID *iscp.ChainID) ([]iscp.AgentID, error) {
	accs, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.ViewAccounts.Name, nil)
	if err != nil {
		return nil, err
	}

	ret := make([]iscp.AgentID, 0)
	for k := range accs {
		agentid, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, agentid)
	}
	return ret, nil
}

func (d *Dashboard) fetchTotalAssets(chainID *iscp.ChainID) (*iscp.FungibleTokens, error) {
	bal, err := d.wasp.CallView(chainID, accounts.Contract.Name, accounts.ViewTotalAssets.Name, nil)
	if err != nil {
		return nil, err
	}
	return iscp.FungibleTokensFromDict(bal)
}

func (d *Dashboard) fetchBlobs(chainID *iscp.ChainID) (map[hashing.HashValue]uint32, error) {
	ret, err := d.wasp.CallView(chainID, blob.Contract.Name, blob.ViewListBlobs.Name, nil)
	if err != nil {
		return nil, err
	}
	return blob.DecodeDirectory(ret)
}

type LatestBlock struct {
	Index uint32
	Info  *blocklog.BlockInfo
}

type ChainTemplateParams struct {
	BaseTemplateParams

	ChainID *iscp.ChainID

	Record      *registry.ChainRecord
	LatestBlock *LatestBlock
	ChainInfo   *ChainInfo
	Accounts    []iscp.AgentID
	TotalAssets *iscp.FungibleTokens
	Blobs       map[hashing.HashValue]uint32
	Committee   *chain.CommitteeInfo
}
