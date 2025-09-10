package em

import (
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	m2 "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type provider struct {
	m     metric.Meter
	ready bool
	mu    sync.Mutex
}

var p = &provider{
	m: noop.Meter{},
}

func MustSetup(name string, attrs ...attribute.KeyValue) {
	if err := Setup(name, attrs...); err != nil {
		panic(err)
	}
}

func SetupWithMeter(meter metric.Meter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if meter == nil {
		panic("nil meter provided to SetupWithMeter")
	}
	p = &provider{m: meter, ready: true}
}

func Setup(name string, attrs ...attribute.KeyValue) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ready {
		return nil
	}

	promEx, err := prometheus.New()
	if err != nil {
		return err
	}

	res := resource.NewWithAttributes(semconv.SchemaURL, attrs...)
	exp := m2.NewMeterProvider(m2.WithReader(promEx), m2.WithResource(res))
	p.m = exp.Meter(name)
	p.ready = true
	return nil
}
