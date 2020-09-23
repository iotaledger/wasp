package dashboard

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
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
	renderer[scTplName] = MakeTemplate(tplBootupRecord, tplSc)
	renderer[scListTplName] = MakeTemplate(tplBootupRecord, tplScList)
}

func (n *scNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(scListRoute, func(c echo.Context) error {
		brs, err := registry.GetBootupRecords()
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, scListTplName, &ScListTemplateParams{
			BaseTemplateParams: BaseParams(c, scListRoute),
			BootupRecords:      brs,
		})
	})

	e.GET(scRoute, func(c echo.Context) error {
		addr, err := address.FromBase58(c.Param("address"))
		if err != nil {
			return err
		}

		result := &ScTemplateParams{
			BaseTemplateParams: BaseParams(c, scListRoute),
			Address:            &addr,
		}

		br, err := registry.GetBootupData(&addr)
		if err != nil {
			return err
		}
		if br != nil {
			result.BootupRecord = br
			state, batch, _, err := state.LoadSolidState(&addr)
			if err != nil {
				return err
			}
			result.State = state
			result.Batch = batch
			if state != nil {
				codec := state.Variables().Codec()
				result.ProgramHash, _, err = codec.GetHashValue(vmconst.VarNameProgramHash)
				if err != nil {
					return err
				}
				result.Description, _, err = codec.GetString(vmconst.VarNameDescription)
				if err != nil {
					return err
				}
				result.MinimumReward, _, err = codec.GetInt64(vmconst.VarNameMinimumReward)
				if err != nil {
					return err
				}
			}
		}

		return c.Render(http.StatusOK, scTplName, result)
	})
}

type ScListTemplateParams struct {
	BaseTemplateParams
	BootupRecords []*registry.BootupData
}

const tplBootupRecord = `
{{define "bootup"}}
<p>Owner address:   <code>{{.OwnerAddress}}</code></p>
<p>Color:           <code>{{.Color}}</code></p>
<p>Committee Nodes: <code>{{.CommitteeNodes}}</code></p>
<p>Access Nodes:    <code>{{.AccessNodes}}</code></p>
<p>Active:          <code>{{.Active}}</code></p>
{{end}}
`

const tplScList = `
{{define "title"}}Smart Contracts{{end}}

{{define "body"}}
	<h2>Smart Contracts</h2>
	<div>
	{{range $_, $r := .BootupRecords}}
		<details>
			<summary><code>{{$r.Address}}</code></summary>
			{{template "bootup" $r}}
			<p><a href="/smart-contracts/{{$r.Address}}">Details</a></p>
		</details>
	{{else}}
		<p>(empty list)</p>
	{{end}}
	</div>
{{end}}
`

type ScTemplateParams struct {
	BaseTemplateParams
	Address       *address.Address
	BootupRecord  *registry.BootupData
	State         state.VirtualState
	Batch         state.Batch
	ProgramHash   *hashing.HashValue
	Description   string
	MinimumReward int64
}

const tplSc = `
{{define "title"}}Smart Contracts{{end}}

{{define "body"}}
	<h2>Smart Contract details</h2>

	{{if .BootupRecord}}
		<div>
			<h3>Bootup record</h3>
			<p>Address: <code>{{.BootupRecord.Address}}</code></p>
			{{template "bootup" .BootupRecord}}
		</div>
	{{else}}
		<p>No bootup record for address <code>{{.Address}}</code></p>
	{{end}}
	<hr/>
	{{if .State}}
		<div>
			<h3>State</h3>
			<p>State index: <code>{{.State.StateIndex}}</code></p>
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
	{{if .Batch}}
		<div>
			<h3>Batch</h3>
			<p>State Transaction ID: <code>{{.Batch.StateTransactionId}}</code></p>
			<p>Timestamp: <code>{{formatTimestamp .Batch.Timestamp}}</code></p>
			<p>Essence Hash: <code>{{.Batch.EssenceHash}}</code></p>
			<div>
				<p>Requests: (<code>{{.Batch.Size}}</code> total)</p>
				<ul>
				{{range $_, $reqId := .Batch.RequestIds}}
					<li><code>{{$reqId}}</code></li>
				{{end}}
				</ul>
			</div>
		</div>
	{{else}}
		<p>Batch is empty.</p>
	{{end}}
{{end}}
`
