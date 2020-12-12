package dashboard

import (
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

type chainsNavPage struct{}

func initChains() NavPage {
	return &chainsNavPage{}
}

const chainListRoute = "/chains"
const chainListTplName = "chainList"

const chainRoute = "/chains/:chainid"
const chainTplName = "chain"

func (n *chainsNavPage) Title() string { return "Chains" }
func (n *chainsNavPage) Href() string  { return chainListRoute }

func (n *chainsNavPage) AddTemplates(renderer Renderer) {
	renderer[chainTplName] = MakeTemplate(tplChain)
	renderer[chainListTplName] = MakeTemplate(tplChainList)
}

func (n *chainsNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(chainListRoute, func(c echo.Context) error {
		chains, err := fetchChains()
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, chainListTplName, &ChainListTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			Chains:             chains,
		})
	})

	e.GET(chainRoute, func(c echo.Context) error {
		chainid, err := coretypes.NewChainIDFromBase58(c.Param("chainid"))
		if err != nil {
			return err
		}

		result := &ChainTemplateParams{
			BaseTemplateParams: BaseParams(c, chainListRoute),
			ChainID:            chainid,
		}

		result.ChainRecord, err = registry.GetChainRecord(&chainid)
		if err != nil {
			return err
		}

		if result.ChainRecord != nil && result.ChainRecord.Active {
			_, result.Block, _, err = state.LoadSolidState(&chainid)
			if err != nil {
				return err
			}

			chain := chains.GetChain(chainid)

			result.Committee.Size = chain.Size()
			result.Committee.Quorum = chain.Quorum()
			result.Committee.NumPeers = chain.NumPeers()
			result.Committee.HasQuorum = chain.HasQuorum()
			result.Committee.PeerStatus = chain.PeerStatus()
			result.RootInfo, err = fetchRootInfo(chain)
			if err != nil {
				return err
			}

			result.Accounts, err = fetchAccounts(chain)
			if err != nil {
				return err
			}
		}

		return c.Render(http.StatusOK, chainTplName, result)
	})
}

func fetchChains() ([]*ChainOverview, error) {
	r := make([]*ChainOverview, 0)
	crs, err := registry.GetChainRecords()
	if err != nil {
		return nil, err
	}
	for _, cr := range crs {
		info, err := fetchRootInfo(chains.GetChain(cr.ChainID))
		if err != nil {
			return nil, err
		}
		r = append(r, &ChainOverview{
			ChainRecord: cr,
			RootInfo:    info,
		})
	}
	return r, nil
}

func callView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error) {
	vctx, err := viewcontext.NewFromDB(*chain.ID(), chain.Processors())
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to create context: %v", err))
	}

	ret, err := vctx.CallView(hname, coretypes.Hn(fname), nil)
	if err != nil {
		return nil, fmt.Errorf("root view call failed: %v", err)
	}
	return ret, nil

}

func fetchRootInfo(chain chain.Chain) (ret RootInfo, err error) {
	info, err := callView(chain, root.Interface.Hname(), root.FuncGetInfo, nil)
	if err != nil {
		err = fmt.Errorf("root view call failed: %v", err)
		return
	}

	ret.Contracts, err = root.DecodeContractRegistry(datatypes.NewMustMap(info, root.VarContractRegistry))
	if err != nil {
		err = fmt.Errorf("DecodeContractRegistry() failed: %v", err)
		return
	}

	ret.OwnerID, _, _ = codec.DecodeAgentID(info.MustGet(root.VarChainOwnerID))
	ret.Description, _, _ = codec.DecodeString(info.MustGet(root.VarDescription))
	return
}

