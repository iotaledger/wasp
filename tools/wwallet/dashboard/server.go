package dashboard

import (
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

func StartServer(listenAddr string, scs []SCDashboard) {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${method} ${uri} ${status} error="${error}"` + "\n",
	}))
	e.Use(middleware.Recover())
	e.HideBanner = true

	renderer := Renderer{
		"index": initIndexTemplate(),
	}
	e.Renderer = renderer
	for _, d := range scs {
		d.AddTemplates(renderer)
		navPages = append(navPages, NavPage{Title: d.Config().Name, Href: d.Config().Href()})
	}

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${time_rfc3339} ${level}")
	}
	e.Logger.SetLevel(log.INFO)

	e.GET("/", handleIndex)
	e.GET("/wwallet.json", handleWwalletJson)
	for _, d := range scs {
		d.AddEndpoints(e)
		addWebSocketTab(e, d.Config())
	}

	availableSCs := make(map[string]*sc.Config)
	for _, d := range scs {
		availableSCs[d.Config().Address().String()] = d.Config()
	}

	done := startNanomsgForwarder(e.Logger, availableSCs)
	defer func() { done <- true }()

	e.Logger.Fatal(e.Start(listenAddr))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
