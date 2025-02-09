package em

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

type metrics struct {
	Counter       Counter[int64]       `id:"i_am_a_counter"`
	Gauge         Gauge[int64]         `id:"i_am_a_gauge"`
	UpDownCounter UpDownCounter[int64] `id:"i_am_a_updowncounter"`
	Histogram     Histogram[float64]   `id:"i_am_a_histogram" buckets:"1.0,2.0,3.0"`
	Nested        nested               `attrs:"sub,nested,gotta,bar"`
	*Embedded     `attrs:"sub,embedded,gotta,bar2"`
}

type nested struct {
	Counter  Counter[float64] `id:"example_nested_counter"`
	Gauge    Gauge[float64]   `id:"example_nested_gauge"`
	MoreNest struct {
		Counter Counter[float64] `id:"example_more_nested_counter"`
	}
}
type Embedded struct {
	Histogram     Histogram[int64]     `id:"example_embedded_histogram"`
	UpDownCounter UpDownCounter[int64] `id:"example_embedded_updowncounter"`
}

func TestInit(t *testing.T) {
	t.Run("Without a configured provider", func(t *testing.T) {
		assert.NotPanics(t, initAndCall(t))
	})

	t.Run("With a configured provider", func(t *testing.T) {
		mustSetup(t)
		assert.NotPanics(t, initAndCall(t))
	})
}

func mustSetup(t *testing.T, attrs ...attribute.KeyValue) {
	t.Helper()
	require.NotPanics(t, func() {
		MustSetup("Test", attrs...)
	})
}

func initAndCall(t *testing.T) func() {
	t.Helper()
	return func() {
		var m *metrics
		require.NotPanics(t, func() { m = MustInit[metrics]() })
		require.NotPanics(t, func() {
			m.Counter.Add(1)
			m.UpDownCounter.Add(1)
			m.Gauge.Record(1)
			m.Histogram.Record(1)

			m.Nested.Counter.Add(1)
			m.Nested.Gauge.Record(1)
			m.Nested.MoreNest.Counter.Add(1)

			m.Embedded.UpDownCounter.Add(1)
			m.Embedded.Histogram.Record(1)
		})
	}
}
