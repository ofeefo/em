package main

import (
	em "github.com/ofeefo/em"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
	"time"
)

type metrics struct {
	Counter       em.Counter[int64]         `id:"i_am_a_counter"`
	Gauge         em.Gauge[int64]           `id:"i_am_a_gauge"`
	UpDownCounter em.UpDownCounter[float64] `id:"i_am_a_updowncounter"`
	Histogram     em.Histogram[float64]     `id:"i_am_a_histogram" buckets:"1.0,2.0,3.0"`
	Nested        nested                    `attrs:"sub,nested,gotta,bar"`
	*Embedded     `attrs:"sub,embedded,gotta,bar2"`
}

type nested struct {
	Counter  em.Counter[float64] `id:"example_nested_counter"`
	Gauge    em.Gauge[float64]   `id:"example_nested_gauge"`
	MoreNest struct {
		Counter em.Counter[float64] `id:"example_more_nested_counter"`
	}
}

type Embedded struct {
	Histogram     em.Histogram[int64]       `id:"example_embedded_histogram"`
	UpDownCounter em.UpDownCounter[float64] `id:"example_embedded_updowncounter"`
}

func main() {
	// Setup creates a basic OpenTelemetry configuration to get you started quickly.
	// For more advanced configurations (e.g., exporters, resources), use SetupWithMeter.
	err := em.Setup("some-service", attribute.String("version", "0.0.1"))
	if err != nil {
		panic(err)
	}

	// After setup, all initialized instruments will share the same exporter.
	s := em.MustInit[metrics](attribute.String("layer", "1"))

	// You can initialize the same sampler more than once, but note that
	// if they share the same identifiers, your metrics may be overridden.
	// To avoid conflicts, add unique attributes to each sampler's measurements.
	s2 := em.MustInit[metrics](attribute.String("layer", "2"))

	go func() {
		var i int64
		j := func() float64 { return float64(i) }
		done1 := s.Histogram.Measure(millis[float64], em.Attrs(attribute.String("your", "attr")))
		done2 := s2.Histogram.Measure(millis[float64], em.Attrs(attribute.String("your", "attr")))
		for i = range 10 {
			// Layer 1 instruments
			s.Counter.Add(i, em.Attrs(attribute.String("your", "attr")))
			s.Gauge.Record(i, em.Attrs(attribute.String("your", "attr")))
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
			s2.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 2 nested instruments
			s2.Nested.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Nested.Gauge.Record(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Nested.MoreNest.Counter.Add(j(), em.Attrs(attribute.String("your", "attr")))

			// Layer 2 embedded instruments
			// Layer 1 embedded instruments
			s2.Embedded.UpDownCounter.Add(j(), em.Attrs(attribute.String("your", "attr")))
			s2.Embedded.Histogram.Record(i, em.Attrs(attribute.String("your", "attr")))
			time.Sleep(100 * time.Millisecond)
		}
		done1(em.Attrs(attribute.String("your", "attr")))
		done2()
	}()

	//  Serve your metrics.
	if err = http.ListenAndServe(":8080", promhttp.Handler()); err != nil {
		panic(err)
	}
}

func millis[T float64 | int64](start time.Time) T {
	return T(time.Since(start).Milliseconds())
}
