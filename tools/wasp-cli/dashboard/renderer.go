// +build ignore

package dashboard

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

type Renderer map[string]*template.Template

func (t Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t[name].ExecuteTemplate(w, "base", data)
}
