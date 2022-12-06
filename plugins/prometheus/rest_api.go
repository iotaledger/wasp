package prometheus

import (
	echoprometheus "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func configureRestAPI(registry *prometheus.Registry, e *echo.Echo) {
	if e != nil {
		p := echoprometheus.NewPrometheus("iota_wasp_restapi", nil)
		for _, m := range p.MetricsList {
			registry.MustRegister(m.MetricCollector)
		}
		e.Use(p.HandlerFunc)
	}
}
