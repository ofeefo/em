package em

import (
	"context"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Histogram is a synchronous Instrument which can be used to report arbitrary
// values that are likely to be statistically meaningful. It is intended for
// statistics such as histograms, summaries, and percentile.
// Complete docs:
// https://opentelemetry.io/docs/specs/otel/metrics/api/#histogram
type Histogram[T n64] struct {
	attrs []attribute.KeyValue
	baseRecord[T]
}

// Record adds a value to the distribution.
func (h *Histogram[T]) Record(n T, opts ...metric.RecordOption) {
	h.RecordCtx(context.Background(), n, opts...)
}

// RecordCtx adds a value to the distribution.
func (h *Histogram[T]) RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption) {
	if h.baseRecord.Record == nil {
		return
	}
	h.baseRecord.Record(ctx, n, append(opts, metric.WithAttributes(h.attrs...))...)
}

// Measure creates a starting point using time.Now and returns a function
// that records the time elapsed since the starting point. The returned
// function may also receive record options to further support information
// that may vary along a given procedure.
func (h *Histogram[T]) Measure(sub timeSub[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption) {
	return h.MeasureCtx(context.Background(), sub, opts...)
}

// MeasureCtx creates a starting point using time.Now and returns a function
// that records the time elapsed since the starting point. The returned
// function may also receive record options to further support information
// that may vary along a given procedure.
func (h *Histogram[T]) MeasureCtx(ctx context.Context, sub timeSub[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption) {
	start := time.Now()
	baseOpts := opts
	return func(opts ...metric.RecordOption) {
		h.RecordCtx(ctx, sub(start), append(baseOpts, opts...)...)
	}
}

func (h *Histogram[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	id, err := getID(field)
	if err != nil {
		return err
	}

	h.attrs = attrs

	bounds, err := getBounds(field)
	if err != nil {
		return err
	}

	if useFloat[T]() {
		base, err := p.m.Float64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
		if err != nil {
			return err
		}
		h.baseRecord.Record = func(ctx context.Context, value T, opts ...metric.RecordOption) {
			base.Record(ctx, float64(value), append(opts, metric.WithAttributes(h.attrs...))...)
		}
	} else {
		base, err := p.m.Int64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
		if err != nil {
			return err
		}
		h.baseRecord.Record = func(ctx context.Context, value T, opts ...metric.RecordOption) {
			base.Record(ctx, int64(value), append(opts, metric.WithAttributes(h.attrs...))...)
		}
	}

	return nil
}
