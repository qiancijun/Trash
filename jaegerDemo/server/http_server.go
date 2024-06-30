package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerdemo "github.com/qiancijun/Trash/jaegerDemo"
)

var (
	closer io.Closer
)

func InitHttpJaeger() {
	var jaeger opentracing.Tracer
	var err error
	jaeger, closer, err = jaegerdemo.NewJaegerTracer("my_http_server", "127.0.0.1:6831")
	if err != nil {
		panic(err)
	}
	// 这里不需要 close，等到 server 被 kill 时才 close
	opentracing.SetGlobalTracer(jaeger)
}

func greet(ctx *gin.Context) {
	time.Sleep(100 * time.Millisecond)
	ctx.String(http.StatusOK, "hello")
}

// 向 jaeger 上报数据
func ServerTraceMiddleware(ctx *gin.Context) {
	//拿着Uber-Trace-Id去查询jaeger数据库，查询结果反序列化为一个SpanContext
	clientSpanCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(ctx.Request.Header))
	if err != nil {
		log.Printf("反序列化request header失败: %s", err)
	}

	operationName := ctx.Request.RequestURI
	// 创建 Server 端 Span
	serverSpan := opentracing.StartSpan(
		operationName,
		ext.RPCServerOption(clientSpanCtx)) //把client的metadata带进来。一方面指明继承关系，另一方面指明这是一个Server端的span(Tag里有一条span.kind=server)
	defer serverSpan.Finish() //将SpanContext写入jaeger数据库
	for k, v := range ctx.Request.Header {
		if k == "Uber-Trace-Id" {
			continue
		}
		serverSpan.SetTag(k, v[0])
	}

	ctx.Next()
}

func main() {
	InitHttpJaeger()
	router := gin.Default()
	router.Use(ServerTraceMiddleware)
	router.GET("/greet", greet)
	router.Run("localhost:5678")
}