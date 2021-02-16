// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"net/http"

	"github.com/iotaledger/wasp/plugins/config"
	"github.com/labstack/echo/v4"
)

func configInit(e *echo.Echo, r renderer) Tab {
	route := e.GET("/", handleConfig)
	route.Name = "config"

	r[route.Path] = makeTemplate(e, tplConfig)

	return Tab{
		Path:  route.Path,
		Title: "Configuration",
		Href:  route.Path,
	}
}

func handleConfig(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &ConfigTemplateParams{
		BaseTemplateParams: BaseParams(c),
		Configuration:      config.Dump(),
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
