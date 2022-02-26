package trace

import (
	"context"
	"github.com/google/uuid"
)

const (
	ContextKeyTraceId = "trace.traceid"
	HeaderKeyTraceId  = "X-Trace-ID"
)

func NewTraceId() string {
	return uuid.New().String()
}

func IsTraceId(traceId string) bool {
	_, err := uuid.Parse(traceId)
	return err == nil
}

func GetTraceId(ctx context.Context) (traceId string) {
	if v, ok := ctx.Value(ContextKeyTraceId).(string); ok && v != "" && IsTraceId(v) {
		return v
	}

	return NewTraceId()
}

func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, ContextKeyTraceId, traceId)
}
