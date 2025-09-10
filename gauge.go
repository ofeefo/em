package em

import (
	"context"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Gauge is a synchronous Instrument which can be used to record non-additive
// value(s).
// Complete docs: https://opentelemetry.io/docs/specs/otel/metrics/api/#gauge
type Gauge[T n64] struct {
	attrs []attribute.KeyValue
	baseRecord[T]
}

// Record records the value of n to the gauge.
func (g *Gauge[T]) Record(n T, opts ...metric.RecordOption) {
	g.RecordCtx(context.Background(), n, opts...)
}

// RecordCtx records the value of n to the gauge.
func (g *Gauge[T]) RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption) {
	if g.baseRecord.Record == nil {
		return
	}
	g.baseRecord.Record(ctx, n, append(opts, metric.WithAttributes(g.attrs...))...)
}

func (g *Gauge[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {

	id, err := getID(field)
	if err != nil {
		return err
	}
	g.attrs = attrs

	if useFloat[T]() {
		base, err := p.m.Float64Gauge(id)
		if err != nil {
			return err
		}
		g.baseRecord.Record = func(ctx context.Context, value T, opts ...metric.RecordOption) {
			base.Record(ctx, float64(value), append(opts, metric.WithAttributes(g.attrs...))...)
		}
	} else {
		base, err := p.m.Int64Gauge(id)
		if err != nil {
			return err
		}
		g.baseRecord.Record = func(ctx context.Context, value T, opts ...metric.RecordOption) {
			base.Record(ctx, int64(value), append(opts, metric.WithAttributes(g.attrs...))...)
		}
	}

	return nil
}
