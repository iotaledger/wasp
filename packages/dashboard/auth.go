// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared"
)

//go:embed templates/auth.tmpl
var tplLogin string

const headerXForwardedPrefix = "X-Forwarded-Prefix"

func (d *Dashboard) authInit(e *echo.Echo, r renderer) Tab {
	e.GET(shared.AuthRouteSuccess(), d.handleAuthCheck)
	e.GET("/", d.handleAuthCheck)

	route := e.GET(shared.AuthRoute(), d.RenderAuthView)
	route.Name = "auth"

	r[route.Path] = d.makeTemplate(e, tplLogin)

	return Tab{
		Path:  route.Path,
		Title: "Auth",
		Href:  route.Path,
	}
}

func (d *Dashboard) RenderAuthView(c echo.Context) error {
	auth, ok := c.Get("auth").(*authentication.AuthContext)

	if ok && auth.IsAuthenticated() {
		return d.redirect(c, "/config")
	}

	// TODO: Add sessions?
	loginError := c.QueryParam("error")

	return c.Render(http.StatusOK, "/auth", &AuthTemplateParams{
		BaseTemplateParams: d.BaseParams(c),
		Configuration:      d.wasp.ConfigDump(),
		LoginError:         loginError,
	})
}

func (d *Dashboard) handleAuthCheck(c echo.Context) error {
	auth, ok := c.Get("auth").(*authentication.AuthContext)

	if !ok {
		return d.redirect(c, shared.AuthRoute())
	}

	if auth.IsAuthenticated() {
		return d.redirect(c, "/config")
	}

	return d.redirect(c, shared.AuthRoute())
}

func (d *Dashboard) redirect(c echo.Context, uri string) error {
	return c.Redirect(http.StatusFound, c.Request().Header.Get(headerXForwardedPrefix)+uri)
}

type AuthTemplateParams struct {
	BaseTemplateParams
	Configuration map[string]interface{}
	LoginError    string
}
