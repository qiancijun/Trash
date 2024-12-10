package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ageGuage = promauto.NewGauge(prometheus.GaugeOpts{
		// Namespace: "company",
		// Subsystem: "web",
		Name: "user_age",
	})
	requestTimer = promauto.NewGaugeVec(prometheus.GaugeOpts{
		// Namespace: "company",
		// Subsystem: "web",
		Name: "request_time",
	}, []string{"interface"})
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "company",
		Subsystem: "web",
		Name: "request_counter",
	}, []string{"interface"})
)

type User struct {
	Name string
	Age int
}

func getUserAge(ctx *gin.Context) {
	user := User{
		Age: rand.Intn(100),
	}
	ageGuage.Set(float64(user.Age))
	ctx.JSON(http.StatusOK, user)
}

func getUserName(ctx *gin.Context) {
	user := User{
		Name: "Cheryl",
	}
	ctx.JSON(http.StatusOK, user)
}

func timerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		begin := time.Now()
		ifc := ctx.Request.RequestURI
		requestCounter.WithLabelValues(ifc).Inc()
		ctx.Next()
		requestTimer.WithLabelValues(ifc).Set(float64(time.Since(begin).Microseconds()))
	}
}

func main() {
	engine := gin.Default()
	engine.Use(timerMiddleware())
	engine.GET("/name", getUserName)
	engine.GET("/age", getUserAge)

	engine.GET("/metrics", func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
	engine.Run("127.0.0.1:5678")
}