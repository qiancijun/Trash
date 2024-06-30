package main

import (
	"context"
	"time"

	jaegerdemo "github.com/qiancijun/Trash/jaegerDemo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func main() {
	jaegerdemo.NewJaegerTracer2()
	tracer = otel.Tracer("visit")
	visitWebSite(context.Background(), 1, "cheryl")
}

// 用户访问网站，函数入口
func visitWebSite(ctx context.Context, userId int, userName string) string {
	ctx, span := tracer.Start(ctx, "visit_website")
	defer span.End()

	time.Sleep(1 * time.Millisecond)
	span.SetName("visiv_website")

	recordUV(span, userId)
	return "hello"
}

// 上报用户来访，用于后续统计UV(user visit)
func recordUV(parentSpan trace.Span, userId int) {
	// record_uv  follows from visit_website。FollowsFrom表示parent span不以任何形式依赖child span的结果，当然child span的工作也是由parent span引起的
	// span := opentracing.GlobalTracer().StartSpan("record_uv", opentracing.FollowsFrom(parentSpan.Context()))
	_, span := tracer.Start(context.Background(), "record_uv")
	defer span.End()

	time.Sleep(3 * time.Millisecond)
}