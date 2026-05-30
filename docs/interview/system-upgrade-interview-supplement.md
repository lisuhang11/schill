# 社交内容系统架构升级面试题 - 补充部分

## 维度六：消费者生命周期与优雅上下线（续）

### 问题6.1（续）
发布时的频繁重平衡会导致消费卡顿，你如何优化发布过程，减少对业务的影响？

**考察点：发布策略、Kafka调优、流量控制**

#### 参考答案：

**发布优化方案：**

**一、灰度发布策略**

```yaml
# Kubernetes滚动更新配置
apiVersion: apps/v1
kind: Deployment
metadata:
  name: consumer-deployment
spec:
  strategy:
    type: RollingUpdate
    maxSurge: 1
    maxUnavailable: 0
  minReadySeconds: 30
```

**二、Kafka消费者组优化**

```go
// 配置优化
func OptimizedKafkaConfig() *sarama.Config {
    cfg := sarama.NewConfig()
    
    // 会话超时延长
    cfg.Consumer.Group.Session.Timeout = 60 * time.Second
    cfg.Consumer.Group.Heartbeat.Interval = 20 * time.Second
    
    // 重平衡配置
    cfg.Consumer.Group.Rebalance.Timeout = 120 * time.Second
    cfg.Consumer.Group.Rebalance.Retry.Max = 5
    cfg.Consumer.Group.Rebalance.Retry.Backoff = 2 * time.Second
    
    // 静态成员ID（避免频繁重平衡）
    cfg.Consumer.Group.InstanceId = "consumer-1" // 每个实例不同
    
    return cfg
}
```

**三、延迟注册模式**

```go
func (c *Consumer) Start() error {
    // 先不立即注册到服务发现
    c.healthy.Store(false)
    
    // 启动消费
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        c.client.Consume(c.ctx, c.topics, c.handler)
    }()
    
    // 等待稳定后再注册
    time.Sleep(30 * time.Second)
    c.healthy.Store(true)
    
    return nil
}
```

---

## 维度七：安全基线

### 问题7.3
在微服务架构下，你如何设计统一的权限控制体系？权限校验是放在API网关层，还是每个服务内部自己做？为什么？

**考察点：权限架构设计、分层架构理解**

#### 参考答案：

**统一权限控制架构：**

**一、分层权限设计**

```
┌─────────────────────────────────────────────────────┐
│              API 网关层                             │
│  - JWT Token验证、Token刷新                        │
│  - 基本身份认证                                │
│  - IP白名单                                │
└─────────────────┬───────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────┐
│              服务内部层                             │
│  - 细粒度权限验证                            │
│  - 资源访问控制                             │
│  - 数据权限隔离                             │
└─────────────────────────────────────────────────────┘
```

**二、JWT验证中间件**

```go
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            claims, err := jwt.ParseWithClaims(token, &UserClaims{}, 
                func(token *jwt.Token) (interface{}, error) {
                    return []byte(secret), nil
                })
            
            if err != nil || !claims.Valid {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            
            // 将用户信息放入context
            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**三、服务内部细粒度权限**

```go
func CheckPermission(ctx context.Context, userID uint64, resource string, action string) bool {
    // 从用户服务获取用户权限
    permissions, err := userRpc.GetUserPermissions(ctx, userID)
    
    // 检查权限
    return hasPermission(permissions, resource, action)
}

func (s *PostService) DeletePost(ctx context.Context, postID uint64) error {
    userID := authCtx.GetUserID(ctx)
    
    // 1. 获取帖子
    post, err := s.db.GetPost(postID)
    
    // 2. 检查权限：自己的帖子或管理员
    if post.UserID != userID && !CheckPermission(ctx, userID, "post", "delete") {
        return errors.New("permission denied")
    }
    
    return s.db.DeletePost(postID)
}
```

---

## 维度八：可观测性与回归测试策略

### 问题8.3
假设现在要做一次大规模的架构变更（比如拆分服务），你会设计什么样的发布流程和验证机制，来确保变更安全？

**考察点：变更管理能力、风险控制思维**

#### 参考答案：

**大规模变更发布流程：**

**一、发布前准备**

```go
// 1. 风险评估
type RiskAssessment struct {
    ImpactLevel string // "High" | "Medium" | "Low"
    RollbackPlan string
    MonitoringMetrics []string
}

