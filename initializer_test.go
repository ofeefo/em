package em

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestInitialize(t *testing.T) {
	type samplers struct {
		I64Counter   I64Counter       `id:"i64counter" kind:"c"`
		I64UDC       I64UpDownCounter `id:"i64udc" kind:"udc"`
		I64Gauge     I64Gauge         `id:"i64gauge" kind:"g"`
		I64Histogram I64Histogram     `id:"i64histogram" kind:"h" buckets:"1.0,2.0,3.0"`

		F64Counter   F64Counter       `id:"f64counter" kind:"c"`
		F64UDC       F64UpDownCounter `id:"f64udc" kind:"udc"`
		F64Gauge     F64Gauge         `id:"f64gauge" kind:"g"`
		F64Histogram F64Histogram     `id:"f64histogram" kind:"h" buckets:"1.0,2.0,3.0"`
		Embed        struct {
			EmbedCounter I64Counter `id:"embedcounter" kind:"c"`
		} `attrs:"embed,1"`
	}

	initAndCall := func(t *testing.T) {
		s, err := Init[samplers]()
		require.NoError(t, err)
		require.NotPanics(t, func() {
			s.I64Counter.Add(1, Attrs(attribute.String("some", "value")))
			s.I64UDC.Add(1, Attrs(attribute.String("some", "value")))
			s.I64Gauge.Record(1, Attrs(attribute.String("some", "value")))
			s.I64Histogram.Record(1, Attrs(attribute.String("some", "value")))
			s.F64Counter.Add(1, Attrs(attribute.String("some", "value")))
			s.F64UDC.Add(1, Attrs(attribute.String("some", "value")))
			s.F64Gauge.Record(1, Attrs(attribute.String("some", "value")))
			s.F64Histogram.Record(1, Attrs(attribute.String("some", "value")))
			s.Embed.EmbedCounter.Add(1, Attrs(attribute.String("some", "value")))
		})
	}
	t.Run("Does not require the provider to be initialized", func(t *testing.T) {
		initAndCall(t)
	})

	t.Run("Works as expected with an initialized provider", func(t *testing.T) {
		err := Setup("test")
		require.NoError(t, err)
		initAndCall(t)
	})
}

func TestGetID(t *testing.T) {
	invalid := struct {
		ID string
	}{}
	valid := struct {
		ID string `id:"id"`
	}{}

	t.Run("Fails when a field does not have a ID tag", func(t *testing.T) {
		field := getField0(t, invalid)
		id, err := getID(field)
		require.Equal(t, "", id)
		require.Error(t, err)
	})

	t.Run("Works as expected with fields that have an ID tag", func(t *testing.T) {
		expectedID := "id"
		field := getField0(t, valid)
		id, err := getID(field)
		require.Equal(t, expectedID, id)
		require.NoError(t, err)
	})

}

func TestGetBounds(t *testing.T) {
	t.Run("Returns empty bounds when there is no tag", func(t *testing.T) {
		empty := struct {
			Bounds string
		}{}

		field := getField0(t, empty)
		bounds, err := getBounds(field)
		require.NoError(t, err)
		require.Len(t, bounds, 0)
	})

	t.Run("Fails when bounds cannot be parsed", func(t *testing.T) {
		invalid := struct {
			Bounds string `buckets:"1.0,2.0,3.0,something"`
		}{}

		field := getField0(t, invalid)
		_, err := getBounds(field)
		require.Error(t, err)
	})

	t.Run("Correctly retrieve all buckets", func(t *testing.T) {
		expectedBounds := []float64{
			1.0,
			0.5874697321,
			5.343,
			0.9,
		}
		valid := struct {
			Bounds string `buckets:"1.0,0.5874697321, 5.343, 0.9"`
		}{}

		field := getField0(t, valid)
		bounds, err := getBounds(field)
		require.NoError(t, err)
		require.ElementsMatch(t, expectedBounds, bounds)
	})
}

func TestGetAttrs(t *testing.T) {
	t.Run("Returns empty attributes when there is no tag", func(t *testing.T) {
		empty := struct {
			Embed struct{}
		}{}

		field := getField0(t, empty)
		attrs, err := getAttrs(field)
		require.NoError(t, err)
		require.Len(t, attrs, 0)
	})

	t.Run("Fails with odd number of attributes", func(t *testing.T) {
		invalid := struct {
			Embed struct {
			} `attrs:"bad,attr,count"`
		}{}
		field := getField0(t, invalid)
		attrs, err := getAttrs(field)
		require.Error(t, err)
		require.Len(t, attrs, 0)
	})

	t.Run("Correctly retrieve all attributes", func(t *testing.T) {
		expectedAttrs := []attribute.KeyValue{
			attribute.String("foo", "bar"),
			attribute.String("bar", "baz"),
		}

		valid := struct {
			Embed struct{} `attrs:"foo,bar,bar,baz"`
		}{}
		field := getField0(t, valid)
		attrs, err := getAttrs(field)
		require.NoError(t, err)
		require.ElementsMatch(t, expectedAttrs, attrs)
	})

}

func getField0(t *testing.T, base any) reflect.StructField {
	bType := validateStruct(t, base)
	require.Truef(t, bType.NumField() == 1, "Provided base has more fields than expected")
	return bType.Field(0)
}

func validateStruct(t *testing.T, base any) reflect.Type {
	bType := reflect.TypeOf(base)
	if bType.Kind() != reflect.Struct {
		t.Fatalf("bType should be a struct")
	}
	return bType
}

func TestTypeAndKindFor(t *testing.T) {
	type testCase struct {
		base, expectedType, expectedKind string
	}

	cases := []testCase{
		{"F64Counter", f64Type, counter},
		{"F64Gauge", f64Type, gauge},
		{"F64Histogram", f64Type, histogram},
		{"F64UpDownCounter", f64Type, upDownCounter},
		{"I64Counter", i64Type, counter},
		{"I64Gauge", i64Type, gauge},
		{"I64Histogram", i64Type, histogram},
		{"I64UpDownCounter", i64Type, upDownCounter},
	}

	for _, c := range cases {
		typ, kind := typeAndKindFor(c.base)
		require.Equal(t, c.expectedType, typ)
		require.Equal(t, c.expectedKind, kind)
	}
}
