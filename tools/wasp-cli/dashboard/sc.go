// +build ignore

package dashboard

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/sc"
	"github.com/labstack/echo"
)

type SCDashboard interface {
	Config() *sc.Config
	AddEndpoints(e *echo.Echo)
	AddTemplates(r Renderer)
}
