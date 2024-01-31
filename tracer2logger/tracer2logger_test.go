package tracer2logger

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestNewMyLogger(t *testing.T) {
	cfg := &Config{
		LogDir:      "C:/arqprod_local/testing",
		ServiceName: "testService",
		Tracing:     true,
		//structured:  true,
	}

	err := NewMyLogger(cfg)
	assert.NoError(t, err)

	tp, err := NewTracer("teste")
	assert.NoError(t, err)

	defer tp.End()

	tp.Info("Teste de log")
	tp.Info("Teste de log", zap.String("testeA", "testeB"))
	tp.Warn("Teste de log")
	tp.Error("Teste de log")

	tp, err = NewTracer("teste")
	assert.NoError(t, err)

	tp.Info("Teste de log")
	tp.Warn("Teste de log")
	tp.Error("Teste de log")
}
