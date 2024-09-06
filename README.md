# Easy metrics (em)
Shorthand for [OTEL](https://github.com/open-telemetry/opentelemetry-go) instrumentation initialization.
<br>
## Usage
### TL;DR
#### [Simple usage](./example/simple_usage/main.go)
#### [Complete usage](./example/complete_usage/main.go)

### Get the dependency:
```bash
    go get github.com/ofeefo/em
```

### Define your instruments in a struct
```go
type instruments struct{
    Counter64       em.I64Counter   `id:"my_counter"`
    UpDownCounter64 em.I64Counter   `id:"my_updown_counter"`
    GaugeF64        em.F64Gauge     `id:"my_gauge"`
    // Histograms can use the 'buckets' tag to define explicit boundaries.
    HistogramF64 em.F64Histogram    `id:"i_am_a_histogram" buckets:"1.0,2.0,3.0"`
}
```

### When initializing your app

```go
// ...
func main(){
    // Setup receives the application identifier and optional attributes.
    // It creates a basic OTEL setup to help you get started quickly.
    // For more advanced configurations (exporters, resources, etc.), use SetupWithMeter.
    err := em.Setup("my-app", attribute.String("some", "attr"))
    if err != nil {
        // ...  
    }

    // Initialize your instruments.
    // There's also a MustInit function, which will panic if initialization fails.
    i, err :=  em.Init[instruments]()
    if err != nil {
        //...
    }

    // Record your measurements
    i.Counter64.Add(1, em.Attrs(attribute.String("some", "attr")))

    //  Serve your metrics.
    if err = http.ListenAndServe(":8080", promhttp.Handler()); err != nil {
        panic(err)
    }
}
```

## Features
### Supported tags
#### Instruments
* `id [required]`: The instrument identifier.
* `buckets [optional]`: Defines bucket boundaries for histograms.

#### Nested or Embedded structs:
  * `attrs [optional]`: Comma-separated string attributes to identify specific instruments sets.

### Supported instruments [int64/float64]:
* `Counter`
* `UpDownCounter`
* `Gauge`
* `Histogram`

### Nested & Embedded structs
The following example demonstrates how nested and embedded structs are supported:

```go
type Embedded struct {
    Histogram     em.I64Histogram     `id:"example_embedded_histogram"`
    UpDownCounter em.F64UpDownCounter `id:"example_embedded_updowncounter"`
}

type nested struct {
    Counter  em.F64Counter `id:"example_nested_counter"`
    Gauge    em.F64Gauge   `id:"example_nested_gauge"`
    MoreNest struct {
        Counter em.F64Counter `id:"example_more_nested_counter"`
    }
}

type samplers struct {
    // Nested and embedded structs (and struct pointers) are supported.
    // Instruments in these structs inherit attributes both from the parent struct
    // initialization and through the 'attrs' tag.
    *Embedded `attrs:"sub,embedded"`
    Nested    nested `attrs:"sub,nested"`
}
```







