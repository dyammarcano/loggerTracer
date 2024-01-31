[![Unit Test + Lint + Security](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml/badge.svg)](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml)

# tracer2logger is a logger tracing

## tracer2logger wrap arround [zap](https://github.com/uber-go/zap)

## Installation

`go get -u github.com/dyammarcano/loggerTracing`

Note that zap only supports the two most recent minor versions of Go.

## Quick Start

In contexts where performance is nice, but not critical, use the
`SugaredLogger`. It's 4-10x faster than other structured logging
packages and includes both structured and `printf`-style APIs.

```go
package main

import "github.com/dyammarcano/loggerTracing/tracer2logger"

func main() {
	cfg := &tracer2logger.Config{
		LogDir:      "/var/log/app",
		ServiceName: "testService",
		Tracing:     true,
		Structured:  true,
	}

	if err := tracer2logger.NewMyLogger(cfg); err != nil {
		panic(err)
	}

	tp, err := tracer2logger.SetTracer("teste")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := tp.Shutdown(); err != nil {
			panic(err)
		}
	}()

	tp.Info("test info log")
}
```
