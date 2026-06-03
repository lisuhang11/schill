# SChill 项目复习与审查报告

## 一、项目概览

**SChill** 是一个基于 Go 微服务架构的轻量级内容社区/社交平台，支持用户注册登录、动态发布、评论互动、关注关系、搜索等功能。

| 类别 | 技术选型 |
|------|---------|
| 语言 | Go 1.25.0 |
| 微服务框架 | go-zero v1.10.0 |
| RPC 协议 | gRPC + Protobuf |
| ORM | GORM v1.31.1 |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7.0 |
| 消息队列 | Kafka 3.9.1 (KRaft 模式) |
| 搜索引擎 | Elasticsearch 8.6.1 |
| 服务发现 | Etcd 3.5 |
| 对象存储 | MinIO |
| 认证 | JWT |
| 前端 | Next.js 15 / React 19 / TypeScript 5 / Tailwind CSS 3 |
| 数据同步 | Canal (MySQL binlog → Kafka → ES) |

---

## 二、微服务架构

```
┌──────────────────────────────────────────────────────────┐
│                    Gateway (HTTP :8086)                   │
│              JWT Auth / CORS / REST API                   │
└────┬───────┬──────┬───────┬──────┬───────┬───────┬──────┘
     │       │      │       │      │       │       │
     ▼       ▼      ▼       ▼      ▼       ▼       ▼
  user    relation content comment interaction feed   search
  :8080    :8081    :8082    :8083   :8084    :8087   :8088
  (gRPC)  (gRPC)  (gRPC)  (gRPC)  (gRPC)  (gRPC)  (gRPC)
     │       │      │       │      │       │       │
     └───────┴──────┴───────┴──────┴───────┘       │
                    │                                │
                    ▼                                ▼
              MySQL (写) ←── Canal ──→ Kafka ──→ canal-sync ──→ ES (读/搜索)
                    │
                    ▼
              Redis (缓存/状态/锁)
```

### 服务职责

| 服务 | 职责 |
|------|------|
| **gateway** | HTTP API 网关，JWT 认证，请求代理到各 RPC 服务 |
| **user-rpc** | 用户注册/登录/资料/统计，消费 Kafka 事件更新用户统计 |
| **relation-rpc** | 关注/取关，关注状态检查，互关检测 |
| **content-rpc** | 帖子 CRUD，话题管理，置顶/加精，浏览量，消费 Kafka 事件更新帖子统计 |
| **comment-rpc** | 评论创建/删除/投票，游标分页，Kafka 事件生产 |
| **interaction-rpc** | 帖子点赞/收藏/分享，批量状态检查，Redis 存储交互状态 |
| **feed-rpc** | Feed 聚合：合并帖子 + 作者信息 + 用户交互状态 |
| **search-rpc** | ES 全文搜索帖子/用户/话题 |
| **canal-sync** | 消费 Canal binlog → 同步 ES 索引 |

### 事件驱动数据流

```
content-rpc ──→ Kafka(post-created) ──→ user-rpc (更新用户发帖数)
relation-rpc ──→ Kafka(user-followed) ──→ user-rpc (更新关注统计)
comment-rpc ──→ Kafka(comment-created) ──→ content-rpc (更新帖子评论数)
interaction-rpc ──→ Kafka(post-star/collect) ──→ content-rpc (更新帖子互动数)
Canal ──→ Kafka(canal_topic) ──→ canal-sync ──→ Elasticsearch
```

---

## 三、核心技术知识点

### 3.1 go-zero 框架

go-zero 是一个集成了 Web 和 RPC 框架的 Go 微服务框架，核心特性：

1. **handler → logic → model 三层架构**
   - `handler`: 请求入口，参数校验和协议转换
   - `logic`: 业务逻辑层，每个 RPC 方法对应一个独立的 Logic 结构体
   - `model`: 数据访问层

