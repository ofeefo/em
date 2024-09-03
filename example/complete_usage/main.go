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

	// Nested and embedded structs are supported. The instruments of the struct
	// will have both the attributes provided when registering the parent
	// struct (if any) and the attributes provided through the 'attrs' tag.
	*Embedded `attrs:"sub,embedded"`
	Nested    nested `attrs:"sub,nested"`
}

type nested struct {
	Counter  em.F64Counter `id:"example_nested_counter"`
	Gauge    em.F64Gauge   `id:"example_nested_gauge"`
	MoreNest struct {
		Counter em.F64Counter `id:"example_more_nested_counter"`
	}
}
type Embedded struct {
	Histogram     em.I64Histogram     `id:"example_embedded_histogram"`
	UpDownCounter em.F64UpDownCounter `id:"example_embedded_updowncounter"`
}

func main() {
	// Setup creates a basic OpenTelemetry configuration to get you started quickly.
	// For more advanced configurations (e.g., exporters, resources), use SetupWithMeter.
	err := em.Setup("some-service", attribute.String("version", "0.0.1"))
	if err != nil {
		panic(err)
	}

	// After setup, all initialized instruments will share the same exporter.
	s, err := em.Init[samplers](attribute.String("layer", "1"))
	if err != nil {
		panic(err)
	}

	// You can initialize the same sampler more than once, but note that
	// if they share the same identifiers, your metrics may be overridden.
	// To avoid conflicts, add unique attributes to each sampler's measurements.
	s2, err := em.Init[samplers](attribute.String("layer", "2"))
	if err != nil {
		panic(err)
	}

	go func() {
		var i int64
		j := func() float64 { return float64(i) }
		for i = range 10 {
			// Layer 1 instruments
			s.Counter.Add(i, em.Attrs(attribute.String("your", "attr")))
			s.Gauge.Record(i, em.Attrs(attribute.String("your", "attr")))
			s.Histogram.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 1 nested instruments
			s.Nested.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s.Nested.Gauge.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s.Nested.MoreNest.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 1 embedded instruments
			s.Embedded.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s.Embedded.Histogram.Record(i, em.Attrs(attribute.String("your", "attr")))

			// Layer 2 instruments
			s2.Counter.Add(i, em.Attrs(attribute.String("your", "attr")))
			s2.Gauge.Record(i, em.Attrs(attribute.String("your", "attr")))
			s2.Histogram.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s2.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 2 nested instruments
			s2.Nested.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Nested.Gauge.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Nested.MoreNest.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 2 embedded instruments
			s2.Embedded.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Embedded.Histogram.Record(i, em.Attrs(attribute.String("your", "attr")))

			time.Sleep(1 * time.Second)
		}
	}()

	//  Serve your metrics.
	if err = http.ListenAndServe(":8080", promhttp.Handler()); err != nil {
		panic(err)
	}
}
