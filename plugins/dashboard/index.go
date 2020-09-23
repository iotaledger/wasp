package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/config"
	"github.com/labstack/echo"
)

type indexNavPage struct{}

func initIndex() NavPage {
	return &indexNavPage{}
}

const indexRoute = "/"
const indexTplName = "index"

func (n *indexNavPage) Title() string { return "Index" }
func (n *indexNavPage) Href() string  { return indexRoute }

func (n *indexNavPage) AddTemplates(renderer Renderer) {
	renderer[indexTplName] = MakeTemplate(tplIndex)
}

func (n *indexNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(indexRoute, func(c echo.Context) error {
		return c.Render(http.StatusOK, indexTplName, &IndexTemplateParams{
			BaseTemplateParams: BaseParams(c, indexRoute),
			Configuration:      config.Dump(),
		})
	})
}

type IndexTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
}

const tplIndex = `
{{define "title"}}Node configuration{{end}}

{{define "body"}}
	<h2>Node configuration</h2>

	<table>
		<thead>
			<tr>
				<th>Key</th>
				<th>Value</th>
			</tr>
		</thead>
		<tbody>
		{{range $k, $v := .Configuration}}
			<tr>
				<td><code>{{$k}}</code></td>
				<td><code>{{$v}}</code></td>
			</tr>
		{{end}}
		</tbody>
	</table>
{{end}}
`
