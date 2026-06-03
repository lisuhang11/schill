# 核心接口限流方案

## 1. 背景与目标

SChill 是一个面向 C 平端用户的社交内容台，其核心接口面临以下风险：

| 风险类型 | 典型场景 | 影响 |
|----------|----------|------|
| **暴力破解** | 恶意用户对 `/api/auth/login` 进行高频尝试 | 账号安全风险，数据库压力 |
| **垃圾写入** | 脚本批量发帖、评论、点赞 | 内容质量下降，Kafka 队列堆积 |
| **ES 过载** | 大量并发搜索请求 | Elasticsearch CPU 打满，影响所有搜索功能 |
| **爬虫/抓取** | 爬虫高频拉取 feed 和帖子列表 | 带宽和 CPU 浪费，影响正常用户 |
| **多副本一致性** | Gateway 水平扩容后，单机限流不生效 | 限流失效 |

因此需要在 Gateway 层（唯一 HTTP 入口）对核心接口实施**分级限流**。

## 2. 方案选型对比

### 2.1 候选方案

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| **A. go-zero PeriodLimit** | go-zero 内置的分布式限流器，基于 Redis 滑动窗口 | 框架原生支持，分布式一致，配置简单 | 依赖 Redis；需要 go-zero 的 `redis.Redis` 实例 |
| **B. go-zero TokenLimit** | go-zero 内置的令牌桶限流器 | 内存级，无外部依赖 | 仅限单实例，多副本不共享状态 |
| **C. 自研滑动窗口 (已有 `SlideWindowLimit`)** | `common/redis/redis.go` 中已有的 ZSET 滑动窗口实现 | 已有代码，Redis 依赖已满足 | 非标准组件，需要自行封装中间件；无 go-zero 集成 |
| **D. 网关层中间件 (如 Kong/APISIX)** | 在外部 API Gateway 层做限流 | 与业务代码解耦 | 引入额外组件，运维复杂度高；当前架构无外部网关 |
| **E. K8s Ingress 限流** | 使用 Nginx Ingress 的 rate-limit 注解 | 基础设施层面 | 粒度粗（只能按路径前缀），无法按 IP+用户 多维限流 |

### 2.2 最终选择：**方案 A（go-zero PeriodLimit）+ 本地 fallback**

**选择理由：**

1. **框架原生集成**：`PeriodLimit` 是 go-zero v1.10.0 内置组件，与 `rest.Server` 中间件体系无缝配合，无需引入第三方依赖。
2. **分布式一致性**：基于 Redis 的 Lua 脚本实现滑动窗口，多 Gateway 实例共享计数，水平扩展时限流准确。
3. **Redis 零依赖模式**：当 Redis 不可用时（开发环境、单实例部署），自动降级到本地内存令牌桶，保证可用性。
4. **分级策略**：按接口类型分为 auth / write / read / search 四档，精确控制不同类型请求的流量。
5. **与项目现有 Redis 客户端一致**：项目已使用 `github.com/redis/go-redis/v9`，`PeriodLimit` 依赖的 `go-zero/core/stores/redis` 可复用同一 Redis 实例。

### 2.3 与已有 `SlideWindowLimit` 的关系

`common/redis/redis.go:SlideWindowLimit` 是已有的滑动窗口限流工具，但存在以下问题：
- 没有封装为 go-zero 中间件，需要在每个 handler 中手动调用
- 不支持分布式场景下的原子操作（ZADD + ZREMRANGEBYSCORE 非原子）
- 没有 fail-open 机制（Redis 故障时直接返回错误）

本次实现不修改已有 `SlideWindowLimit`（向后兼容），而是新增 `common/ratelimit` 包统一封装。

## 3. 实现方案

### 3.1 架构图

```
                        ┌─────────────────────┐
                        │   HTTP Request       │
                        └──────────┬──────────┘
                                   │
                        ┌──────────▼──────────┐
                        │  CORS Middleware     │
                        ├─────────────────────┤
                        │  JWT Middleware      │  (已有)
                        ├─────────────────────┤
                        │  RateLimit Middleware│  (新增 ★)
                        │  ├─ Auth  20/60s     │
                        │  ├─ Write 30/60s     │
                        │  ├─ Read  200/60s    │
                        │  └─ Search 30/60s    │
                        ├─────────────────────┤
                        │  Handler             │
                        │  → gRPC → Service    │
                        └─────────────────────┘

        RateLimit Middleware
              │
    ┌─────────┴─────────┐
    │  Redis available?  │
    └─────────┬─────────┘
         Yes  │  No
    ┌─────────▼─────────┐
    │ PeriodLimit       │     │ LocalLimiter    │
    │ (分布式, Redis)    │     │ (内存令牌桶)     │
    │ Lua 脚本原子操作    │     │ sync.Mutex 锁    │
    └───────────────────┘     └─────────────────┘
```

### 3.2 代码变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `common/ratelimit/ratelimit.go` | **新增** | 限流核心包：PeriodLimit 封装、LocalLimiter、Middleware |
| `common/error/ErrorCode.go` | 修改 | 新增 `ErrRateLimitExceed = 998` |
| `service/gateway/internal/config/config.go` | 修改 | Config 新增 `RateLimit` 字段 |
| `service/gateway/etc/gateway.yaml` | 修改 | 新增 `RateLimit` 配置段 |
| `service/gateway/gateway.go` | 修改 | 注册限流中间件，路径分类逻辑 |

### 3.3 核心接口分级策略

