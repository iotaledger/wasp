package prometheus

import (
	echoprometheus "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"

	"github.com/prometheus/client_golang/prometheus"
)

func newRestAPICollector(e *echo.Echo) []prometheus.Collector {
	collectors := make([]prometheus.Collector, 0)
	if e != nil {
		p := echoprometheus.NewPrometheus("iota_wasp_restapi", nil)
		for _, m := range p.MetricsList {
			collectors = append(collectors, m.MetricCollector)
		}
		e.Use(p.HandlerFunc)
	}

	return collectors
}
