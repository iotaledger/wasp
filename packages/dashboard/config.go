// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed templates/config.tmpl
var tplConfig string

func (d *Dashboard) configInit(e *echo.Echo, r renderer) Tab {
	route := e.GET("/config", d.handleConfig)
	route.Name = "config"

	r[route.Path] = d.makeTemplate(e, tplConfig)

	return Tab{
		Path:  route.Path,
		Title: "Configuration",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleConfig(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &ConfigTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Configuration:      d.wasp.ConfigDump(),
	})
}

type ConfigTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
}
