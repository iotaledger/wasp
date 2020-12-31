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
	return t[name].ExecuteTemplate(w, "base", data)
}
