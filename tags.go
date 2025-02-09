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

func getID(f reflect.StructField) (string, error) {
	id := f.Tag.Get(idTag)
	if id == "" {
		return "", fmt.Errorf("missing id tag for field %s", f.Name)
	}
	return id, nil
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
