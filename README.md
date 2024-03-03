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
	innerFunc1(tracer, num)
	innerFunc2(tracer, num)
	innerFunc3(tracer, num)

	tracer.Error(fmt.Sprintf("Function Error %d", num), loggerTracer.AddFieldError(fmt.Errorf("error %d", num)))

	<-time.After(1 * time.Second)
}

func innerFunc1(tracer *loggerTracer.Trace4U, num int) {
	tracer.Info(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 1", String: "funcValue 1"})
}

func innerFunc2(tracer *loggerTracer.Trace4U, num int) {
	tracer.Info(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 2", String: "funcValue 2"})
}

func innerFunc3(tracer *loggerTracer.Trace4U, num int) {
	tracer.Warn(fmt.Sprintf("Inner Function %d", num), loggerTracer.Entry{Key: "funcKey 3", String: "funcValue 3"})
}
```

````text
{"level":"info","timestamp":"2024-03-02T20:29:36.580-0300","message":"Function Info 0","key 1":"value 1","traceId":"9884ab13520ce381fb222efd5a74cd7a","spanId":"36a568c359621f37","name":"lap 0"}
{"level":"info","timestamp":"2024-03-02T20:29:36.581-0300","message":"Inner Function 0","funcKey 2":"funcValue 2","traceId":"9884ab13520ce381fb222efd5a74cd7a","spanId":"2a762904928626d3","name":"lap 0"}
{"level":"info","timestamp":"2024-03-02T20:29:36.581-0300","message":"Inner Function 0","funcKey 3":"funcValue 3","traceId":"9884ab13520ce381fb222efd5a74cd7a","spanId":"8c2d417be28fc6a4","name":"lap 0"}
{"level":"info","timestamp":"2024-03-02T20:29:36.581-0300","message":"Inner Function 0","funcKey 4":"funcValue 4","traceId":"9884ab13520ce381fb222efd5a74cd7a","spanId":"ff14b77da13a1fa2","name":"lap 0"}
{"level":"info","timestamp":"2024-03-02T20:29:37.592-0300","message":"Function Info 1","key 1":"value 1","traceId":"319ea01785dc663d526d4faa7dc5f1d1","spanId":"aa1c39c906556587","name":"lap 1"}
{"level":"info","timestamp":"2024-03-02T20:29:37.592-0300","message":"Inner Function 1","funcKey 2":"funcValue 2","traceId":"319ea01785dc663d526d4faa7dc5f1d1","spanId":"acc17e71fac4e35d","name":"lap 1"}
{"level":"info","timestamp":"2024-03-02T20:29:37.592-0300","message":"Inner Function 1","funcKey 3":"funcValue 3","traceId":"319ea01785dc663d526d4faa7dc5f1d1","spanId":"b4a1577727a96ea1","name":"lap 1"}
{"level":"info","timestamp":"2024-03-02T20:29:37.592-0300","message":"Inner Function 1","funcKey 4":"funcValue 4","traceId":"319ea01785dc663d526d4faa7dc5f1d1","spanId":"9f036ebc7b722c6d","name":"lap 1"}
```
