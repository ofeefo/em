package em

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

const (
	idTag      = "id"
	bucketsTag = "buckets"
	attrsTag   = "attrs"
)

const (
	i64Type = "I"
	f64Type = "F"

	counter       = "Counter"
	upDownCounter = "UpDownCounter"
	gauge         = "Gauge"
	histogram     = "Histogram"
)

var (
	i64c      = reflect.TypeOf((*add[int64])(nil)).Elem()
	f64c      = reflect.TypeOf((*add[float64])(nil)).Elem()
	i64r      = reflect.TypeOf((*record[int64])(nil)).Elem()
	f64r      = reflect.TypeOf((*record[float64])(nil)).Elem()
	supported = []reflect.Type{i64c, i64r, f64c, f64r}
)

func typeAndKindFor(typeName string) (t, kind string) {
	t = string(typeName[0])
	kind = typeName[3:]
	return
}

func MustInit[T any](attrs ...attribute.KeyValue) *T {
	res, err := Init[T](attrs...)
	if err != nil {
		panic(err)
	}
	return res
}

func Init[T any](attrs ...attribute.KeyValue) (*T, error) {
	s := new(T)
	if err := initRef(s, attrs...); err != nil {
		return nil, err
	}
	return s, nil
}

func initRef(base any, attrs ...attribute.KeyValue) error {
	sType := reflect.TypeOf(base)
	sVal := reflect.ValueOf(base)
	if sType.Kind() == reflect.Pointer {
		sType = sType.Elem()
		sVal = sVal.Elem()
	}

	if sType.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", sType.Kind().String())
	}

	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		if !field.IsExported() {
			continue
		}

		fVal := sVal.Field(i)
		fTName := field.Type.Name()
		if fTName == "" {
			fTName = sType.Name()
		}

		// nolint: nestif
		if fVal.Kind() == reflect.Struct || fVal.Kind() == reflect.Ptr {
			innerAttrs, err := extractTag(field, getAttrs)
			if err != nil {
				return err
			}

			n := reflect.New(field.Type)
			isPtr := reflect.Indirect(n).Kind() == reflect.Ptr
			if isPtr {
				n = reflect.New(field.Type.Elem())
			}

			eAttrs := append(attrs, innerAttrs...)
			if err = initRef(n.Interface(), eAttrs...); err != nil {
				return fmt.Errorf("field initialization failed: %s", err)
			}

			if isPtr {
				fVal.Set(n)
			} else {
				fVal.Set(n.Elem())
			}
			continue
		}

		if implementsOneOf(field.Type, supported...) {
			t, kind := typeAndKindFor(fTName)
			val, err := initializeByKind(t, kind, field, attrs...)
			if err != nil {
				return fmt.Errorf("error initializing field: %s", err)
			}
			fVal.Set(reflect.ValueOf(val))
		}
	}
	return nil
}

func initializeByKind(t, kind string, field reflect.StructField, attrs ...attribute.KeyValue) (any, error) {
	var (
		id  string
		res any
		err error
	)

	id, err = extractTag(field, getID)
	if err != nil {
		return nil, err
	}

	switch kind {
	case counter, upDownCounter:
		if t == i64Type {
			res, err = prov.i64c(kind, id, attrs...)
		} else {
			res, err = prov.f64c(kind, id, attrs...)
		}
	case gauge, histogram:
		var bounds []float64
		if kind == histogram {
			bounds, err = extractTag(field, getBounds)
			if err != nil {
				return nil, err
			}
		}

		if t == i64Type {
			res, err = prov.i64r(kind, id, bounds, attrs...)
		} else {
			res, err = prov.f64r(kind, id, bounds, attrs...)
		}
	}

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, fmt.Errorf("unsupported kind found: %s", kind)
	}

	return res, nil
}

func getID(f reflect.StructField) (string, error) {
	id := f.Tag.Get(idTag)
	if id == "" {
		return "", fmt.Errorf("missing id tag for field %s", f.Name)
	}
	return id, nil
}

func getBounds(f reflect.StructField) ([]float64, error) {
	rawBounds := f.Tag.Get(bucketsTag)
	bounds := []float64{}
	if rawBounds == "" {
		return bounds, nil
	}

	sRawBounds := strings.Split(rawBounds, ",")
	bounds = make([]float64, 0, len(sRawBounds))
	for _, b := range sRawBounds {
		b = strings.TrimSpace(b)
		bucket, err := strconv.ParseFloat(b, 64)
		if err != nil {
			return nil, fmt.Errorf("failed parsing buckets [%s]: %s", rawBounds, err)
		}
		bounds = append(bounds, bucket)
	}
	return bounds, nil
}

func getAttrs(f reflect.StructField) ([]attribute.KeyValue, error) {
	rawAttrs := f.Tag.Get(attrsTag)
	attrs := []attribute.KeyValue{}
	if rawAttrs == "" {
		return attrs, nil
	}

	sAttrs := strings.Split(rawAttrs, ",")
	if len(sAttrs)%2 != 0 {
		return nil, fmt.Errorf("invalid number of attributes on field %s: %d", f.Name, len(sAttrs))
	}

	attrs = make([]attribute.KeyValue, 0, len(sAttrs)/2)
	for i := 0; i < len(sAttrs)-1; i += 2 {
		k := strings.TrimSpace(sAttrs[i])
		v := strings.TrimSpace(sAttrs[i+1])
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs, nil
}

func extractTag[T any](f reflect.StructField, fn func(f reflect.StructField) (T, error)) (T, error) {
	res, err := fn(f)
	if err != nil {
		return *new(T), err
	}
	return res, nil
}

func implementsOneOf(t1 reflect.Type, types ...reflect.Type) bool {
	for _, t := range types {
		if t1.Implements(t) {
			return true
		}
	}
	return false
}
