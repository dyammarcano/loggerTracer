package loggerTracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// getFields returns the fields with the context.
func getFields(ctx context.Context, fields ...Entry) []zap.Field {
	var zapFields []zap.Field

	if fields == nil {
		return zapFields
	}

	for i := range fields {
		if fields[i].String != "" {
			zapFields = append(zapFields, zap.Any(fields[i].Key, fields[i].String))
		}

		if fields[i].Integer != 0 {
			zapFields = append(zapFields, zap.Any(fields[i].Key, fields[i].Integer))
		}

		if fields[i].Interface != nil {
			zapFields = append(zapFields, zap.Any(fields[i].Key, fields[i].Interface))
		}
	}

	// get val.name from context
	if ctx != nil {
		name, ok := ctx.Value("serviceName").(string)
		if ok {
			_, newSpan := tracerProvider.Tracer("logEntry").Start(ctx, "logEntry")

			zapFields = append(zapFields, zap.String("traceId", trace.SpanContextFromContext(ctx).TraceID().String()))
			zapFields = append(zapFields, zap.String("spanId", newSpan.SpanContext().SpanID().String()))
			zapFields = append(zapFields, zap.String("name", name))

			return zapFields
		}
	}

	return zapFields
}

// AddField returns a new Entry.
func AddField(key string, value any) Entry {
	return Entry{Key: key, Interface: value}
}

func AddFieldFormat(key string, format string, a ...any) Entry {
	return Entry{Key: key, String: fmt.Sprintf(format, a...)}
}

// AddFieldError returns a new Entry with the error.
func AddFieldError(err error) Entry {
	if err != nil {
		return Entry{Key: "err", String: err.Error()}
	}

	return Entry{}
}
