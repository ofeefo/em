package em

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Attrs(attrs ...attribute.KeyValue) metric.MeasurementOption {
	return metric.WithAttributes(attrs...)
}

// Millis returns the time elapsed since a given start point in milliseconds.
func Millis[T n64](start time.Time) T {
	return T(time.Since(start).Milliseconds())
}

// Micros returns the time elapsed since a given start point in microseconds.
func Micros[T n64](start time.Time) T {
	return T(time.Since(start).Milliseconds())
}

// Seconds returns the time elapsed since a given start point in seconds.
func Seconds[T n64](start time.Time) T {
	return T(time.Since(start).Seconds())
}