2. **通过 goctl 代码生成**
   - 从 `.proto` 文件自动生成 gRPC 服务代码
   - 从 `.api` 文件生成 HTTP 网关代码

3. **内置功能**
   - 服务注册发现 (Etcd)
   - 负载均衡
   - 熔断器 (breaker)
   - 限流器 (limiter)
   - 超时控制
   - 链路追踪 (OpenTelemetry + Zipkin)

### 3.2 gRPC + Protobuf

项目中定义了 7 个 `.proto` 文件，核心要点：

- **枚举在 Protobuf 中默认值为 0**，必须注意未设置和零值的区别
- **字段编号**是 Protobuf 序列化的关键，不可随意变更
- **gRPC 状态码**：标准码 0-16，业务错误应使用 `status.New(codes.Internal, msg)` + details 而非混用
- **流式 RPC**：当前项目全部使用 Unary RPC，大列表场景可考虑 Server Streaming

### 3.3 JWT 认证

```go
// Claims 结构
type Claims struct {
    jwt.RegisteredClaims
    UserID       uint64
    TokenVersion int64   // 支持令牌撤销（当前未实现验证）
}
```

关键知识点：
- **Access Token**：短期有效（如 15 分钟），用于 API 鉴权
- **Refresh Token**：长期有效（如 7 天），用于刷新 Access Token
- **HS256 vs RS256**：HS256 是对称加密，secret 泄露则全部令牌可伪造；RS256 是非对称加密，更安全
- **Token Version**：在 Redis 存储用户 token version，修改密码时递增，实现旧令牌批量失效

### 3.4 Kafka 事件驱动

项目中的 Kafka 使用模式：

- **事件类型**：`post-created`, `comment-created`, `user-followed` 等
- **事件信封**：包含 EventID、EventType、AggregateID 等元数据
- **幂等性**：通过 EventID + Redis SetNX 实现
- **DLQ（死信队列）**：处理失败的消息路由到 DLQ topic
- **批量消费**：Canal 同步使用批量消费提升吞吐

**Producer 压缩与攒批机制（面试重点）：**

```
用户操作 → 应用层 go func() 异步 SendEvent（每条消息一个调用）
                ↓
         Sarama AsyncProducer（客户端层）
           ├── Snappy 压缩（减少网络传输量）
           ├── 攒批策略：每 10ms 或攒够 100 条 flush 一次
           └── acks=WaitForAll（等待所有 ISR 副本确认）
                ↓
         Kafka Broker（单节点，开发环境）
```

| Producer 类型 | 压缩 | 攒批 | 使用方 | 适用场景 |
|--------------|------|------|--------|---------|
| AsyncProducer | **Snappy** | 10ms/100条 | interaction, comment | 高吞吐、可接受短暂延迟 |
| SyncProducer | 无 | 无（逐条同步） | content, relation | 需要立即确认的场景 |

**为什么不需要应用层时间窗口合并？**
1. Sarama AsyncProducer 已经在网络层自动攒批（10ms/100条），应用层不需要重复造轮子
2. 点赞/收藏计数操作是 `UPDATE ... SET count = count + 1`，天然幂等，即使重复消费也正确
3. 应用层做 5s 窗口合并需要额外维护内存 Map `{post_id: delta}`，增加了内存开销和进程崩溃丢失风险
4. 5s 延迟会显著降低计数的实时性，用户体验变差

关键知识点：
- **至少一次 vs 精确一次**：Kafka 默认至少一次，需通过幂等性实现精确一次
- **幂等性标记时机**：必须在业务操作**成功后**标记，而非之前
- **事务消息**：Kafka 支持事务，可实现"本地操作 + 发消息"的原子性
- **Consumer Group Rebalance**：会导致重复消费，需要幂等性保护

### 3.5 MySQL Binlog + Canal + ES 同步

```
MySQL Binlog → Canal Server → Kafka → canal-sync → Elasticsearch
```

