package em

import (
	"context"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Counter[T n64] interface {
	// Add records a change of n to the counter.
	Add(n T, opts ...metric.AddOption)

	// AddCtx records a change of n to the counter.
	AddCtx(ctx context.Context, n T, opts ...metric.AddOption)

	// Inc records a change of 1 to the counter.
	Inc(opts ...metric.AddOption)

	// IncCtx records a change of 1 to the counter.
	IncCtx(ctx context.Context, opts ...metric.AddOption)
}

type counter[T n64] struct {
	attrs []attribute.KeyValue
	baseAdd[T]
}

func (c *counter[T]) Add(n T, opts ...metric.AddOption) {
	c.AddCtx(context.Background(), n, opts...)
}

func (c *counter[T]) AddCtx(ctx context.Context, n T, opts ...metric.AddOption) {
	c.baseAdd.Add(ctx, n, append(opts, metric.WithAttributes(c.attrs...))...)
}

func (c *counter[T]) Inc(opts ...metric.AddOption) {
	c.AddCtx(context.Background(), 1, opts...)
}

func (c *counter[T]) IncCtx(ctx context.Context, opts ...metric.AddOption) {
	c.AddCtx(ctx, 1, opts...)
}

func newCounter[T n64]() buildable { return new(counter[T]) }

func (c *counter[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	if p == nil {
		c.baseAdd = dummy[T]{}
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
		base, err = p.m.Float64Counter(id)
	} else {
		base, err = p.m.Int64Counter(id)
	}
	if err != nil {
		return err
	}

	c.baseAdd = base.(baseAdd[T])
	c.attrs = attrs
	return nil
}
