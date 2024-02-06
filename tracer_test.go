package loggerTracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTracer(t *testing.T) {
	cfg := &Config{
		LogDir:      "C:/arqprod_local/testing",
		ServiceName: "testService",
	}

	err := NewLogger(cfg)
	assert.NoError(t, err)

	tracer1 := NewTracer("testService 1")
	defer tracer1.End()

	tracer1.Info("Test Info 1", Fields{"key": "value 1"})
	tracer1.Info("Test Info 2", Fields{"key": "value 2"})

	tracer2 := NewTracer("testService 2")
	defer tracer2.End()

	tracer2.Info("Test Info 3", Fields{"key": "value 3"})
	tracer2.Info("Test Info 4", Fields{"key": "value 4"})
	tracer2.Error("Test Error 1", Fields{"key": "value 5"})
}
