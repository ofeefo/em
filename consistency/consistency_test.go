package consistency

import (
	"bytes"
	"fmt"
	"github.com/ofeefo/em"
	ioprometheusclient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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

// This test will leverage the project's complete_example.
// All differentiation can be found there.
func TestLabelConsistency(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	if filepath.Base(wd) != "consistency" {
		wd = filepath.Join("consistency")
	}
	// sample.txt is the raw payload of the complete_example /metrics endpoint.
	dataPath := filepath.Join(wd, "raw.txt")
	data, err := os.ReadFile(dataPath)
	require.NoError(t, err)

	ids := getAllIds(t, metrics{})

	p := &expfmt.TextParser{}
	mfs, err := p.TextToMetricFamilies(bytes.NewBuffer(data))
	require.NoError(t, err)

	for _, id := range ids {
		mf, ok := mfs[id]
		require.True(t, ok)

		// There are two different samplers on the complete_example, so we're
		// expected to have two different metrics for each identifier.
		m := mf.GetMetric()
		require.Len(t, m, 2)
		ensureLayersFound(t, m)

		if strings.HasPrefix(id, "example_nested") {
			ensureNestedLabels(t, m)
		}

		if strings.HasPrefix(id, "example_embedded") {
			ensureEmbedded(t, m)
		}
	}
}

// Layers are the attribute differentiating the samplers on the complete_example.
func ensureLayersFound(t *testing.T, metrics []*ioprometheusclient.Metric) {
	found := map[string]bool{}
	for _, m := range metrics {
		for _, l := range m.GetLabel() {
			if l.GetName() != "layer" {
				continue
			}

			val := l.GetValue()
			_, ok := found[val]
			require.Falsef(t, ok, "Label values should not repeat in different label sets")
			found[val] = true
			break
		}
	}
	require.Len(t, found, 2)
}

func ensureNestedLabels(t *testing.T, metrics []*ioprometheusclient.Metric) {
	found := 0
	for _, m := range metrics {
		for _, l := range m.GetLabel() {
			if l.GetName() != "sub" {
				continue
			}
			require.Equal(t, l.GetValue(), "nested")
			found++
		}
	}
	require.Equal(t, 2, found)
}

func ensureEmbedded(t *testing.T, metrics []*ioprometheusclient.Metric) {
	found := 0
	for _, m := range metrics {
		for _, l := range m.GetLabel() {
			if l.GetName() != "sub" {
				continue
			}
			require.Equal(t, l.GetValue(), "embedded")
			found++
		}
	}
	require.Equal(t, 2, found)
}

func getAllIds(t *testing.T, bases ...any) []string {
	res := []string{}
	for _, b := range bases {
		ids := getAllIdsOf(t, b)
		for k := range ids {
			res = append(res, k)
		}
	}
	return res
}

func getAllIdsOf(t *testing.T, base any) map[string]struct{} {
	bType := reflect.TypeOf(base)
	bValue := reflect.ValueOf(base)
	if bType.Kind() != reflect.Struct {
		if bType.Kind() != reflect.Ptr {
			t.Fatalf("base is not a struct or a pointer")
		}

		bType = bType.Elem()
		bValue = bValue.Elem()
	}

	ids := make(map[string]struct{}, bType.NumField())

	for i := 0; i < bType.NumField(); i++ {
		field := bType.Field(i)
		fType := field.Type
		if fType.Kind() == reflect.Struct || fType.Kind() == reflect.Pointer {
			innerBase := bValue.Field(i)
			if reflect.Indirect(innerBase).Kind() == reflect.Ptr {
				innerBase = reflect.New(field.Type.Elem())
			}

			innerIds := getAllIdsOf(t, innerBase.Interface())
			for id := range innerIds {
				ids[id] = struct{}{}
			}
			continue
		}

		id, ok := field.Tag.Lookup("id")
		require.True(t, ok)
		// prometheus registers counters with a name prefixed with '_total'.
		// UpDown counters for the test were registered with ids suffixed by
		// '_updowncounters' so we can easily separate them here
		if strings.Contains(id, "_counter") {
			id = fmt.Sprintf("%s_total", id)
		}
		ids[id] = struct{}{}
	}
	return ids
}
