package em

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"reflect"
)

type Gauge[T n64] interface {
	// Record records the value of n to the gauge.
	Record(n T, opts ...metric.RecordOption)

	// RecordCtx records the value of n to the gauge.
	RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption)
}

type gauge[T n64] struct {
	attrs []attribute.KeyValue
	baseRecord[T]
}

func (g *gauge[T]) Record(n T, opts ...metric.RecordOption) {
	g.RecordCtx(context.Background(), n, opts...)
}

func (g *gauge[T]) RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption) {
	g.baseRecord.Record(ctx, n, append(opts, metric.WithAttributes(g.attrs...))...)
}

func newGauge[T n64]() buildable { return new(gauge[T]) }

func (g *gauge[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	if p == nil {
		g.baseRecord = dummy[T]{}
		return nil
	}

	var (
		base any
		err  error
	)

	id, err := getID(field)
	if err != nil {
		return err
	}

	if useFloat[T]() {
		base, err = p.m.Float64Gauge(id)
	} else {
		base, err = p.m.Int64Gauge(id)
	}

	if err != nil {
		return err
	}

	g.baseRecord = base.(baseRecord[T])
	g.attrs = attrs

	return nil
}
