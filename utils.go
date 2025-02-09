package em

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Attrs(attrs ...attribute.KeyValue) metric.MeasurementOption {
	return metric.WithAttributes(attrs...)
}

func Millis[T n64](start time.Time) T {
	return T(time.Since(start).Milliseconds())
}

func Micros[T n64](start time.Time) T {
	return T(time.Since(start).Milliseconds())
}

func Seconds[T n64](start time.Time) T {
	return T(time.Since(start).Seconds())
}
