package em

import (
	"fmt"
	"reflect"
	"sync"

	"go.opentelemetry.io/otel/attribute"
)

var bmMu = &sync.Mutex{}

var bmap = map[reflect.Type]func() buildable{
	reflect.TypeFor[Gauge[int64]]():   newGauge[int64],
	reflect.TypeFor[Gauge[float64]](): newGauge[float64],

	reflect.TypeFor[Counter[int64]]():   newCounter[int64],
	reflect.TypeFor[Counter[float64]](): newCounter[float64],

	reflect.TypeFor[Histogram[int64]]():   newHistogram[int64],
	reflect.TypeFor[Histogram[float64]](): newHistogram[float64],

	reflect.TypeFor[UpDownCounter[int64]]():   newUpDownCounter[int64],
	reflect.TypeFor[UpDownCounter[float64]](): newUpDownCounter[float64],
}

func MustInit[T any](attrs ...attribute.KeyValue) *T {
	res, err := Init[T](attrs...)
	if err != nil {
		panic(err)
	}
	return res
}

func Init[T any](attrs ...attribute.KeyValue) (*T, error) {
	base := new(T)
	if err := initRef(base, attrs...); err != nil {
		return nil, err
	}
	return base, nil
}

func initRef(base any, attrs ...attribute.KeyValue) error {
	bType := reflect.TypeOf(base)
	bVal := reflect.ValueOf(base)

	if bType.Kind() == reflect.Pointer {
		bType = bType.Elem()
		bVal = bVal.Elem()
	}

	if bType.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", bType.Kind().String())
	}

	for i := 0; i < bType.NumField(); i++ {
		field := bType.Field(i)
		fieldVal := bVal.Field(i)
		if !field.IsExported() {
			continue
		}
		bmMu.Lock()
		fn, ok := bmap[field.Type]
		bmMu.Unlock()
		n := reflect.New(field.Type)
		switch {
		case ok:
			builder := fn()
			if err := builder.init(field, attrs...); err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(builder))

		case n.Elem().Kind() == reflect.Struct:
			if err := initNested(field, n, fieldVal, attrs...); err != nil {
				return err
			}

		case n.Elem().Kind() == reflect.Ptr:
			n = reflect.New(field.Type.Elem())
			if err := initNested(field, n, fieldVal, attrs...); err != nil {
				return err
			}
		}
	}
	return nil
}

func initNested(field reflect.StructField, newVal, fieldVal reflect.Value, attrs ...attribute.KeyValue) error {
	innerAttrs, err := getAttrs(field)
	if err != nil {
		return err
	}

	if err = initRef(newVal.Interface(), append(attrs, innerAttrs...)...); err != nil {
		return err
	}

	if fieldVal.Kind() == reflect.Ptr {
		fieldVal.Set(newVal)
	} else {
		fieldVal.Set(newVal.Elem())
	}

	return nil
}