关键知识点：
- **Binlog 格式**：必须使用 ROW 格式才能捕获具体数据变更
- **Canal 原理**：伪装成 MySQL Slave，接收 Binlog 事件
- **最终一致性**：MySQL 写入成功后，ES 索引更新有延迟（通常 100ms-1s）
- **全量重建**：项目提供了 `cmd/rebuild/main.go` 支持全量索引重建
- **软删除处理**：Canal 消息中 `deleted_at` 字段非空时触发 ES 文档删除

### 3.6 Redis 缓存策略

**多层缓存架构：**

| 层级 | 内容 | TTL | 说明 |
|------|------|-----|------|
| Post Base | 帖子基础信息（标题、内容） | 长 TTL | 不常变，缓存命中率高 |
| Post Stats | 点赞数、评论数等 | 短 TTL | 频繁变化，接受短暂不一致 |
| Feed | 聚合后的 Feed 数据 | 10s-30s | 极短 TTL，保证新鲜度 |
| Interaction | 点赞/收藏状态 (Set) | 7 天 | 交互状态的源之一 |
| Idempotency | 事件幂等标记 | 24h | 防重复消费 |

**Redis 数据结构使用：**
- String: 帖子基础信息、统计计数、Feed 数据
- Set: 点赞用户集合、收藏用户集合
- Hash: 用户交互关系（toggle 脚本）
- ZSet: 评论列表（按时间排序）

### 3.7 GORM ORM 使用模式

```go
// 软删除 - gorm.DeletedAt 类型会自动过滤
type User struct {
    DeletedAt gorm.DeletedAt  // GORM 自动添加 WHERE deleted_at IS NULL
}

// 软删除 - *time.Time 类型需手动过滤
type Comment struct {
    DeletedAt *time.Time  // 需手动添加 WHERE deleted_at IS NULL
}

// 事务
db.Transaction(func(tx *gorm.DB) error { ... })

// 批量查询
db.Where("id IN ?", ids).Find(&results)
```

关键知识点：
- `gorm.DeletedAt` vs `*time.Time` 的软删除行为不同
- `Updates(map)` 不会触发 Hook，`Updates(struct)` 会
- 读写分离通过两个 DB 连接实现 (`DBRead` / `DBWrite`)
- N+1 查询问题：循环中执行单条查询是常见性能杀手

### 3.8 Elasticsearch 搜索

```json
// 多字段匹配搜索
{
  "multi_match": {
    "query": "关键词",
    "fields": ["title^4", "summary^3", "content^2"],
    "type": "best_fields"
  }
}
```

关键知识点：
- **倒排索引**：ES 的核心数据结构，实现快速全文搜索
- **IK 分词器**：中文分词插件，支持 ik_smart（粗粒度）和 ik_max_word（细粒度）
- **multi_match 类型**：best_fields（取最佳匹配字段）、most_fields（所有字段匹配度求和）、cross_fields（跨字段匹配）
- **Bulk API**：批量索引/更新/删除，减少网络开销
- **Refresh Interval**：默认 1s，设置为 false 可提升写入性能但增加搜索延迟

### 3.9 Redis Lua 脚本

项目中使用 Lua 脚本实现原子操作：

```lua
-- 点赞操作（原子性）
local is_member = redis.call('SISMEMBER', KEYS[1], ARGV[1])
if is_member == 1 then
    return {1, tonumber(redis.call('GET', KEYS[2]) or 0)}
end
redis.call('SADD', KEYS[1], ARGV[1])
local new_count = redis.call('INCR', KEYS[2])
return {0, new_count}
```

关键知识点：
- **原子性**：Lua 脚本在 Redis 中执行期间不会被其他命令打断
- **阻塞风险**：复杂脚本会阻塞 Redis，应保持简短
- **集群兼容**：所有操作的 key 必须在同一个 hash slot

---

## 四、发现的问题（按严重程度排列）

### 🔴 P0 - 严重问题

#### 4.1 密码哈希使用 SHA-256 而非 bcrypt

