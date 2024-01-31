[![CI Build + Unit Test](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml/badge.svg)](https://github.com/dyammarcano/loggerTracing/actions/workflows/ci.yml)

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
cfg := &Config{
		logDir:      "/var/log/app",
		serviceName: "testService",
		tracing:     true,
		structured:  true,
	}

if err := NewMyLogger(cfg); err != nil {
  panic(err)
}

tp, _ := tracer2logger.SetTracer("teste")
defer func() { tracer2logger.Shutdown() }()

tp.Info("test info log")
```
