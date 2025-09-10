package em

import (
	"context"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// UpDownCounter is a synchronous Instrument which supports increments and decrements.
// Complete docs:
// https://opentelemetry.io/docs/specs/otel/metrics/api/#updowncounter
type UpDownCounter[T n64] struct {
	attrs []attribute.KeyValue
	baseAdd[T]
}

// Add records a change of n to the counter.
func (u *UpDownCounter[T]) Add(n T, opts ...metric.AddOption) {
	u.AddCtx(context.Background(), n, opts...)
}

// AddCtx records a change of n to the counter.
func (u *UpDownCounter[T]) AddCtx(ctx context.Context, n T, opts ...metric.AddOption) {
	if u.baseAdd.Add == nil {
		return
	}
	u.baseAdd.Add(ctx, n, append(opts, metric.WithAttributes(u.attrs...))...)
}

// Inc records a change of 1 to the counter.
func (u *UpDownCounter[T]) Inc(opts ...metric.AddOption) {
	u.AddCtx(context.Background(), 1, opts...)
}

// IncCtx records a change of 1 to the counter.
func (u *UpDownCounter[T]) IncCtx(ctx context.Context, opts ...metric.AddOption) {
	u.AddCtx(ctx, 1, opts...)
}

// Dec records a change of -1 to the counter.
func (u *UpDownCounter[T]) Dec(opts ...metric.AddOption) {
	u.AddCtx(context.Background(), -1, opts...)
}

// DecCtx records a change of -1 to the counter.
func (u *UpDownCounter[T]) DecCtx(ctx context.Context, opts ...metric.AddOption) {
	u.AddCtx(ctx, 1, opts...)
}

func (u *UpDownCounter[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	id, err := getID(field)
	if err != nil {
		return err
	}
	u.attrs = attrs

	if useFloat[T]() {
		base, err := p.m.Float64UpDownCounter(id)
		if err != nil {
			return err
		}
		u.baseAdd.Add = func(ctx context.Context, t T, option ...metric.AddOption) {
			base.Add(ctx, float64(t), append(option, metric.WithAttributes(u.attrs...))...)
		}
	} else {
		base, err := p.m.Int64UpDownCounter(id)
		if err != nil {
			return err
		}
		u.baseAdd.Add = func(ctx context.Context, t T, option ...metric.AddOption) {
			base.Add(ctx, int64(t), append(option, metric.WithAttributes(u.attrs...))...)
		}
	}

	return nil
}
