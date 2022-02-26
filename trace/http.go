package trace

import (
	"github.com/gin-gonic/gin"
)

func GinTraceHandler(ctx *gin.Context) {
	traceId := ctx.Request.Header.Get(HeaderKeyTraceId)

	if traceId == "" || !IsTraceId(traceId) {
		traceId = NewTraceId()
	}

	ctx.Set(ContextKeyTraceId, traceId)
	ctx.Header(HeaderKeyTraceId, traceId)
	ctx.Next()
}
