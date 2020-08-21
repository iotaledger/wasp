package dashboard

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo"
)

func handleIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "index", &IndexTemplateParams{
		baseParams(c, "index"),
	})
}

type IndexTemplateParams struct {
	BaseTemplateParams
}

func initIndexTemplate() *template.Template {
	return makeTemplate(tplIndex)
}

const tplIndex = `
{{define "title"}}Index{{end}}

{{define "body"}}
{{end}}
`
