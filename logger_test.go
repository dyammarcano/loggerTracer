package loggerTracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLogger(t *testing.T) {
	cfg := &Config{
		LogDir:      "C:/dev/testing",
		ServiceName: "testService",
	}

	err := NewLogger(cfg)
	assert.NoError(t, err)

	Info("Test Info witout fields")
	Info("Test Info witout fields", AddField("hello", "world"))
	Error("Test Error witout fields")
}