**位置**: `common/cryptx/crypt.go:9-11`

```go
func PasswordEncrypt(password string) string {
    hash := sha256.Sum256([]byte(password))
    return fmt.Sprintf("%x", hash[:])
}
```

**问题**：
- SHA-256 是通用哈希算法，设计目标是速度而非安全性
- **没有盐值（salt）**，相同密码产生相同哈希，可被彩虹表攻击
- 单次 SHA-256 在 GPU 上每秒可计算数十亿次

**修复方案**：
```go
import "golang.org/x/crypto/bcrypt"

func PasswordEncrypt(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) // cost=12
    return string(bytes), err
}

func PasswordVerify(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

#### 4.2 授权绕过：DeleteComment 信任客户端 userId

**位置**: `service/comment/rpc/comment.proto:38`

```protobuf
message DeleteCommentReq {
    uint64 commentId = 1;
    uint64 userId = 2;  // ⚠️ 来自客户端，不可信！
}
```

**问题**：客户端可以传入任意 userId 删除他人评论。

**修复方案**：userId 应从 JWT token 的 context 中提取，不在 protobuf 请求中暴露：
```protobuf
message DeleteCommentReq {
    uint64 commentId = 1;
    // userId 从 gRPC metadata 中获取
}
```

#### 4.3 UUID 包为空

**位置**: `common/uuid/uuid.go`

```go
package uuid
// 文件只有 package 声明，没有任何实现
```

**问题**：分布式环境依赖数据库自增 ID，存在单点瓶颈和 ID 碰撞风险。

**修复方案**：使用 Snowflake 或 sonyflake：
```go
import "github.com/sony/sonyflake"

var sf = sonyflake.NewSonyflake(sonyflake.Settings{})
func NextID() (uint64, error) { return sf.NextID() }
```

#### 4.4 Kafka 消费者幂等性标记时机错误

**位置**: `common/kafka/consumer.go:149-168`

```go
// ⚠️ 问题：标记在业务处理之前
if h.idempotencyStore != nil {
    ok, _ := h.idempotencyStore.TryMark(ctx, h.group, envelope.EventID)
    if !ok { continue } // 已处理，跳过
}
// 如果这里 panic 或返回 error，消息已标记为"已处理"但实际未处理
if err := h.handler.Handle(ctx, msg, envelope); err != nil {
    continue // 消息丢失！
}
```

**修复方案**：标记必须在 Handle 成功之后：
```go
if err := h.handler.Handle(ctx, msg, envelope); err != nil {
    continue // 不标记，等待重试
}
h.idempotencyStore.TryMark(ctx, h.group, envelope.EventID) // 成功后再标记
session.MarkMessage(msg, "")
```

#### 4.5 GORM 软删除机制不一致

**位置**: 
- `user_model.go:22` 使用 `gorm.DeletedAt`
- `comment_model.go:22` 使用 `*time.Time`

**问题**：
- `gorm.DeletedAt` 自动在查询中添加 `WHERE deleted_at IS NULL`
- `*time.Time` 不会自动过滤，需要手动添加
- 导致代码中大量手写 `deleted_at IS NULL`，且容易遗漏

**修复方案**：统一使用 `gorm.DeletedAt`，或封装统一的查询 Scope。

---

### 🟠 P1 - 高优先级

#### 4.6 基础设施全部无认证

**位置**: `deploy/docker-compose.yml`, `deploy/docker-compose.prod.yml`

| 组件 | 问题 |
|------|------|
| Etcd | `ALLOW_NONE_AUTHENTICATION=yes` |
| Elasticsearch | `xpack.security.enabled=false` |
| Redis | 无 `requirepass` |
| Kafka | 全部 `PLAINTEXT`，无 SASL/SSL |

#### 4.7 Kubernetes 使用 hostPath 存储

**位置**: `deploy/k8s/docker-desktop/infra-mysql.yaml`

```yaml
volumes:
  - name: data
    hostPath:
      path: /var/lib/schill/mysql
