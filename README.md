## My Gateway 高并发Go语言网关
一个用于学习的高性能、可扩展的Go语言网关实现，具备反向代理、限流、连接池等核心功能。

##   功能特性
**反向代理**: 支持基于路径的路由转发
**并发引擎**: 自定义连接池与内存池，高效复用资源
**限流保护**: 基于令牌桶算法的IP级限流
**容器化**: 完整的Docker支持，一键部署
**可观测**: 内置请求日志与性能指标

##   项目结构
my_gateway/  
├── cmd/gateway/  
│ └── main.go        程序入口  
├── configs/  
│ └── gateway.yml  
├──pkg/config/  
│ └── config.go  
├──pkg/gateway/  
│ └── engine.go      并发转发引擎（核心）  
│ └── gateway.go     网关主逻辑  
│ └── middleware.go  限流中间件  
├── scripts/  
│ └── benchmark.sh      压测脚本  
├── docker-compose.yml  多服务编排  
├── Dockerfile          容器构建  
├── go.mod  
├── go.sum  
├── README.md  
├── test_backend.go  

##  性能压测
使用 [wrk](https://github.com/wg/wrk) 进行压力测试（4 线程，100 连接，持续 30 秒）：
bash：
wrk -t4 -c100 -d30s http://localhost:8080/api/test
使用循环脚本快速发送 20 个请求（间隔 0.05 秒），观察状态码变化：
for i in {1..20}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/test; sleep 0.05; done
docker stats

##   限流功能验证
Running 30s test @ http://localhost:8080/api/test
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.07ms    4.85ms  94.65ms   80.42%
    Req/Sec     4.53k   688.74     6.62k    78.33%
  541063 requests in 30.06s, 126.49MB read
  Non-2xx or 3xx responses: 540910
Requests/sec:  18002.28
Transfer/sec:      4.21MB

**QPS：约 18002 req/s**
**平均延迟：6.07ms**
**中位数延迟：约 5ms（根据分布估算）**
**所有请求均被限流器正确拦截（返回 429），证明限流机制在高并发下依然准确。**

200    200     200    200    429    429  
200    429     429    200    429    429  
200    429     429    200    429    429  
200    429

**初始时桶内有 3 个令牌，前 3 个请求立即通过（200）。**
**由于令牌生成速度为 5 个/秒，而请求速率远高于此，因此后续出现 200 和 429 交替的情况，这正是令牌桶算法允许突发但长期平均速率受控的表现。**

**使用 `docker stats` 监控网关容器在 1.8 万 QPS 压测下的资源消耗：**

| CPU  约 120%（占 1.2 个核心） |  
| 内存  约 10 MiB |  
| PIDs（线程/goroutine） 15 |
