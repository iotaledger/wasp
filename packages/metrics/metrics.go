package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	server *http.Server
	log    *logger.Logger
}

func New(log *logger.Logger) *Metrics {
	return &Metrics{log: log}
}

func (m *Metrics) Start(addr string) error {
	if m.server == nil {
		e := echo.New()
		e.HideBanner = true
		e.Use(middleware.Recover())
		e.GET("/metrics", func(c echo.Context) error {
			handler := promhttp.Handler()
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		})
		m.server = &http.Server{Addr: addr, Handler: e}
	}
	registerMempoolMetrics(m.log)
	m.log.Infof("Prometheus metrics accessible at: %s", addr)
	return m.server.ListenAndServe()
}

func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.server.Shutdown(ctx)
}
