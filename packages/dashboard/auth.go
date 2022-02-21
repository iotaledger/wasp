// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"github.com/iotaledger/wasp/packages/authentication"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed templates/auth.tmpl
var tplLogin string

func (d *Dashboard) loginInit(e *echo.Echo, r renderer) Tab {
	e.GET(authentication.AuthRouteSuccess(), d.handleLoginSuccess)
	route := e.GET(authentication.AuthRoute(), d.handleLogin)
	route.Name = "auth"

	r[route.Path] = d.makeTemplate(e, tplLogin)

	return Tab{
		Path:  route.Path,
		Title: "Auth",
		Href:  route.Path,
	}
}

func (d *Dashboard) handleLogin(c echo.Context) error {
	return c.Render(http.StatusOK, c.Path(), &LoginTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Configuration:      d.wasp.ConfigDump(),
	})
}

func (d *Dashboard) handleLoginSuccess(c echo.Context) error {
	auth, ok := c.Get("auth").(*authentication.AuthContext)

	if !ok {
		return c.Redirect(http.StatusMovedPermanently, authentication.AuthRoute())
	}

	if auth.IsAuthenticated {
		return c.Redirect(http.StatusMovedPermanently, "/config")
	}

	return c.Redirect(http.StatusMovedPermanently, authentication.AuthRoute())
}

type LoginTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
}
