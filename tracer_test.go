package loggerTracer

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTracer(t *testing.T) {
	cfg := &Config{
		LogDir:      "C:/dev/testing",
		ServiceName: "testService",
	}

	err := NewLogger(cfg)
	assert.NoError(t, err)

	tracer1 := NewTracer("testService 1")
	defer tracer1.End()

	tracer1.Info("Test Info 1", AddField("key 1", "value 1"), Entry{Key: "key 2", String: "value 2"})
	tracer1.Info("Test Info 2", Entry{Key: "key 2", String: "value 2"})

	tracer2 := NewTracer("testService 2")
	defer tracer2.End()

	fields := Entry{Key: "key 1", String: "value 1"}

	tracer2.Info("Test Info 3", fields)
	tracer2.Info("Test Info 4", AddField("key 2", "value 2"), AddField("key 3", "value 3"))
	tracer2.Info("Test Info format", AddFieldFormat("key 1", "%d + %d = %d", 1, 2, 3))
	tracer2.Error("Test Error 1", AddFieldError(err))
	tracer2.Error("Test Error 2", AddFieldError(errors.New("test error")))
	tracer2.Info("Test Info no fields")
}
