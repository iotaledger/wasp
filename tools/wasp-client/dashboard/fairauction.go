package dashboard

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo"
)

func handleFA(c echo.Context) error {
	return c.Render(http.StatusOK, "fairauction", &FATemplateParams{
		baseParams(c, "fairauction"),
	})
}

type FATemplateParams struct {
	BaseTemplateParams
}

func initFATemplate() *template.Template {
	t := template.Must(template.New("").Parse(tplBase))
	t = template.Must(t.Parse(tplFairAuction))
	return t
}

const tplFairAuction = `
{{define "title"}}FairRoulette{{end}}

{{define "body"}}
	<p>:)</p>
{{end}}
`