```

**问题**：Pod 重新调度到不同节点会丢失数据，不支持多副本。

**修复方案**：使用 StatefulSet + PersistentVolumeClaim。

#### 4.8 数据库缺失关键索引

**post 表** Feed 查询需要复合索引：
```sql
-- 当前只有单列索引
KEY `idx_user_id` (`user_id`),
KEY `idx_created_at` (`created_at`)

-- 需要添加
KEY `idx_user_deleted_top_created` (`user_id`, `deleted_at`, `is_top` DESC, `created_at` DESC)
```

**comment 表** 评论热排需要复合索引：
```sql
-- 当前
KEY `idx_post_parent` (`post_id`, `parent_id`)

-- 需要
KEY `idx_post_parent_deleted_hot` (`post_id`, `parent_id`, `deleted_at`, `like_count` DESC, `created_at` DESC)
```

#### 4.9 Feed 聚合 5 次串行 RPC 调用

**位置**: `service/feed/rpc/internal/logic/getfeedlistlogic.go:57-81`

```go
postResp, _ := l.svcCtx.ContentRpc.GetPostList(...)   // 1
summaryMap := l.loadSummaries(postIDs)                  // 2
authorMap := l.loadAuthors(userIDs)                     // 3
starMap, collectionMap := l.loadViewerInteraction(...)   // 4
followMap := l.loadFollowStatus(...)                     // 5
```

**问题**：5 次 RPC 串行执行，每次都有网络延迟。假设每次 10ms，总共 50ms。

**修复方案**：使用 `errgroup.Group` 并行调用：
```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { summaryMap = l.loadSummaries(postIDs); return nil })
g.Go(func() error { authorMap = l.loadAuthors(userIDs); return nil })
g.Go(func() error { starMap, collectionMap = l.loadViewerInteraction(...); return nil })
g.Go(func() error { followMap = l.loadFollowStatus(...); return nil })
g.Wait()
```

#### 4.10 IncViewCount 是空操作

**位置**: `service/content/rpc/internal/logic/incviewcountlogic.go:29-47`

```go
// 只检查帖子是否存在，但没有实际增加 view_count
var post model.Post
l.svcCtx.DB.Where("id = ? AND deleted_at IS NULL", in.GetPostId()).Take(&post)
return &pb.IncViewCountResp{Success: true}, nil
```

#### 4.11 搜索暴露私有帖子

搜索查询没有添加 visibility 过滤条件，用户可以搜索到其他人的私有帖子。

**修复方案**：
```json
{
  "query": {
    "bool": {
      "must": [{ "multi_match": { ... } }],
      "filter": [{ "term": { "visibility": 90 } }]
    }
  }
}
```

#### 4.12 帖子更新无乐观锁

**位置**: `service/content/rpc/internal/logic/updatepostlogic.go:80-92`

```go
// 先查询（事务外）
post, _ := loadOwnedPost(...)
// 再更新（事务内）- 没有 version 检查
tx.Model(&model.Post{}).Where("id = ?", postId).Updates(map[string]interface{}{...})
```

**修复方案**：添加乐观锁（version 字段或 `updated_at` 条件）：
```go
result := tx.Model(&model.Post{}).
    Where("id = ? AND updated_at = ?", postId, post.UpdatedAt).
    Updates(data)
if result.RowsAffected == 0 {
    return ErrConcurrentUpdate
}
```

#### 4.13 分布式锁无所有权验证

**位置**: `common/cacheprotect/cacheprotect.go:96-102`

```go
func TryLock(ctx, client, key, ttl) (bool, error) {
    return client.SetNX(ctx, key, lockValue, ttl)  // lockValue 是常量！
}

