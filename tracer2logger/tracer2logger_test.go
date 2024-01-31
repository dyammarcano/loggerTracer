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

	tp, err := SetTracer("teste")
	assert.NoError(t, err)

	tp.Info("Teste de log", zap.String("testeA", "testeB"))
	tp.Warn("Teste de log")
	tp.Error("Teste de log")

	tp, err = SetTracer("teste")
	assert.NoError(t, err)

	tp.Info("Teste de log")
	tp.Warn("Teste de log")
	tp.Error("Teste de log")
	//
	//newTracer, err := myLogger.NewTracer("teste de span log")
	//assert.NoError(t, err)
	//
	//newTracer.Start("span test 1 with trace")
	//
	//Info("Teste de log")
	//Warn("Teste de log")
	//Error("Teste de log")
	//
	//newTracer.End()
}