// 2. 自动化回归测试
func RunRegressionTests() bool {
    // 核心业务链路测试
    tests := []struct{
        Name string
        TestFunc func() error
    }{
        {"LoginTest",
        {"CreatePostTest",
        {"CommentTest",
        {"LikeTest",
    }
    
    for _, t := range tests {
        if err := t.TestFunc(); err != nil {
            return false
        }
    }
    return true
}
```

**二、灰度发布策略**

```
┌─────────────────────────────────────────────┐
│           发布阶段                          │
│                                         │
│  阶段1: 内部测试环境 → 内部用户100%    │
│  阶段2: 灰度1% → 10% → 50%              │
│  阶段3: 全量发布                       │
│  阶段4: 监控72小时                     │
└─────────────────────────────────────────────┘
```

**三、流量镜像/双写验证**

```go
// 双写模式
func MigratedCreatePost(ctx context.Context, req CreatePostRequest) error {
    // 旧逻辑
    oldResp, oldErr := oldService.CreatePost(ctx, req)
    
    // 新逻辑异步调用
    go func() {
        newResp, newErr := newService.CreatePost(ctx, req)
        
        // 对比结果
        if !reflect.DeepEqual(oldResp, newResp) {
            logInconsistency(oldResp, newResp)
        }
    }()
    
    return oldErr
}
```

---

## 维度九：性能优化与容量规划

### 问题9.2
在社交内容系统中，你会优先优化哪些性能瓶颈？如何量化优化效果？

**考察点：性能优化方法论、指标意识**

#### 参考答案：

**性能优化优先级：**

**一、瓶颈分析与优先级排序**

| 瓶颈类型 | 优先级 | 预期收益 | 优化方案 |
|---------|-------|--------|
| 数据库慢查询 | 高 | 50%+ | 索引优化、分库分表 |
| 缓存穿透/击穿 | 高 | 40%+ | 多级缓存、热点治理 |
| N+1查询 | 中高 | 30%+ | 批量查询、DataLoader |
| 前端渲染 | 中 | 20%+ | CDN、SSR |

**二、性能基准测试**

```go
func BenchmarkFeedQuery(b *testing.B) {
    for i := 0; i < b.N; i++ {
        GetFeed(1, 20)
    }
}

// 对比优化前后
func CompareBeforeAfter(before time.Duration, after time.Duration) {
    improvement := float64(before-after) / float64(before) * 100
    log.Printf("Performance improved by %.1f%%", improvement)
}
```

---

## 维度十：架构演进与技术债务

### 问题10.2
作为Tech Lead，你如何让团队接受架构重构？如何让业务方理解和支持架构优化工作？

**考察点：领导力、沟通协调能力**

#### 参考答案：

**架构变革管理方案：**

**一、数据驱动的说服**

```markdown
# 架构优化提案模板

## 问题现状
- 当前系统响应时间P99: 800ms
- 生产事故频率: 每月5次
- 开发效率下降30%

## 预期收益
- 响应时间P99降低到200ms
- 生产事故降低到每月1次
- 新功能开发周期缩短50%

## 投入评估
- 开发: 2人2周
- 测试: 1人1周
```

**二、最小可行验证**

```go
// 试点项目：选择低风险高收益部分先验证
func PilotRefactorProject() {
    // 选择评论模块作为试点
    result := refactorCommentModule()
    
    // 展示成果
    showDemo(result)
    
    // 获取反馈
    getFeedback()
}
```

**三、渐进式重构**

```
里程碑1: 重构用户模块 (1个月)
里程碑2: 重构内容模块 (1个月)  
里程碑3: 重构互动模块 (1个月)
```

---

## 额外补充面试题

### 补充问题1：系统韧性设计
假设社交内容系统中，如何设计熔断和降级策略？比如当评论服务挂了，feed流还能正常看吗？

**参考答案：**

```go
type CircuitBreaker struct {
    failureThreshold int
    failureCount int
    state string // "closed" | "open" | "half-open"
    lastFailureTime time.Time
}

func (cb *CircuitBreaker) Call(service func() (interface{}, error), fallback func() interface{}) interface{} {
    if cb.state == "open" {
        if time.Since(cb.lastFailureTime) > 30*time.Second {
            cb.state = "half-open"
        } else {
            return fallback()
        }
    }
    
    result, err := service()
    if err != nil {
        cb.failureCount++
        if cb.failureCount >= cb.failureThreshold {
            cb.state = "open"
            cb.lastFailureTime = time.Now()
        }
        return fallback()
    }
    
    cb.failureCount = 0
    cb.state = "closed"
    return result
}

// 使用
func GetPostFeed(ctx context.Context, req GetFeedRequest) (*GetFeedResponse, error) {
    posts, _ := contentService.GetPostList(req)
    
    // 评论服务调用加熔断
    comments := circuitBreaker.Call(
        func() (interface{}, error) {
            return commentService.BatchGetHotComments(postIds)
        },
        func() interface{} {
            return emptyComments() // 降级返回空评论
        },
    )
    
    return aggregateResult(posts, comments)
}
```

---

### 补充问题2：数据一致性
社交内容系统中，点赞数、评论数等统计数据如何保证一致性？

**参考答案：**

```go
// 最终一致+补偿
func LikePost(ctx context.Context, userId, postId uint64) error {
    // 1. 更新数据库
    tx, _ := db.Begin()
    tx.Exec("INSERT INTO likes (user_id, post_id) VALUES (?, ?)", userId, postId)
    tx.Exec("UPDATE posts SET like_count = like_count + 1 WHERE id = ?", postId)
    tx.Commit()
    
    // 2. 删除缓存
    cache.Delete(fmt.Sprintf("post:%d", postId))
    
    // 3. 发送消息
    mq.Send("post-liked", LikeEvent{PostId: postId, UserId: userId})
    return nil
}

// 定期对账
func ReconcileLikeCount(ctx context.Context, startTime time.Time) {
    posts, _ := db.Query("SELECT id FROM posts WHERE updated_at > ?", startTime)
    
    for _, postId := range posts {
        dbCount, _ := db.QueryInt("SELECT COUNT(*) FROM likes WHERE post_id = ?", postId)
        cacheCount, _ := getFromCache(postId)
        
        if dbCount != cacheCount {
            fixCache(postId, dbCount)
        }
    }
}
```

---

### 补充问题3：分库分表
如果用户量和帖子量非常大，你会如何设计分库分表策略？

**参考答案：**

```
用户库分表：按用户ID哈希分表
posts_00, posts_01,... posts_99

话题库分表：按话题ID分表
topics_00, topics_01,... topics_99

// 分库路由
func GetShard(id uint64) int {
    return int(id % 100)
}

func GetPostDB(postId uint64) *Post {
    shard := GetShard(postId)
    db := GetDBByShard(shard)
    return db.QueryPost(postId)
}
```

---

## 面试官总结

这份面试题通过以下方式考察候选人：

1. **基础知识扎实 - 缓存、消息、搜索等技术的理解
2. **架构思维 - 服务拆分、边界设计能力
3. **工程实践 - 实际落地经验、风险控制
4. **沟通协调 - 架构协作、变革管理
5. **学习能力 - 新技术的理解和应用

面试中要追问细节，比如追问技术选型的trade-off，比如为什么选A不选B，要知道什么场景用什么方案。