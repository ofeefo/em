package em

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type baseAdd[T any] interface {
	Add(context.Context, T, ...metric.AddOption)
}

type baseRecord[T any] interface {
	Record(context.Context, T, ...metric.RecordOption)
}

type add[T any] interface {
	Add(n T, opts ...metric.AddOption)
	AddCtx(ctx context.Context, n T, opts ...metric.AddOption)
}

type record[T any] interface {
	Record(n T, opts ...metric.RecordOption)
	RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption)
}

type I64Counter add[int64]

type I64UpDownCounter add[int64]

type F64Counter add[float64]

type F64UpDownCounter add[float64]

type I64Gauge record[int64]

type I64Histogram record[int64]

type F64Gauge record[float64]

type F64Histogram record[float64]

type addImpl[T any] struct {
	baseAdd[T]
	parentAttrs []attribute.KeyValue
}

type recordImpl[T any] struct {
	baseRecord[T]
	attrs []attribute.KeyValue
}

func (a *addImpl[T]) Add(n T, opts ...metric.AddOption) {
	a.AddCtx(context.Background(), n, opts...)
}

func (a *addImpl[T]) AddCtx(ctx context.Context, n T, opts ...metric.AddOption) {
	o := append([]metric.AddOption{metric.WithAttributes(a.parentAttrs...)}, opts...)
	a.baseAdd.Add(ctx, n, o...)
}

func (r *recordImpl[T]) Record(n T, opts ...metric.RecordOption) {
	r.RecordCtx(context.Background(), n, opts...)
}

func (r *recordImpl[T]) RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption) {
	o := append([]metric.RecordOption{metric.WithAttributes(r.attrs...)}, opts...)
	r.baseRecord.Record(ctx, n, o...)
}

func (p *provider) i64c(kind, id string, attrs ...attribute.KeyValue) (add[int64], error) {
	if p == nil {
		return new(nilProv[int64]), nil
	}

	var (
		base baseAdd[int64]
		err  error
	)

	switch kind {
	case counter:
		base, err = prov.m.Int64Counter(id)
	case upDownCounter:
		base, err = prov.m.Int64UpDownCounter(id)
	}
	if err != nil {
		return nil, err
	}

	return &addImpl[int64]{base, attrs}, nil
}

func (p *provider) f64c(kind, id string, attrs ...attribute.KeyValue) (add[float64], error) {
	if p == nil {
		return new(nilProv[float64]), nil
	}

	var (
		base baseAdd[float64]
		err  error
	)
	switch kind {
	case counter:
		base, err = prov.m.Float64Counter(id)
	case upDownCounter:
		base, err = prov.m.Float64UpDownCounter(id)
	}
	if err != nil {
		return nil, err
	}

	return &addImpl[float64]{base, attrs}, nil
}

func (p *provider) i64r(kind, id string, bounds []float64, attrs ...attribute.KeyValue) (record[int64], error) {
	if p == nil {
		return new(nilProv[int64]), nil
	}
	var (
		base baseRecord[int64]
		err  error
	)
	switch kind {
	case gauge:
		base, err = prov.m.Int64Gauge(id)
	case histogram:
		base, err = prov.m.Int64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
	}

	if err != nil {
		return nil, err
	}
	return &recordImpl[int64]{base, attrs}, nil
}

func (p *provider) f64r(kind, id string, bounds []float64, attrs ...attribute.KeyValue) (record[float64], error) {
	if p == nil {
		return new(nilProv[float64]), nil
	}
	var (
		base baseRecord[float64]
		err  error
	)
	switch kind {
	case gauge:
		base, err = prov.m.Float64Gauge(id)
	case histogram:
		base, err = prov.m.Float64Histogram(id, metric.WithExplicitBucketBoundaries(bounds...))
	}

	if err != nil {
		return nil, err
	}

	return &recordImpl[float64]{base, attrs}, nil
}

type nilProv[T any] struct{}

func (n2 nilProv[T]) Add(T, ...metric.AddOption) {}

func (n2 nilProv[T]) AddCtx(context.Context, T, ...metric.AddOption) {}

func (n2 nilProv[T]) Record(T, ...metric.RecordOption) {}

func (n2 nilProv[T]) RecordCtx(context.Context, T, ...metric.RecordOption) {}
