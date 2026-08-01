# trace-it

`trace-it` is a lightweight HTTP transport that measures the lifecycle of outgoing HTTP requests using `net/http/httptrace`.

It collects timing information such as:

- DNS lookup
- TLS handshake
- Time to first response byte
- Total request duration

The transport can also report metadata about the completed request (URL, method, status code) through a callback.

## Installation

```bash
go get github.com/DionisiyGri/trace-it
```

## Basic Usage

```go
package main

import (
    "fmt"
    "io"
    "net/http"

    "github.com/DionisiyGri/tracer"
)

func main() {
    client := &http.Client{
        Transport: &tracer.TracingTransport{
            OnResult: func(r tracer.TraceResult) {
                fmt.Printf("%+v\n", r)
            },
        },
    }

    resp, err := client.Get("https://google.com")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    _, _ = io.ReadAll(resp.Body)
}
```

## Using a Custom Transport

If no transport is provided, `http.DefaultTransport` is used.

To customize the underlying transport:

```go
transport := &http.Transport{
    MaxIdleConns:       100,
    MaxIdleConnsPerHost: 10,
}

client := &http.Client{
    Transport: &tracer.TracingTransport{
        RoundTripper: transport,
    },
}
```

## Receiving Results

Every completed request invokes the `OnResult` callback.

```go
client := &http.Client{
    Transport: &tracer.TracingTransport{
        OnResult: func(r tracer.TraceResult) {
            fmt.Printf(
                "%s %s -> %d (%v)\n",
                r.Method,
                r.URL,
                r.StatusCode,
                r.Timings.Total,
            )
        },
    },
}
```


## Notes

- The transport is fully compatible with `net/http`.
- If `RoundTripper` is nil, `http.DefaultTransport` is used.
- Timing information is finalized when the response body is closed.
- To ensure accurate timing measurements, always close the response body:

```go
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

or consume it completely:

```go
_, _ = io.Copy(io.Discard, resp.Body)
resp.Body.Close()
```