func fetchAccounts(chain chain.Chain) ([]coretypes.AgentID, error) {
	accounts, err := callView(chain, accountsc.Interface.Hname(), accountsc.FuncAccounts, nil)
	if err != nil {
		return nil, fmt.Errorf("accountsc view call failed: %v", err)
	}

	ret := make([]coretypes.AgentID, 0)
	for k := range accounts {
		agentid, _, err := codec.DecodeAgentID([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, agentid)
	}
	return ret, nil
}

type ChainListTemplateParams struct {
	BaseTemplateParams
	Chains []*ChainOverview
}

type ChainOverview struct {
	ChainRecord *registry.ChainRecord
	RootInfo    RootInfo
}

type RootInfo struct {
	OwnerID     coretypes.AgentID
	Description string
	Contracts   map[coretypes.Hname]*root.ContractRecord
}

const tplChainList = `
{{define "title"}}Chains{{end}}

{{define "body"}}
	<h2>Chains</h2>
	<table>
		<thead>
			<tr>
				<th>ID</th>
				<th>Description</th>
				<th>#Nodes</th>
				<th>#Contracts</th>
				<th>Status</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
		{{range $_, $c := .Chains}}
			<tr>
				<td><abbr title="{{$c.ChainRecord.ChainID.String}}"><code>{{printf "%.8s" $c.ChainRecord.ChainID}}…</code></abbr></td>
				<td>{{printf "%.50s" $c.RootInfo.Description}}</td>
				<td>{{len $c.ChainRecord.CommitteeNodes}}</td>
				<td>{{len $c.RootInfo.Contracts}}</td>
				<td>{{if $c.ChainRecord.Active}}active{{else}}inactive{{end}}</td>
				<td><a href="/chains/{{$c.ChainRecord.ChainID}}">Details</a></td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`

type ChainTemplateParams struct {
	BaseTemplateParams

	ChainID coretypes.ChainID

	ChainRecord *registry.ChainRecord
	Block       state.Block
	RootInfo    RootInfo
	Accounts    []coretypes.AgentID
	Committee   struct {
		Size       uint16
		Quorum     uint16
		NumPeers   uint16
		HasQuorum  bool
		PeerStatus []*chain.PeerStatus
	}
}

const tplChain = `
{{define "title"}}Chain details{{end}}

{{define "body"}}
	<h2>Chain <tt>{{printf "%.8s" .ChainID}}…</tt></h2>

	{{if .ChainRecord}}
		<div>
			<p>ChainID: <code>{{.ChainRecord.ChainID}}</code></p>
			<p>Chain Address: {{template "address" .ChainRecord.ChainID.Address}}</p>
			<p>Chain Color: <code>{{.ChainRecord.Color}}</code></p>
			<p>Active: <code>{{.ChainRecord.Active}}</code></p>
			{{if .ChainRecord.Active}}
				<p>Owner ID: {{template "agentid" .RootInfo.OwnerID}}</p>
				<p>Description: <code>{{.RootInfo.Description}}</code></p>
			{{end}}
		</div>
		{{if .ChainRecord.Active}}
			<hr/>
			<div>
				<h3>Contracts</h3>
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Description</th>
							<th>Program Hash</th>
						</tr>
					</thead>
					<tbody>
					{{range $_, $c := .RootInfo.Contracts}}
						<tr>
							<td><code>{{printf "%.15s" $c.Name}}</code></td>
							<td>{{printf "%.50s" $c.Description}}</td>
							<td><code>{{$c.ProgramHash.Short}}</code></td>
						</tr>
					{{end}}
					</tbody>
				</table>
			</div>

			<hr/>
			<div>
				<h3>Accounts</h3>
				<table>
					<thead>
						<tr>
							<th>AgentID</th>
						</tr>
					</thead>
					<tbody>
					{{range $_, $agentid := .Accounts}}
						<tr>
							<td>{{template "agentid" $agentid}}</td>
						</tr>
					{{end}}
					</tbody>
				</table>
			</div>

			<hr/>
			<div>
				<h3>Block</h3>
				<p>State index: <code>{{.Block.StateIndex}}</code></p>
				<p>State Transaction ID: <code>{{.Block.StateTransactionID}}</code></p>
				<p>Timestamp: <code>{{formatTimestamp .Block.Timestamp}}</code></p>
				<p>Essence Hash: <code>{{.Block.EssenceHash}}</code></p>
				<div>
					<table>
						<caption>Requests</caption>
						<thead>
							<tr>
								<th>RequestID</th>
							</tr>
						</thead>
						<tbody>
						{{range $_, $reqId := .Block.RequestIDs}}
							<tr>
								<td><code>{{$reqId}}</code></td>
							</tr>
						{{end}}
						</tbody>
					</table>
				</div>
			</div>

			<hr/>
			<div>
				<h3>Committee</h3>
				<p>Size:           <code>{{.Committee.Size}}</code></p>
				<p>Quorum:         <code>{{.Committee.Quorum}}</code></p>
				<p>NumPeers:       <code>{{.Committee.NumPeers}}</code></p>
				<p>HasQuorum:      <code>{{.Committee.HasQuorum}}</code></p>
				<table>
				<caption>Peer status</caption>
				<thead>
					<tr>
						<th>Index</th>
						<th>ID</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>
				{{range $_, $s := .Committee.PeerStatus}}
					<tr>
						<td>{{$s.Index}}</td>
						<td><code>{{$s.PeeringID}}</code></td>
						<td>{{if $s.Connected}}up{{else}}down{{end}}</td>
					</tr>
				{{end}}
				</tbody>
				</table>
			</div>
		{{end}}
	{{else}}
		<p>No chain record for ID <code>{{.ChainID}}</code></p>
	{{end}}
{{end}}
`
