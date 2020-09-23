package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/committees"
	"github.com/labstack/echo"
)

type committeesNavPage struct{}

func initCommittees() NavPage {
	return &committeesNavPage{}
}

const committeesRoute = "/committees"
const committeesTplName = "committees"

func (n *committeesNavPage) Title() string { return "Committees" }
func (n *committeesNavPage) Href() string  { return committeesRoute }

func (n *committeesNavPage) AddTemplates(renderer Renderer) {
	renderer[committeesTplName] = MakeTemplate(tplCommittees)
}

func (n *committeesNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(committeesRoute, func(c echo.Context) error {
		return c.Render(http.StatusOK, committeesTplName, &CommitteesTemplateParams{
			BaseTemplateParams: BaseParams(c, committeesRoute),
			Status:             committees.GetStatus(),
		})
	})
}

type CommitteesTemplateParams struct {
	BaseTemplateParams
	Status *committees.Status
}

const tplCommittees = `
{{define "title"}}Committees{{end}}

{{define "body"}}
	<h2>Committees</h2>
	<div>
	{{range $_, $c := .Status.Committees}}
		<details>
			<summary><code>{{$c.Address}}</code></summary>
			<p>Owner address:  <code>{{$c.OwnerAddress}}</code></p>
			<p>Color:          <code>{{$c.Color}}</code></p>
			<p>Size:           <code>{{$c.Size}}</code></p>
			<p>Quorum:         <code>{{$c.Quorum}}</code></p>
			<p>NumPeers:       <code>{{$c.NumPeers}}</code></p>
			<p>HasQuorum:      <code>{{$c.HasQuorum}}</code></p>
			<table>
			<caption>Peer status</caption>
			<thead>
				<tr>
					<th>Index</th>
					<th>NodeId</th>
					<th>IsSelf</th>
					<th>Connected</th>
				</tr>
			</thead>
			<tbody>
			{{range $_, $s := $c.PeerStatus}}
				<tr>
					<td><code>{{$s.Index}}</code></td>
					<td><code>{{$s.NodeId}}</code></td>
					<td><code>{{$s.IsSelf}}</code></td>
					<td><code>{{$s.Connected}}</code></td>
				</tr>
			{{end}}
			</tbody>
			</table>
		</details>
	{{else}}
		<p>(empty list)</p>
	{{end}}
	</div>
{{end}}
`
