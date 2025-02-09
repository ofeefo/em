package em

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"reflect"
	"time"
)

type Histogram[T n64] interface {
	// Record adds a value to the distribution.
	Record(n T, opts ...metric.RecordOption)

	// RecordCtx adds a value to the distribution.
	RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption)

	// Measure creates a starting point using time.Now and returns a function
	// that records the time elapsed since the starting point. The returned
	// function may also receive record options to further support information
	// that may vary along a given procedure.
	Measure(transform timeTransform[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption)

	// MeasureCtx creates a starting point using time.Now and returns a function
	// that records the time elapsed since the starting point. The returned
	// function may also receive record options to further support information
	// that may vary along a given procedure.
	MeasureCtx(ctx context.Context, transform timeTransform[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption)
}

type histogram[T n64] struct {
	attrs []attribute.KeyValue
	baseRecord[T]
}

func (h *histogram[T]) Record(n T, opts ...metric.RecordOption) {
	h.RecordCtx(context.Background(), n, opts...)
}

func (h *histogram[T]) RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption) {
	h.baseRecord.Record(ctx, n, append(opts, metric.WithAttributes(h.attrs...))...)
}

func (h *histogram[T]) Measure(transform timeTransform[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption) {
	return h.MeasureCtx(context.Background(), transform, opts...)
}

func (h *histogram[T]) MeasureCtx(ctx context.Context, transform timeTransform[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption) {
	now := time.Now()
	baseOpts := opts
	return func(opts ...metric.RecordOption) {
		h.RecordCtx(ctx, transform(now), append(baseOpts, opts...)...)
	}
}

func newHistogram[T n64]() buildable { return new(histogram[T]) }

func (h *histogram[T]) init(field reflect.StructField, attrs ...attribute.KeyValue) error {
	if p == nil {
		h.baseRecord = dummy[T]{}
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

	bounds, err := getBounds(field)
	if err != nil {
		return err
	}

	if useFloat[T]() {
		base, err = p.m.Float64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
	} else {
		base, err = p.m.Int64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
	}

	if err != nil {
		return err
	}

	h.baseRecord = base.(baseRecord[T])
	h.attrs = attrs
	return nil
}
