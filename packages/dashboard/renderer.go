package dashboard

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

type renderer map[string]*template.Template

func (t renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t[name].ExecuteTemplate(w, "base", data)
}
