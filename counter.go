package em

import (
	"context"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Counter is a synchronous Instrument which supports non-negative increments.
// Complete docs:
// https://opentelemetry.io/docs/specs/otel/metrics/api/#counter
type Counter[T n64] struct {
	attrs []attribute.KeyValue
	baseAdd[T]
}

// Add records a change of n to the counter.
func (c *Counter[T]) Add(n T, opts ...metric.AddOption) {
	c.AddCtx(context.Background(), n, opts...)
}

// AddCtx records a change of n to the counter.
func (c *Counter[T]) AddCtx(ctx context.Context, n T, opts ...metric.AddOption) {
	if c.baseAdd.Add == nil {
		return
	}
	c.baseAdd.Add(ctx, n, append(opts, metric.WithAttributes(c.attrs...))...)
}

// Inc records a change of 1 to the counter.
func (c *Counter[T]) Inc(opts ...metric.AddOption) {
	c.AddCtx(context.Background(), 1, opts...)
}

// IncCtx records a change of 1 to the counter.
func (c *Counter[T]) IncCtx(ctx context.Context, opts ...metric.AddOption) {
	c.AddCtx(ctx, 1, opts...)
}

func (c *Counter[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	id, err := getID(field)
	if err != nil {
		return err
	}
	c.attrs = attrs

	if useFloat[T]() {
		base, err := p.m.Float64Counter(id)
		if err != nil {
			return err
		}
		c.baseAdd.Add = func(ctx context.Context, t T, option ...metric.AddOption) {
			base.Add(ctx, float64(t), append(option, metric.WithAttributes(c.attrs...))...)
		}
	} else {
		base, err := p.m.Int64Counter(id)
		if err != nil {
			return err
		}
		c.baseAdd.Add = func(ctx context.Context, t T, option ...metric.AddOption) {
			base.Add(ctx, int64(t), append(option, metric.WithAttributes(c.attrs...))...)
		}
	}

	return nil
}
