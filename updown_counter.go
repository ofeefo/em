package em

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"reflect"
)

type UpDownCounter[T n64] interface {
	// Add records a change of n to the counter.
	Add(n T, opts ...metric.AddOption)

	// AddCtx records a change of n to the counter.
	AddCtx(ctx context.Context, n T, opts ...metric.AddOption)

	// Inc records a change of 1 to the counter.
	Inc(opts ...metric.AddOption)

	// IncCtx records a change of 1 to the counter.
	IncCtx(ctx context.Context, opts ...metric.AddOption)

	// Dec records a change of -1 to the counter.
	Dec(opts ...metric.AddOption)

	// DecCtx records a change of -1 to the counter.
	DecCtx(ctx context.Context, opts ...metric.AddOption)
}

type upDownCounter[T n64] struct {
	attrs []attribute.KeyValue
	baseAdd[T]
}

func (u *upDownCounter[T]) Add(n T, opts ...metric.AddOption) {
	u.AddCtx(context.Background(), n, opts...)
}

func (u *upDownCounter[T]) AddCtx(ctx context.Context, n T, opts ...metric.AddOption) {
	u.baseAdd.Add(ctx, n, append(opts, metric.WithAttributes(u.attrs...))...)
}

func (u *upDownCounter[T]) Inc(opts ...metric.AddOption) {
	u.AddCtx(context.Background(), 1, opts...)
}

func (u *upDownCounter[T]) IncCtx(ctx context.Context, opts ...metric.AddOption) {
	u.AddCtx(ctx, 1, opts...)
}

func (u *upDownCounter[T]) Dec(opts ...metric.AddOption) {
	u.AddCtx(context.Background(), -1, opts...)
}

func (u *upDownCounter[T]) DecCtx(ctx context.Context, opts ...metric.AddOption) {
	u.AddCtx(ctx, 1, opts...)
}

func newUpDownCounter[T n64]() buildable { return new(upDownCounter[T]) }

func (u *upDownCounter[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	if p == nil {
		u.baseAdd = dummy[T]{}
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
		base, err = p.m.Float64UpDownCounter(id)
	} else {
		base, err = p.m.Int64UpDownCounter(id)
	}

	if err != nil {
		return err
	}

	u.baseAdd = base.(baseAdd[T])
	u.attrs = attrs
	return nil
}
