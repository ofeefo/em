package em

import (
	"context"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/attribute"
	_ "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// timeSub are functions that returns the time elapsed since a given starting
// point.
type timeSub[T n64] func(start time.Time) T

// n64 represents available types for creating metrics.
type n64 interface {
	int64 | float64
}

type buildable interface {
	init(field reflect.StructField, attrs ...attribute.KeyValue) error
}

type RecordFn[T n64] func(ctx context.Context, value T, opts ...metric.RecordOption)

type baseRecord[T n64] struct {
	Record RecordFn[T]
}

type baseAdd[T n64] struct {
	Add AddTFn[T]
}

type AddTFn[T n64] func(context.Context, T, ...metric.AddOption)

func useFloat[T any]() bool {
	return reflect.TypeFor[T]().Kind() == reflect.Float64
}
