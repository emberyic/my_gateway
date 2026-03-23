package gateway

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter  *rate.Limiter
	lastseen time.Time
}

type RateLimitMiddleware struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	rate     rate.Limit
	bucket   int
}

// NewRateLimitMiddleware 创建限流中间件，r 为每秒令牌数，b 为桶容量
func NewRateLimitMiddleware(r float64, b int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*RateLimiter),
		rate:     rate.Limit(r),
		bucket:   b,
	}
}

// 定期清理长时间未访问的 IP 限流器，防止内存泄漏
func (m *RateLimitMiddleware) cleanupLimiters() {
	for {
		time.Sleep(10 * time.Minute)
		m.mu.Lock()
		for ip, limiter := range m.limiters {
			if time.Since(limiter.lastseen) > 15*time.Minute {
				delete(m.limiters, ip)
			}
		}
		m.mu.Unlock()
	}
}

// 获取或创建指定 IP 的限流器
func (m *RateLimitMiddleware) getLimiter(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[ip]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(m.rate, m.bucket),
			lastseen: time.Now(),
		}
		m.limiters[ip] = limiter
	} else {
		limiter.lastseen = time.Now()
	}
	return limiter.limiter
}

// Middleware 返回限流包装函数
func (m *RateLimitMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	go m.cleanupLimiters()
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端 IP
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		limiter := m.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", strconv.FormatFloat(float64(m.rate), 'f', 1, 64)+"次/秒")
			http.Error(w, `{"error":"请求过于频繁，请稍后再试"}`, http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", "正常")
		next(w, r)
	}
}
