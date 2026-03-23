package gateway

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Engine struct {
	client     *http.Client
	bufferPool sync.Pool
}

// NewEngine 创建转发引擎，配置连接池和内存池
func NewEngine() *Engine {
	return &Engine{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,              // 最大空闲连接数
				MaxIdleConnsPerHost: 10,               // 每个目标主机最大空闲连接
				IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
				DisableCompression:  true,             // 不压缩，减少 CPU 开销
			},
			Timeout: 30 * time.Second, // 请求超时时间
		},
		bufferPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 64*1024) // 64KB缓冲区
				return &buf
			},
		},
	}
}

// ForwardRequest 将客户端请求转发到后端
func (e *Engine) ForwardRequest(targetURL string, w http.ResponseWriter, r *http.Request) {
	// 构建后端请求，保留原始 body
	backendReq, err := http.NewRequest(r.Method, targetURL+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, "创建后端请求失败", http.StatusInternalServerError)
		return
	}

	// 复制请求头，并添加网关标识
	for key, values := range r.Header {
		for _, value := range values {
			backendReq.Header.Add(key, value)
		}
	}
	backendReq.Header.Set("X-Forwarded-By", "my-gateway")

	start := time.Now()

	// 发送请求（利用连接池）
	resp, err := e.client.Do(backendReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("后端请求失败: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("X-Forward-Time", elapsed.String())
	w.WriteHeader(resp.StatusCode)

	// 使用内存池缓冲区复制响应体，减少分配
	bufferPtr := e.bufferPool.Get().(*[]byte)
	buffer := *bufferPtr
	defer e.bufferPool.Put(bufferPtr)

	_, err = io.CopyBuffer(w, resp.Body, buffer)
	if err != nil {
		fmt.Printf("复制响应体失败: %v\n", err)
	}

	fmt.Printf("[转发] %s %s -> %s | 状态: %d | 耗时: %v\n",
		r.Method, r.URL.Path, targetURL, resp.StatusCode, elapsed)
}

func (e *Engine) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"engine": "v1.0",
	}
}
