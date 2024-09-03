package em

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	m2 "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type provider struct {
	m metric.Meter
}

var prov *provider = nil

func SetupWithMeter(meter metric.Meter) {
	if meter == nil {
		return
	}
	prov = &provider{m: meter}
}

func Setup(name string, attrs ...attribute.KeyValue) error {
	if prov != nil {
		return nil
	}

	promEx, err := prometheus.New()
	if err != nil {
		return err
	}

	res := resource.NewWithAttributes(semconv.SchemaURL, attrs...)
	exp := m2.NewMeterProvider(m2.WithReader(promEx), m2.WithResource(res))
	prov = &provider{exp.Meter(name)}
	return nil
}
