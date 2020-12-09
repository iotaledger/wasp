package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/config"
	"github.com/labstack/echo"
)

type configNavPage struct{}

func initConfig() NavPage {
	return &configNavPage{}
}

const configRoute = "/"
const configTplName = "config"

func (n *configNavPage) Title() string { return "Configuration" }
func (n *configNavPage) Href() string  { return configRoute }

func (n *configNavPage) AddTemplates(renderer Renderer) {
	renderer[configTplName] = MakeTemplate(tplConfig)
}

func (n *configNavPage) AddEndpoints(e *echo.Echo) {
	e.GET(configRoute, func(c echo.Context) error {
		return c.Render(http.StatusOK, configTplName, &ConfigTemplateParams{
			BaseTemplateParams: BaseParams(c, configRoute),
			Configuration:      config.Dump(),
		})
	})
}

type ConfigTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
}

const tplConfig = `
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
