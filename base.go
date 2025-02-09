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

type baseRecord[T n64] interface {
	Record(ctx context.Context, value T, opts ...metric.RecordOption)
}

type baseAdd[T n64] interface {
	Add(ctx context.Context, n T, opts ...metric.AddOption)
}

func useFloat[T any]() bool {
	return reflect.TypeFor[T]().Kind() == reflect.Float64
}

type dummy[T n64] struct{}

func (d dummy[T]) Record(ctx context.Context, value T, opts ...metric.RecordOption) {}

func (d dummy[T]) Add(ctx context.Context, n T, opts ...metric.AddOption) {}
