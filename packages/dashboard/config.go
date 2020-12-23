// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/config"
	"github.com/labstack/echo"
)

type configNavPage struct{}

const configRoute = "/"
const configTplName = "config"

func (n *configNavPage) Title() string { return "Configuration" }
func (n *configNavPage) Href() string  { return configRoute }

func (n *configNavPage) AddTemplates(r renderer) {
	r[configTplName] = MakeTemplate(tplConfig)
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
<div class="card fluid">
	<h2 class="section">Node configuration</h2>

	<dl>
		{{range $k, $v := .Configuration}}
				<dt><tt>{{$k}}</tt></dt>
				<dd><tt>{{$v}}</tt></dd>
		{{end}}
	</dl>
</div>
{{end}}
`
