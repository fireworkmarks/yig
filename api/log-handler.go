package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/journeymidnight/yig/meta"
)

type logHandler struct {
	handler http.Handler
}

func (l logHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Serves the request.
	start := time.Now().Nanosecond()
	logger := ContextLogger(r)
	end := time.Now().Nanosecond()
	wasteTime := end - start
	fmt.Println("消耗的时间为：", wasteTime)
	start = time.Now().Nanosecond()
	logger.Info("Start serving", r.Method, r.Host, r.URL)
	end = time.Now().Nanosecond()
	wasteTime = end - start
	fmt.Println("第一个info消耗的时间为：", wasteTime)
	l.handler.ServeHTTP(w, r)
	start = time.Now().Nanosecond()
	logger.Info("Completed", r.Method, r.Host, r.URL)
	end = time.Now().Nanosecond()
	wasteTime = end - start
	fmt.Println("第二个info消耗的时间为：", wasteTime)
}

func SetLogHandler(h http.Handler, _ *meta.Meta) http.Handler {
	return logHandler{handler: h}
}
