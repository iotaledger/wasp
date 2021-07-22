package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	server *http.Server
}

func New() *Metrics {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.GET("/metrics", func(c echo.Context) error {
		handler := promhttp.Handler()
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
	bindAddr := parameters.GetString(parameters.MetricsBindAddress)
	server := &http.Server{Addr: bindAddr, Handler: e}
	return &Metrics{server: server}
}

func (m *Metrics) Start() error {
	return m.server.ListenAndServe()
}

func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.server.Shutdown(ctx)
}
