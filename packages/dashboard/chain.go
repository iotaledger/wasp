package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp/color"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chain.tmpl
var tplChain string

func chainBreadcrumb(e *echo.Echo, chainID iscp.ChainID) Tab {
	return Tab{
		Path:  e.Reverse("chain"),
		Title: fmt.Sprintf("Chain %.8s…", chainID.Base58()),
		Href:  e.Reverse("chain", chainID.Base58()),
	}
}

func (d *Dashboard) initChain(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid", d.handleChain)
	route.Name = "chain"
	r[route.Path] = d.makeTemplate(e, tplChain, tplWs)
}

func (d *Dashboard) handleChain(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	tab := chainBreadcrumb(c.Echo(), *chainID)

	result := &ChainTemplateParams{
		BaseTemplateParams: d.BaseParams(c, tab),
		ChainID:            chainID,
	}

	result.Record, err = d.wasp.GetChainRecord(chainID)
	if err != nil {
		return err
	}

	if result.Record != nil && result.Record.Active {
		result.State, err = d.wasp.GetChainState(chainID)
		if err != nil {
			return err
		}

		theChain := d.wasp.GetChain(chainID)

		result.Committee = theChain.GetCommitteeInfo()

		result.RootInfo, err = d.fetchRootInfo(theChain)
		if err != nil {
			return err
		}

		result.Accounts, err = d.fetchAccounts(theChain)
		if err != nil {
			return err
		}

		result.TotalAssets, err = d.fetchTotalAssets(theChain)
		if err != nil {
			return err
		}

		result.Blobs, err = d.fetchBlobs(theChain)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

func (d *Dashboard) fetchAccounts(ch chain.ChainCore) ([]*iscp.AgentID, error) {
	accs, err := d.wasp.CallView(ch, accounts.Contract.Hname(), accounts.FuncViewAccounts.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("accountsc view call failed: %v", err)
	}

	ret := make([]*iscp.AgentID, 0)
	for k := range accs {
		agentid, _, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, &agentid)
	}
	return ret, nil
}

func (d *Dashboard) fetchTotalAssets(ch chain.ChainCore) (color.Balances, error) {
	bal, err := d.wasp.CallView(ch, accounts.Contract.Hname(), accounts.FuncViewTotalAssets.Name, nil)
	if err != nil {
		return nil, err
	}
	return accounts.DecodeBalances(bal)
}

func (d *Dashboard) fetchBlobs(ch chain.ChainCore) (map[hashing.HashValue]uint32, error) {
	ret, err := d.wasp.CallView(ch, blob.Contract.Hname(), blob.FuncListBlobs.Name, nil)
	if err != nil {
		return nil, err
	}
	return blob.DecodeDirectory(ret)
}

type ChainState struct {
	Index             uint32
	Hash              hashing.HashValue
	Timestamp         int64
	ApprovingOutputID ledgerstate.OutputID
}

type ChainTemplateParams struct {
	BaseTemplateParams

	ChainID *iscp.ChainID

	Record      *registry.ChainRecord
	State       *ChainState
	RootInfo    RootInfo
	Accounts    []*iscp.AgentID
	TotalAssets color.Balances
	Blobs       map[hashing.HashValue]uint32
	Committee   *chain.CommitteeInfo
}
