// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed templates/login.tmpl
var tplLogin string

func (d *Dashboard) loginInit(e *echo.Echo, r renderer) Tab {
	route := e.GET("/login", d.handleLogin)
	route.Name = "login"

	r[route.Path] = d.makeTemplate(e, tplLogin)

	return Tab{
		Path:  route.Path,
		Title: "Login",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleLogin(c echo.Context) error {

	return c.Render(http.StatusOK, c.Path(), &LoginTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Configuration:      d.wasp.ConfigDump(),
	})
}

type LoginTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
}
