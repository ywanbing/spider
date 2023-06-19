package middleware

import (
	"github.com/ywanbing/spider"
	"github.com/ywanbing/spider/message"
	"github.com/ywanbing/spider/xtrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// WithClientTracing 注册客户端链路追踪中间件
func WithClientTracing(c *spider.Context) {
	ctx, span := otel.Tracer(xtrace.OTEL_TRACER_NAME).Start(c.GetCtx(), "spider_client_request",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.Int64("msg_req_id", int64(c.GetReqMsgId())),
			attribute.String("seq", c.GetReqMsg().GetHeader()[message.MsgSeq]),
		),
	)

	// 注入链路到上下文中
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(c.GetReqMsg().GetHeader()))
	defer span.End()

	c.SetCtx(ctx)

	c.Next()
}

// WithServerTracing 注册服务端链路追踪中间件
func WithServerTracing(c *spider.Context) {
	// 从请求头中提取链路
	ctx := otel.GetTextMapPropagator().Extract(c.GetCtx(), propagation.MapCarrier(c.GetReqMsg().GetHeader()))

	ctx, span := otel.Tracer(xtrace.OTEL_TRACER_NAME).Start(ctx, "spider_server_request", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	c.SetCtx(ctx)

	c.Next()
}
