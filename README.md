[![Unit Test + Lint + Security](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml/badge.svg)](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml)

# tracer2logger is a logger tracing

### `For Internal use in PoC`

### tracer2logger wrap arround [zap](https://github.com/uber-go/zap) and [opentelemetry](https://github.com/open-telemetry/opentelemetry-go)

## Installation

`go get -u github.com/dyammarcano/loggerTracing`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than other structured logging
packages and includes both structured and `printf`-style APIs.

```go
package main

import (
	"github.com/dyammarcano/loggerTracing/tracer2logger"
	"go.uber.org/zap"
)

func main() {
	cfg := &tracer2logger.Config{
		LogDir:      "/var/log/app",
		ServiceName: "testService",
	}

	if err := tracer2logger.NewMyLogger(cfg); err != nil {
		panic(err)
	}

	tp, err := tracer2logger.NewTracer("teste")
	if err != nil {
		panic(err)
	}

	defer tp.End()

	tp.Info("test info log")
	tp.Info("Teste de log", zap.String("testeA", "testeB"))
}
```