func ReleaseLock(ctx, client, key) error {
    return client.Del(ctx, key)  // 可能删除其他实例的锁
}
```

**修复方案**：使用唯一锁值 + Lua 脚本验证所有权后释放：
```lua
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
```

---

### 🟡 P2 - 中优先级

#### 4.14 comment/comment_stat 双重计数冗余

`comment` 表和 `comment_stat` 表都维护 `reply_count`、`like_count`、`dislike_count`，容易出现数据不一致。

**建议**：二选一，推荐只保留 `comment_stat` 作为计数的单一来源。

#### 4.15 post_topic 表无软删除

`post_topic` 表没有 `deleted_at` 字段，删除帖子时对关联记录执行物理删除，与项目统一的软删除策略不一致。

#### 4.16 网关评论列表 N+1 查询

**位置**: `service/gateway/internal/handler/comment.go:128-144`

每个根评论都单独调用一次 `GetReplyList` RPC。20 个根评论 = 20 次 RPC。

**修复方案**：批量获取或支持多评论 ID 的回复查询接口。

#### 4.17 Canal 同步无幂等性保护

Canal 消费者没有实现幂等性，Kafka rebalance 导致重复消费时会产生重复 ES 写入。

#### 4.18 搜索缓存 key 未做关键词归一化

"Hello" 和 "hello" 产生不同的缓存 key，缓存利用率低。

#### 4.19 MinIO 对象名可预测

```go
fmt.Sprintf("avatars/%d/%d_%d.%s", userId, timestamp, userId, ext)
```

攻击者可枚举用户头像。建议加入随机 UUID 组件。

#### 4.20 user.proto 状态值与数据库定义不一致

- proto: `0:正常 1:封禁 2:删除`
- db.sql: `1正常，2禁言，3冻结`

---

### 🟢 P3 - 低优先级 / 改进建议

#### 4.21 加密随机数用于 TTL Jitter

`cacheutil.go` 使用 `crypto/rand` 生成 TTL jitter，在熵不足的容器环境中可能阻塞。建议使用 `math/rand`。

#### 4.22 搜索 multi_match 缺少 fuzziness

无拼写容错，搜索 "helo" 匹配不到 "hello"。建议添加 `fuzziness: "AUTO"`。

#### 4.23 ES 客户端缺少连接池配置

无超时、重试、连接池大小配置。

#### 4.24 JWT Token Version 未实现验证

`Claims` 中有 `TokenVersion` 字段，但验证逻辑缺失，令牌撤销功能形同虚设。

#### 4.25 Kafka Producer SendRaw 在 stop 时静默丢消息

```go
case <-p.stopChan:
    return nil  // 应该返回 error
```

#### 4.26 三种不同的 Kafka Consumer 实现

项目中有 3 种消费者实现（common/kafka、comment consumer、interaction consumer），各实现不同的 bug，缺乏统一抽象。

---

## 五、架构改进建议

### 5.1 双写问题：事务性发件箱模式

**当前问题**：数据库写入和 Kafka 消息发送不是原子的。

```go
// createpostlogic.go - 当前代码
tx.Commit()  // DB 写入
go func() {
    producer.SendEvent(...)  // 可能失败！
}()
```

**推荐方案**：使用 Transactional Outbox 模式：

```sql
CREATE TABLE outbox (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregate_type VARCHAR(50),
    aggregate_id BIGINT,
    event_type VARCHAR(50),
    payload JSON,
    created_at DATETIME(3),
    INDEX idx_created (created_at)
);
```

```go
// 同一事务中写入业务数据 + outbox
tx.Create(&post)
tx.Create(&Outbox{EventType: "post-created", Payload: ...})
tx.Commit()

// 独立的后台进程轮询 outbox 表，发送到 Kafka
```

### 5.2 缓存一致性：Cache-Aside + 版本号

**当前问题**：Feed 缓存无失效机制，新帖子发布后 Feed 可能过时。

**推荐方案**：
```go
// 用户发帖时递增 Feed 版本号
redis.Incr("feed_version:user:" + userID)