```yaml
RateLimit:
  Auth:     # /api/auth/* (register/login/refresh)
    PeriodSec: 60    # 窗口 60 秒
    Quota: 20        # 每 IP 最多 20 次
  Write:    # POST/PUT/DELETE (发帖/评论/点赞/关注/收藏/分享)
    PeriodSec: 60
    Quota: 30        # 每 IP 最多 30 次
  Read:     # GET (feed/帖子列表/用户信息/评论列表)
    PeriodSec: 60
    Quota: 200       # 每 IP 最多 200 次
  Search:   # /api/search/* (ES 搜索)
    PeriodSec: 60
    Quota: 30        # 每 IP 最多 30 次 (比 Read 严格)
```

**设计原则：**

1. **Auth 最严格**：防止暴力破解和撞库，20 次/分钟足够正常用户注册登录。
2. **Search 比 Read 严格**：搜索直接查询 Elasticsearch，资源消耗远高于 MySQL 读，需要单独限制。
3. **Write 适中**：正常用户的写操作频率不高，30 次/分钟覆盖正常使用。
4. **Read 宽松**：浏览是核心体验，200 次/分钟 ≈ 3.3 次/秒，足够流畅。
5. **/health 不限流**：K8s liveness/readiness probe 必须永远可达。

### 3.4 限流键（Key）设计

```
限流键 = ClientIP（取自 X-Forwarded-For > X-Real-IP > RemoteAddr）

Redis Key 格式: ratelimit:{category}:{timestamp_window}
例如: ratelimit:auth:1717401600
```

**为什么用 IP 而不是 UserID？**
- Auth 接口（login/register）请求时用户尚未认证，无法获取 UserID。
- 对已认证的写操作，未来可以扩展为 `IP + UserID` 联合限流，当前 IP 维度已足够。

### 3.5 Fail-Open 策略

当 Redis 不可用或 PeriodLimit 创建失败时：
1. 自动降级到 `LocalLimiter`（内存令牌桶），保证服务可用。
2. 中间件内部 `Take()` 返回 error 时，**放行请求**（fail-open），记录错误日志。
3. 不会因为限流组件故障而拒绝所有请求。

### 3.6 响应行为

被限流时返回：
```
HTTP 429 Too Many Requests
Headers:
  X-RateLimit-Name: auth|write|read|search
  Retry-After: 1
Body: "Too Many Requests"
```

## 4. 配置说明

### 4.1 使用 Redis（推荐，生产环境）

在 `gateway.yaml` 中取消 Redis 注释：

```yaml
RateLimit:
  Redis:
    Host: 127.0.0.1:6379   # 复用项目已有 Redis
    Type: node
    Pass: ""
  Auth:
    PeriodSec: 60
    Quota: 20
  # ... 其余配置
```

### 4.2 不使用 Redis（开发/单实例）

保持 `RateLimit.Redis` 注释掉或 `Host` 为空，系统自动使用内存限流：

```yaml
RateLimit:
  # Redis 配置为空 → 自动使用本地内存令牌桶
  Auth:
    PeriodSec: 60
    Quota: 20
  # ...
```

启动日志会显示：
```
rate limit: no Redis configured, using in-memory fallback
rate limiting enabled auth=20/60s write=30/60s read=200/60s search=30/60s redis=false
```

### 4.3 调参建议

| 场景 | 建议调整 |
|------|----------|
| 测试环境 | Auth Quota 调高到 100，避免自动化测试被限 |
| 大促/活动期间 | Read Quota 可适当降低，Write Quota 保持不变 |
| 被 DDoS 攻击 | 临时降低所有 Quota，Auth 可降至 5/60s |
| 正常运营 | 使用默认值，观察监控后微调 |

## 5. 可观测性

### 5.1 日志

限流中间件会输出以下日志：

**正常启动：**
```
rate limiting enabled auth=20/60s write=30/60s read=200/60s search=30/60s redis=true
```

**Redis 故障降级：**
```
rate limit: failed to connect Redis, using in-memory fallback
```

**限流拒绝（后续可加）：** 当前被限流的请求只返回 429，不记录日志以避免日志风暴。如需监控，可在中间件中添加采样日志。

### 5.2 建议监控指标（未来扩展）

| 指标 | 含义 |
|------|------|
| `ratelimit_rejected_total{category="auth|write|read|search"}` | 各分类被拒绝的请求数 |
| `ratelimit_allowed_total{category="..."}` | 各分类放行的请求数 |
| `ratelimit_redis_errors_total` | Redis 调用失败次数 |

## 6. 测试验证

### 6.1 手动测试

```bash
# 测试 Auth 限流（20 req/60s）
for i in $(seq 1 25); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST http://localhost:8086/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}'
done
# 预期：前 20 次返回 200/400/401，后 5 次返回 429

# 测试 Search 限流（30 req/60s）
for i in $(seq 1 35); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    "http://localhost:8086/api/search/post?keyword=test"
done
# 预期：前 30 次正常，后 5 次返回 429

# 验证 /health 不限流
for i in $(seq 1 100); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    "http://localhost:8086/health"
done
# 预期：全部返回 200
```

### 6.2 单元测试位置

限流核心逻辑的单元测试可以添加在 `common/ratelimit/ratelimit_test.go` 中，覆盖：
- `LocalLimiter.Allow()` 的基本行为
- `ClientIP()` 的各种 Header 优先级
- `localAdapter.Take()` 的返回值映射

## 7. 未来扩展方向

1. **用户级限流**：对已认证用户，限流键从 IP 升级为 `UserID`，实现更精细的 per-user 控制。
2. **动态配置**：通过配置中心（如 etcd）动态调整 Quota，无需重启。
3. **分级告警**：达到 Quota 的 80% 时发送告警，提前预警异常流量。
4. **白名单**：对内部服务、管理后台的 IP 或 UserID 豁免限流。
5. **Prometheus 指标**：导出限流指标到 Prometheus，接入 Grafana 大盘。
