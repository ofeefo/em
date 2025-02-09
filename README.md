# Easy metrics (em)
Shorthand for [OTEL](https://github.com/open-telemetry/opentelemetry-go) metrics
initialization.


## Usage
### TL;DR
#### [Simple usage](./example/simple_usage/main.go)
#### [Complete usage](./example/complete_usage/main.go)

### Installation:
```bash
    go get github.com/ofeefo/em
```

### Define your metrics within a struct
```go
type metrics struct {
    Counter       em.Counter[int64]         `id:"a_counter"`
    Gauge         em.Gauge[int64]           `id:"a_gauge"`
    UpDownCounter em.UpDownCounter[float64] `id:"a_updowncounter"`
    Histogram em.Histogram[float64]         `id:"a_histogram" buckets:"1.0,2.0,3.0"`
}
```

### Setup your meter provider

```go
// ...
func main() {
    // Setup receives the application identifier and optional attributes.
    // It creates a basic meter provider setup to help you get started quickly.
    // For more advanced configurations (exporters, resources, etc.), 
    // use SetupWithMeter.
    err := em.Setup("my-app", attribute.String("some", "attr"))
    if err != nil {
        // ...  
    }

    // Initialize your metrics.
    m, err :=  em.Init[metrics]()
    if err != nil {
        //...
    }

    // Record your measurements
    i.Counter.Inc(em.Attrs(attribute.String("some", "attr")))

    //  Serve your metrics.
    if err = http.ListenAndServe(":8080", promhttp.Handler()); err != nil {
        panic(err)
    }
}
```

## Features
### Supported Metrics
All available metrics are wrappers for the official [Go OTEL sdk](https://github.com/open-telemetry/opentelemetry-go),
and can be created using both `int64` and `float64`, which are the available 
types for creating metrics with the official sdk (and limited in this package 
through the `n64` interface).

### Counter
```go
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
```
### UpDownCounter
```go
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
```
### Gauge
```go
type Gauge[T n64] interface {
    // Record records the value of n to the gauge.
    Record(n T, opts ...metric.RecordOption)

    // RecordCtx records the value of n to the gauge.
    RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption)
}

```
### Histogram
```go
type Histogram[T n64] interface {
    // Record adds a value to the distribution.
    Record(n T, opts ...metric.RecordOption)

    // RecordCtx adds a value to the distribution.
    RecordCtx(ctx context.Context, n T, opts ...metric.RecordOption)

    // Measure creates a starting point using time.Now and returns a function
    // that records the time elapsed since the starting point. The returned
    // function may also receive record options to further support information
    // that may vary along a given procedure.
    Measure(transform timeSub[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption)

    // MeasureCtx creates a starting point using time.Now and returns a function
    // that records the time elapsed since the starting point. The returned
    // function may also receive record options to further support information
    // that may vary along a given procedure.
    MeasureCtx(ctx context.Context, transform timeTransform[T], opts ...metric.RecordOption) func(opts ...metric.RecordOption)
}
```

### Initialization Tags
- `id [required]`: The metric identifier.
- `buckets [optional]`: Defines bucket boundaries for histograms.
- `attrs [optional]`: Defines additional identifiers for different components
  and [subsystems](#subsystems).

### Subsystems
The following example demonstrates how to build metrics subsystems with nested 
and embedded structs:
```go
type PostgresMetrics struct{
	OpCount em.Counter[int64]    `id:"operations_performed"`
	OpTime  em.Histogram[int64]  `id:"operation_time"`
}

type dbMetrics struct {
    Master struct {
        PostgresMetrics
        WriteCounter em.Counter[int64] `id:"db_write_counter"`
    } `attrs:"sub, master"`

    PostgresMetrics `attrs:"sub, replica"`
}

type Metrics struct {
    DbMetrics dbMetrics `attrs:"sub,postgres"`
}
```

## Facilities 
- `timeSub[T]`: Functions that return the time elapsed since a given starting
  point. Already present in this package:
  - `Micros`
  - `Millis`
  - `Seconds`

### Initialization for tests
When no `Setup` functions are called, a `noop` handler is used to avoid issues
regarding metric identifier collision, allowing metrics to be initialized as usual 
without risking errors or nil pointer issues.

## Limitations
1. Only synchronous metrics are supported.
2. Initialization of instruments only allows for identifiers and 
bucket boundaries for `histograms`.






