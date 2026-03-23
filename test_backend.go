package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// 模拟后端处理耗时
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("X-Backend", "test-server")
		fmt.Fprintf(w, "后端响应: %s | 时间: %v", r.URL.Path, time.Now().Format("15:04:05.000"))
	})

	fmt.Println("测试后端运行在: http://localhost:8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("后端服务器启动失败: %v\n", err)
	}
}