// 读取时检查版本
cachedVersion := redis.Get("feed_version:cached:" + cacheKey)
currentVersion := redis.Get("feed_version:user:" + userID)
if cachedVersion != currentVersion {
    // 缓存过时，回源数据库
}
```

### 5.3 分布式 ID 生成

**当前**：MySQL AUTO_INCREMENT

**建议**：Snowflake / Sonyflake，支持水平扩展。

### 5.4 限流与熔断

go-zero 框架内置了 breaker 和 limiter，但项目中未显式配置：

```go
// gateway 级别配置
rest.WithTimeout(5000),
rest.WithMaxConns(10000),
```

建议：
- 登录接口：令牌桶限流，防暴力破解
- 搜索接口：限制 QPS，防 ES 过载
- Feed 接口：熔断保护，降级返回缓存数据

### 5.5 读写分离优化

项目已使用 `DBRead` / `DBWrite`，但部分查询仍使用 `DBWrite`。建议审查所有只读操作统一使用 `DBRead`。

### 5.6 链路追踪与监控

项目引入了 OpenTelemetry + Zipkin + Prometheus + Pyroscope，建议补充：
- 业务指标埋点（注册量、发帖量、活跃用户）
- Kafka 消费延迟监控
- ES 同步延迟告警

---

## 六、面试高频考点

### 6.1 微服务拆分原则
- **按业务领域拆分**（DDD）：User、Content、Comment、Relation 各为一个服务
- **数据库独立**：每个服务有自己的数据表，通过 RPC 通信
- **避免分布式事务**：使用事件驱动实现最终一致性

### 6.2 最终一致性实现
- **Kafka 事件**：发布订阅模式，生产者不关心消费者处理结果
- **幂等性**：通过 EventID + Redis SetNX 保证
- **补偿机制**：DLQ 死信队列处理失败消息

### 6.3 缓存策略
- **Cache-Aside**：先查缓存，未命中查 DB 并回写缓存
- **Write-Through / Write-Behind**：写操作同步/异步更新缓存
- **缓存穿透**：布隆过滤器或缓存空值（本项目使用空值缓存）
- **缓存击穿**：分布式锁（本项目 cacheprotect 包）
- **缓存雪崩**：随机 TTL（本项目 cacheutil Jitter 函数）

### 6.4 gRPC vs REST
- gRPC 基于 HTTP/2，支持多路复用、流式传输
- Protobuf 序列化效率高于 JSON
- 强类型接口定义，代码自动生成
- 适用于内部微服务通信

### 6.5 ES 搜索原理
- 倒排索引：词 → 文档映射
- 分词器：IK 中文分词
- 相关性评分：TF-IDF → BM25
- 准实时搜索：refresh interval

### 6.6 Kafka 核心概念
- Topic / Partition / Consumer Group
- Offset 管理（自动提交 vs 手动提交）
- 消息语义：at-most-once / at-least-once / exactly-once
- ISR（In-Sync Replicas）与高可用

### 6.7 常见 Go 并发模式
- `errgroup.Group`：并行执行 + 错误传播
- `sync.WaitGroup` + channel：收集并发结果
- Context 传递：超时控制、取消信号
- `select` 多路复用

---

## 七、总结

### 项目亮点
1. **完整的技术栈**：go-zero + gRPC + Kafka + ES + Canal，涵盖微服务全链路
2. **清晰的服务拆分**：按领域划分，职责明确
3. **事件驱动架构**：解耦服务间依赖，支持异步处理
4. **多层缓存策略**：Redis 基础缓存 + 本地缓存 + 分层 TTL
5. **搜索同步方案**：Canal CDC 实现 MySQL → ES 实时同步

### 需要改进的核心点
1. **安全**：密码哈希、授权验证、基础设施认证
2. **可靠性**：Kafka 幂等性、分布式锁、双写一致性
3. **性能**：Feed 聚合并行化、数据库索引、N+1 查询
4. **代码质量**：软删除一致性、消费者实现统一、重复代码消除
