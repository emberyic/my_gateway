package main

import (
	"log"
	"net/http"
	"os"

	"github.com/emberyic/my_gateway/pkg/gateway"
)

func main() {
	engine := gateway.NewEngine()

	// 限流：每秒5个请求，桶容量3
	rateLimiter := gateway.NewRateLimitMiddleware(5.0, 3)

	// 从环境变量获取后端地址（Docker Compose 中设置）
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8081"
		log.Println("未设置 BACKEND_URL 环境变量，使用默认值:", backendURL)
	} else {
		log.Println("从环境变量读取 BACKEND_URL:", backendURL)
	}

	// 核心转发逻辑：/api/* 的请求转发到后端，其他返回欢迎信息
	forwardHandler := func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[0:4] == "/api" {
			engine.ForwardRequest(backendURL, w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("My Gateway v1.0\n支持功能:反向代理、限流、并发连接池"))
	}

	// 包装限流中间件
	finalHandler := rateLimiter.Middleware(forwardHandler)

	http.HandleFunc("/", finalHandler)

	log.Println("网关启动: http://0.0.0.0:8080")
	log.Printf("转发规则: /api/* -> %s", backendURL)
	log.Printf("限流配置: %.1f req/s, 突发 %d", 5.0, 3)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("启动失败: ", err)
	}
}
