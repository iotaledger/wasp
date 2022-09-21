// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dashboard

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type renderer map[string]*template.Template

func (t renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return template.Must(t[name].Clone()).Funcs(template.FuncMap{
		"uri": func(s string, p ...interface{}) string {
			return c.Request().Header.Get(headerXForwardedPrefix) + c.Echo().Reverse(s, p...)
		},
		"href": func(s string) string {
			return c.Request().Header.Get(headerXForwardedPrefix) + s
		},
	}).ExecuteTemplate(w, "base", data)
}
