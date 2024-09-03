package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ofeefo/em"
)

// Define your instruments
type samplers struct {
	Counter       em.I64Counter       `id:"i_am_a_counter"`
	Gauge         em.I64Gauge         `id:"i_am_a_gauge"`
	UpDownCounter em.F64UpDownCounter `id:"i_am_a_updowncounter"`

	// Histograms may have the 'buckets' tag to define explicit boundaries.
	Histogram em.F64Histogram `id:"i_am_a_histogram" buckets:"1.0,2.0,3.0"`
}

func main() {
	// Setup creates a basic OpenTelemetry configuration for easy initialization.
	// For more advanced configurations (exporters, resources, etc.), use SetupWithMeter.
	err := em.Setup("some-service", attribute.String("version", "0.0.1"))
	if err != nil {
		panic(err)
	}

	// Once setup is complete, all instruments are registered with the same exporter.
	s, err := em.Init[samplers](attribute.String("layer", "1"))
	if err != nil {
		panic(err)
	}

	go func() {
		var i int64
		j := func() float64 { return float64(i) }
		for i = range 10 {
			// Record measurements for each instrument.
			s.Counter.Add(i, em.Attrs(attribute.String("your", "attr")))
			s.Gauge.Record(i, em.Attrs(attribute.String("your", "attr")))
			s.Histogram.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			time.Sleep(1 * time.Second)
		}
	}()

	//  Serve your metrics.
	if err = http.ListenAndServe(":8080", promhttp.Handler()); err != nil {
		panic(err)
	}
}
