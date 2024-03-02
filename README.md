[![Unit Test + Lint + Security](https://github.com/dyammarcano/loggerTracer/actions/workflows/ci.yml/badge.svg)](https://github.com/dyammarcano/loggerTracer/actions/workflows/ci.yml)

# loggerTracer is a logger tracing

### `For Internal use in PoC`

### loggerTracer wrap arround [zap](https://github.com/uber-go/zap) and [opentelemetry](https://github.com/open-telemetry/opentelemetry-go)

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
	"fmt"
	"github.com/dyammarcano/loggerTracer"
	"time"
)

func main() {
	cfg := &loggerTracer.Config{
		LogDir:      "/var/log/app",
		ServiceName: "testService",
	}

	if err := loggerTracer.NewLogger(cfg); err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		mainFunc(i)
	}
}

func mainFunc(num int) {
	tracer := loggerTracer.NewTracer(fmt.Sprintf("lap %d", num))
	defer tracer.End()

	tracer.Info(fmt.Sprintf("Function Info %d", num), loggerTracer.AddField("key 1", "value 1"))
	innerFunc(tracer, num)
	innerFunc2(tracer, num)
	innerFunc3(tracer, num)

	tracer.Error(fmt.Sprintf("Function Error %d", num), loggerTracer.AddFieldError(fmt.Errorf("error %d", num)))

	<-time.After(1 * time.Second)
}

func innerFunc(tracer *loggerTracer.Trace4U, num int) {
	tracer.Info(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 1", String: "funcValue 1"})
}

func innerFunc2(tracer *loggerTracer.Trace4U, num int) {
	tracer.Info(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 2", String: "funcValue 2"})
}

func innerFunc3(tracer *loggerTracer.Trace4U, num int) {
	tracer.Warn(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 3", String: "funcValue 3"})
}
```
