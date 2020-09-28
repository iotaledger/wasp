package dashboard

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo"
)

func handleIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "index", &IndexTemplateParams{
		BaseParams(c, "index"),
	})
}

type IndexTemplateParams struct {
	BaseTemplateParams
}

func initIndexTemplate() *template.Template {
	return MakeTemplate(tplIndex)
}

const tplIndex = `
{{define "title"}}Index{{end}}

{{define "body"}}
{{end}}
`
