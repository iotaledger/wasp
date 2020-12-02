package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/plugins/chains"
	"github.com/labstack/echo"
)

type scNavPage struct{}

func initSc() NavPage {
	return &scNavPage{}
}

const scListRoute = "/smart-contracts"
const scListTplName = "scList"

const scRoute = "/smart-contracts/:address"
const scTplName = "sc"

func (n *scNavPage) Title() string { return "Smart Contracts" }
func (n *scNavPage) Href() string  { return scListRoute }

func (n *scNavPage) AddTemplates(renderer Renderer) {
	renderer[scTplName] = MakeTemplate(tplSc)
	renderer[scListTplName] = MakeTemplate(tplScList)
}

func (n *scNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(scListRoute, func(c echo.Context) error {
		scs, err := fetchSmartContracts()
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, scListTplName, &ScListTemplateParams{
			BaseTemplateParams: BaseParams(c, scListRoute),
			SmartContracts:     scs,
		})
	})

	e.GET(scRoute, func(c echo.Context) error {
		cid, err := coretypes.NewContractIDFromBase58(c.Param("address"))
		if err != nil {
			return err
		}

		result := &ScTemplateParams{
			BaseTemplateParams: BaseParams(c, scListRoute),
		}
		chainID := cid.ChainID()
		br, err := registry.GetChainRecord(&chainID)
		if err != nil {
			return err
		}
		if br != nil {
			result.ChainRecord = br
			state, batch, _, err := state.LoadSolidState(&chainID)
			if err != nil {
				return err
			}
			result.State = state
			result.Batch = batch
			if state != nil {
				vars := state.Variables()
				result.ProgramHash, _, err = codec.DecodeHashValue(vars.MustGet(vmconst.VarNameProgramData))
				if err != nil {
					return err
				}
				result.Description, _, err = codec.DecodeString(vars.MustGet(vmconst.VarNameDescription))
				if err != nil {
					return err
				}
				result.MinimumReward, _, err = codec.DecodeInt64(vars.MustGet(vmconst.VarNameMinimumReward))
				if err != nil {
					return err
				}
			}
			result.Committee = chains.GetStatus(&br.ChainID)
		}

		return c.Render(http.StatusOK, scTplName, result)
	})
}

func fetchSmartContracts() ([]*SmartContractOverview, error) {
	r := make([]*SmartContractOverview, 0)
	brs, err := registry.GetChainRecords()
	if err != nil {
		return nil, err
	}
	for _, br := range brs {
		desc, err := fetchDescription(br)

		if err != nil {
			return nil, err
		}
		r = append(r, &SmartContractOverview{
			ChainRecord: br,
			Description: desc,
		})
	}
	return r, nil
}

func fetchDescription(br *registry.ChainRecord) (string, error) {
	chainID := br.ChainID
	state, _, _, err := state.LoadSolidState(&chainID)
	if err != nil || state == nil {
		return "", err
	}
	d, _, err := codec.DecodeString(state.Variables().MustGet(vmconst.VarNameDescription))
	return d, err
}

type ScListTemplateParams struct {
	BaseTemplateParams
	SmartContracts []*SmartContractOverview
}

type SmartContractOverview struct {
	ChainRecord *registry.ChainRecord
	Description string
}

const tplScList = `
{{define "title"}}Smart Contracts{{end}}

{{define "body"}}
	<h2>Smart Contracts</h2>
	<table>
		<thead>
			<tr>
				<th>Target / Description</th>
				<th>Status</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
		{{range $_, $sc := .SmartContracts}}
			<tr>
				<td><code>{{$sc.ChainRecord.Target}}</code><br/>{{$sc.Description}}</td>
				<td>{{if $sc.ChainRecord.Active}}active{{else}}inactive{{end}}</td>
				<td><a href="/smart-contracts/{{$sc.ChainRecord.Target}}">Details</a></td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`

type ScTemplateParams struct {
	BaseTemplateParams
	ChainRecord   *registry.ChainRecord
	State         state.VirtualState
	Batch         state.Block
	ProgramHash   *hashing.HashValue
	Description   string
	MinimumReward int64
	Committee     *chains.ChainStatus
}

const tplSc = `
{{define "title"}}Smart Contracts{{end}}

{{define "body"}}
	<h2>Smart Contract details</h2>

	{{if .ChainRecord}}
		<div>
			<h3>Chain record</h3>
			<p>Target: {{template "address" .ChainRecord.Target}}</p>
			<p>Owner address:   {{template "address" .ChainRecord.OriginatorAddress}}</p>
			<p>Color:           <code>{{.ChainRecord.Color}}</code></p>
			<p>Committee Nodes: <code>{{.ChainRecord.CommitteeNodes}}</code></p>
			<p>Access Nodes:    <code>{{.ChainRecord.AccessNodes}}</code></p>
			<p>Active:          <code>{{.ChainRecord.Active}}</code></p>
		</div>
	{{else}}
		<p>No chain record for address {{template "address" .Target}}</p>
	{{end}}
	<hr/>
	{{if .State}}
		<div>
			<h3>State</h3>
			<p>State index: <code>{{.State.BlockIndex}}</code></p>
			<p>Timestamp: <code>{{formatTimestamp .State.Timestamp}}</code></p>
			<p>State Hash: <code>{{.State.Hash}}</code></p>
			<p>SC Program Hash: <code>{{.ProgramHash}}</code></p>
			<p>SC Description: <code>{{.Description}}</code></p>
			<p>SC Minimum Reward: <code>{{.MinimumReward}}</code></p>
		</div>
	{{else}}
		<p>State is empty.</p>
	{{end}}
	<hr/>
	{{if .Block}}
		<div>
			<h3>Block</h3>
			<p>State Transaction ID: <code>{{.Block.AnchorTransactionID}}</code></p>
			<p>Timestamp: <code>{{formatTimestamp .Block.Timestamp}}</code></p>
			<p>Essence Hash: <code>{{.Block.EssenceHash}}</code></p>
			<div>
				<p>Requests: (<code>{{.Block.Size}}</code> total)</p>
				<ul>
				{{range $_, $reqId := .Block.RequestIDs}}
					<li><code>{{$reqId}}</code></li>
				{{end}}
				</ul>
			</div>
		</div>
	{{else}}
		<p>Block is empty.</p>
	{{end}}
	<hr/>
	{{if .Committee}}
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
	{{else}}
		<p>No committee available for this smart contract.</p>
	{{end}}
{{end}}
`
