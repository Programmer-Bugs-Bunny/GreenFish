package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	"go.uber.org/zap"
)

var zipkinTracer *zipkin.Tracer

// SetZipkinTracer 设置全局Zipkin tracer
func SetZipkinTracer(tracer *zipkin.Tracer) {
	zipkinTracer = tracer
}

// GetZipkinTracer 获取全局Zipkin tracer
func GetZipkinTracer() *zipkin.Tracer {
	return zipkinTracer
}

// TracingMiddleware 链路追踪中间件
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if zipkinTracer == nil {
			c.Next()
			return
		}

		// 创建span名称
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())

		// 创建span
		span := zipkinTracer.StartSpan(spanName)
		defer span.Finish()

		// 设置span标签
		span.Tag("http.method", c.Request.Method)
		span.Tag("http.url", c.Request.URL.String())
		span.Tag("http.path", c.Request.URL.Path)
		span.Tag("user_agent", c.GetHeader("User-Agent"))
		span.Tag("client_ip", c.ClientIP())

		// 将span上下文传递给后续处理
		ctx := zipkin.NewContext(c.Request.Context(), span)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 设置响应状态码
		span.Tag("http.status_code", fmt.Sprintf("%d", c.Writer.Status()))

		// 如果是错误状态码，记录错误
		if c.Writer.Status() >= 400 {
			span.Tag("error", "true")
			if len(c.Errors) > 0 {
				span.Tag("error.message", c.Errors.String())
			}
		}

		Logger.Debug("链路追踪记录完成",
			zap.String("span_name", spanName),
			zap.String("trace_id", span.Context().TraceID.String()),
			zap.Int("status_code", c.Writer.Status()),
		)
	}
}

// StartSpan 在指定上下文中开始一个新的span
func StartSpan(ctx context.Context, operationName string) (context.Context, zipkin.Span) {
	if zipkinTracer == nil {
		return ctx, nil
	}

	span := zipkinTracer.StartSpan(operationName, zipkin.Parent(zipkin.SpanFromContext(ctx).Context()))
	newCtx := zipkin.NewContext(ctx, span)

	return newCtx, span
}

// FinishSpan 完成span并记录可选的错误信息
func FinishSpan(span zipkin.Span, err error) {
	if span == nil {
		return
	}

	if err != nil {
		span.Tag("error", "true")
		span.Tag("error.message", err.Error())
	}

	span.Finish()
}
