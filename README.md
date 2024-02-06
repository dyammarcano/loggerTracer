[![Unit Test + Lint + Security](https://github.com/dyammarcano/loggerTracer/actions/workflows/ci.yml/badge.svg)](https://github.com/dyammarcano/loggerTracer/actions/workflows/ci.yml)

# loggerTracer is a logger tracing

### `For Internal use in PoC`

### tracer2logger wrap arround [zap](https://github.com/uber-go/zap) and [opentelemetry](https://github.com/open-telemetry/opentelemetry-go)

## Installation

`go get -u github.com/dyammarcano/loggerTracer`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than other structured logging
packages and includes both structured and `printf`-style APIs.

```go
package main

import (
	"github.com/dyammarcano/loggerTracer"
)

func main() {
	cfg := &loggerTracer.Config{
		LogDir:      "/var/log/app",
		ServiceName: "testService",
	}

	if err := loggerTracer.NewMyLogger(cfg); err != nil {
		panic(err)
	}

	tracer1 := loggerTracer.NewTracer("testService 1")
	defer tracer1.End()

	tracer1.Info("Test Info 1", loggerTracer.Fields{"key": "value 1"})
	tracer1.Info("Test Info 2", loggerTracer.Fields{"key": "value 2"})
}
```
