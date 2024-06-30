package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	jaegerdemo "github.com/qiancijun/Trash/jaegerDemo"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")


func main() {
	jaeger, closer, err := jaegerdemo.NewJaegerTracer("my_service", "127.0.0.1:6831")
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	// 设置全局 tracer，避免传递 trace 导致额外开销
	opentracing.SetGlobalTracer(jaeger)
	
	ctx := context.Background()
	userId := 8
	userName := "cheryl"
	content := visitWebSite(ctx, userId, userName)
	fmt.Println(content)
}

// 用户访问网站，函数入口
func visitWebSite(ctx context.Context, userId int, userName string) string {
	span := opentracing.GlobalTracer().StartSpan("visit_website")
	defer span.Finish() // 只有调用 Finish() 才会写入 jaeger 数据库

	time.Sleep(1 * time.Millisecond)

	span.SetTag("访问时段", "上午") // Tag 侧重于查询
	// Log 记录 event 发生的时间
	span.LogFields(
		log.Int("user_visit", userId),
		log.String("visit_page", "/home"),
	)

	// BaggageItem 可看成是一种特殊的 Log(event=baggage)，它里面的数据可以传递给子孙后代，普通Log和Tag传不给后代
	span.SetBaggageItem("trace_id",RandStringRunes(10))
	span.SetBaggageItem("user_id", strconv.Itoa(userId))

	recordUV(span, userId)
	// go recordUV(span, userId)
	reccommend := getReccommend(span, userId)
	return reccommend
}

// 上报用户来访，用于后续统计UV(user visit)
func recordUV(parentSpan opentracing.Span, userId int) {
	// record_uv  follows from visit_website。FollowsFrom表示parent span不以任何形式依赖child span的结果，当然child span的工作也是由parent span引起的
	span := opentracing.GlobalTracer().StartSpan("record_uv", opentracing.FollowsFrom(parentSpan.Context()))
	defer span.Finish()

	time.Sleep(3 * time.Millisecond)
}

// 调推荐的微服务，获取推荐列表
func getReccommend(parentSpan opentracing.Span, userId int) string {
	// ChildOf表示parent span某种程度上依赖child span的结果
	span := opentracing.GlobalTracer().StartSpan("reccommend", opentracing.ChildOf(parentSpan.Context()))
	defer span.Finish()

	userRole := getUserRole(span, userId)
	list := make([]string, 0, 10)
	if "vip" != strings.ToLower(userRole) {
		list = append(list, "广告视频")
	}
	list = append(list, "gorm教程")
	list = append(list, "grpc教程")
	return strings.Join(list, "\n")
}

// 从MySQL里获取用户的角色
func getUserRole(parentSpan opentracing.Span, userId int) string {
	// ChildOf表示parent span某种程度上依赖child span的结果
	span := opentracing.GlobalTracer().StartSpan("get_user_role", opentracing.ChildOf(parentSpan.Context()))
	defer span.Finish()

	// visit_website --> reccommend --> get_user_role。这里的BaggageItem是从它爷爷那儿继承过来的
	fmt.Println("打印BaggageItem")
	span.Context().ForeachBaggageItem(func(k, v string) bool {
		fmt.Println(k, v)
		return true
	})
	fmt.Println()

	return "VIP"
}

// 生成随机字符串
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
