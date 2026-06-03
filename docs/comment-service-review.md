# Comment Service 全面技术审查报告

> 审查日期：2026-06-03
> 服务：`comment.rpc` (端口 8083)
> 范围：proto 定义 → 数据模型 → 业务逻辑 → Kafka 消费 → 缓存策略 → 中间件原理

---

## 目录

1. [一、服务概览](#一服务概览)
2. [二、数据模型](#二数据模型)
3. [三、API 接口全览](#三api-接口全览)
4. [四、核心业务流程](#四核心业务流程)
5. [五、Kafka 事件驱动架构](#五kafka-事件驱动架构)
6. [六、Redis 缓存架构详解](#六redis-缓存架构详解)
7. [七、投票系统设计](#七投票系统设计)
8. [八、ServiceContext 依赖关系](#八servicecontext-依赖关系)
9. [九、存在的问题](#九存在的问题)
10. [十、技术方案选型对比](#十技术方案选型对比)
11. [十一、改进建议优先级](#十一改进建议优先级)
12. [十二、服务依赖图](#十二服务依赖图)

---

## 一、服务概览

### 1.1 技术栈

| 组件 | 技术选型 | 版本/配置 |
|------|----------|-----------|
| RPC 框架 | gRPC + go-zero zrpc | port 8083 |
| 服务注册 | Etcd | `127.0.0.1:2379`, key `comment.rpc` |
| 数据库 | MySQL (GORM) | `schill`, 连接池 100/20 |
| 缓存 | Redis (go-redis v9 + go-zero cache) | `127.0.0.1:6379`, DB 0 |
| 消息队列 | Kafka (Sarama) | `127.0.0.1:9092`, 4 topics + DLQ |
| 上游依赖 | Content RPC (端口 8082) | 验证帖子存在 |

### 1.2 端口与依赖

```
comment.rpc:8083
├── MySQL:3306 (主存储)
├── Redis:6379 (缓存 + 分布式锁 + 限流 + 投票计数)
├── Kafka:9092 (生产者: 4 topics / 消费者: 3 topics)
├── ContentRpc:8082 (帖子验证)
└── Etcd:2379 (服务注册)
```

### 1.3 Kafka Topic 清单

| Topic | 方向 | 用途 |
|-------|------|------|
| `comment-create` | 生产 + 消费 | 评论创建事件（缓存同步） |
| `comment-created` | 仅生产 | 评论创建消息（统计同步） |
| `comment-deleted` | 生产 + 消费 | 评论删除事件 |
| `comment-vote` | 生产 + 消费 | 投票事件 |
| `comment-dlq` | 仅消费（重试耗尽） | 死信队列 |

---

## 二、数据模型

### 2.1 4 张核心表

```
┌─────────────┐     ┌──────────────────┐
│   comment    │────→│ comment_content   │
│  (主表)      │ 1:1 │ (内容分离)        │
└──────┬──────┘     └──────────────────┘
       │
       ├──→ comment_vote (用户投票记录, N:1)
       │
       └──→ comment_stat (聚合统计, 1:1)
```

### 2.2 Comment（评论主表）

```go
type Comment struct {
    ID            uint64     // 主键
    PostID        uint64     // 帖子ID（联合索引 idx_post_parent）
    UserID        uint64     // 发表用户ID（索引 idx_user_created）
    ParentID      uint64     // 父评论ID，0=根评论
    ReplyToUserID *uint64    // @回复的目标用户ID（冗余，避免关联查询）
    Level         uint8      // 层级：1=根评论, 2=二级回复, 3=三级...
    Status        uint8      // 状态：1=正常, 2=隐藏, 3=删除
    ReplyCount    int32      // 直接回复数（冗余计数器）
    LikeCount     int32      // 点赞数（冗余计数器）
    DislikeCount  int32      // 点踩数（冗余计数器）
    Ip            string     // 发布IP
    IpLoc         string     // IP所在地
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time // 软删除
}
```

**表设计分析：**

- **冗余计数器的权衡**：`reply_count`、`like_count`、`dislike_count` 存储在 comment 主表上，避免了 JOIN `comment_stat` 的开销。代价是需要通过 Kafka 消费者异步同步计数，存在短暂不一致窗口。
- **`ReplyToUserID` 冗余**：避免显示回复列表时 JOIN user 表获取"回复给谁"的信息，提升读性能。
- **联合索引 `idx_post_parent(post_id, parent_id)`**：覆盖"获取某帖子下根评论"（`parent_id=0`）和"获取某评论的回复"（`parent_id=commentId`）两个高频查询。

### 2.3 CommentContent（内容分离表）

```go
type CommentContent struct {
    ID          uint64  // 自增主键
    CommentID   uint64  // 评论ID（唯一索引 uk_comment_id）
    Content     string  // 内容（mediumtext，支持富文本/Markdown）
    ContentType uint8   // 类型：1=纯文本, 2=Markdown, 3=HTML
}
```

**设计意图**：将大文本字段（`mediumtext`）从主表分离，避免每次查询评论列表时扫描大字段，减少 InnoDB 行溢出页。

### 2.4 CommentStat（统计表）

```go
type CommentStat struct {
    CommentID    uint64  // 主键（= comment.id）
    ReplyCount   uint32
    LikeCount    uint32
    DislikeCount uint32
}
```

**设计意图**：独立统计表，与 comment 主表形成读写分离。写入走 Kafka 异步，读取时可选择从 comment 冗余字段或此表获取。

### 2.5 CommentVote（投票记录表）

```go
type CommentVote struct {
    ID        uint64  // 自增主键
    CommentID uint64  // 评论ID
    UserID    uint64  // 用户ID
    VoteType  uint8   // 1=赞, 2=踩
}
```

**唯一约束**：`uk_comment_user(comment_id, user_id)` 保证每用户对每条评论只有一条投票记录。

---

## 三、API 接口全览

| RPC 方法 | 请求关键字段 | 响应 | 认证 | 幂等 |
|----------|-------------|------|------|------|
| `CreateComment` | userId, postId, parentId, content | CommentInfo | ❌ 信任 userId | ✅ DB 事务 |
| `DeleteComment` | commentId, userId | success | ❌ 信任 userId，比对 owner | ✅ 软删除 |
| `GetCommentList` | postId, cursor, pageSize, sortType | list, total, hasMore, nextCursor | ❌ 公开 | ✅ 只读 |
| `GetReplyList` | commentId, cursor, pageSize | list, total, hasMore, nextCursor | ❌ 公开 | ✅ 只读 |
| `VoteComment` | commentId, userId, voteType(0/1/2) | likeCount, dislikeCount, isLiked, isDisliked | ❌ 信任 userId | ✅ Lua 原子 |

> **注意**：CreateComment、DeleteComment、VoteComment 都存在**P0-1 同款问题**：信任客户端传入的 `userId`，可越权操作。

---

## 四、核心业务流程

### 4.1 创建评论 (`CreateComment`)

```
┌──────────┐    ┌───────────┐    ┌──────────┐    ┌──────────┐
│ 参数校验  │───→│ 验证帖子   │───→│ 查父评论   │───→│ DB 事务   │
│ postId≠0 │    │ 存在性     │    │ (计算层级) │    │ 3表写入   │
│ content≠""│    │ ContentRpc │    │           │    │           │
└──────────┘    └───────────┘    └──────────┘    └────┬─────┘
                                                       │
                                          ┌────────────┴────────────┐
                                          ↓                         ↓
                                    ┌──────────┐            ┌──────────────┐
                                    │ 失效缓存  │            │ Kafka 异步发送 │
                                    │ 评论列表  │            │ 2条消息        │
                                    │ 回复列表  │            │ cache_sync    │
                                    └──────────┘            │ stat_sync     │
                                                            └──────────────┘
```

**事务内操作（原子性保证）**：
1. `INSERT comment` — 创建评论记录
2. `INSERT comment_content` — 创建内容记录
3. `INSERT comment_stat` — 创建统计记录（初始为 0）

**事务外操作**：
1. 同步失效缓存（Redis DEL）
2. 异步发送 2 条 Kafka 消息

### 4.2 删除评论 (`DeleteComment`)

```
┌──────────┐    ┌───────────┐    ┌───────────┐    ┌──────────┐
│ 查评论    │───→│ 权限校验   │───→│ 失效缓存   │───→│ Kafka    │
│          │    │ userId==  │    │           │    │ 异步发送  │
│          │    │ owner?    │    │           │    │          │
└──────────┘    └───────────┘    └──────────┘    └──────────┘
```

**注意**：DeleteComment 的 logic 层没有执行 DB 写操作（软删除），仅失效缓存 + 发送 Kafka。真正的软删除在 **Kafka 消费者** `handleCommentDeletedEvent` 中执行。这意味着如果 Kafka 消息丢失，评论不会真正被删除。

### 4.3 获取评论列表 (`GetCommentList`) — 最复杂流程

```
请求进入
    │
    ▼
┌──────────────────────────────────────────────────────┐
│ 1. getRootCommentIDsFromRedis(postId, cursor, sortType) │
│    从 Redis ZSet 获取分页 ID 列表 + 缓存元数据 Entry      │
│    ┌──────────────────────────────────────────────┐   │
│    │ ZREVRANGEBYSCORE key max (cursor score)      │   │
│    │ LIMIT 0 pageSize+1                          │   │
│    │ → 多取 1 条判断 hasMore                      │   │
│    │ → 用最后一条的 score 作为 nextCursor           │   │
│    └──────────────────────────────────────────────┘   │
└────────────────────┬─────────────────────────────────┘
                     │
          ┌──────────┴──────────┐
          │ cacheState 判断      │
          └──────────┬──────────┘
                     │
     ┌───────────────┼───────────────┐
     │               │               │
  空列表          有数据             有数据但缓存过期
  (cache miss)   (cache hit)        (stale cache)
     │               │               │
     ▼               ▼               ▼
 rebuildCache   直接使用数据     async refreshCache
 (singleflight                  (goroutine 异步重建)
  + 分布式锁)                     ↓
     │                        使用当前数据返回
     ▼                        (不阻塞用户请求)
 重新查询 Redis
     │
     ▼
┌──────────────────────────────────────────────┐
│ 2. batchGetCommentInfo(ids)                  │
│    Redis Pipeline HGetAll → miss → DB 回源    │
│                                              │
│ 3. batchGetCommentContent(ids)               │
│    Redis Pipeline Get → miss → DB 回源       │
│                                              │
│ 4. assembleCommentList → 组装 CommentInfo[]   │
└──────────────────────────────────────────────┘
```

**缓存重建 (`rebuildCommentCache`) 详细流程：**

```
singleflight.Do("comment:list:{postId}:{sortType}")
    │
    ▼
 再次检查 Redis 缓存是否已由其他请求重建
    │ (双重检查)
    │ 已新鲜 → 直接返回
    │ 不新鲜 ↓
    ▼
TryLock(lockKey, 10s)
    │
    ├── 获取锁成功 ──→ DB 查询 → ZAdd → StoreMarker → ReleaseLock
    │
    └── 获取锁失败 ──→ WaitFor(20次×50ms=1s)
                        │ 轮询检查缓存是否被其他实例重建
                        │ 重建完成 → 返回
                        │ 超时 → 返回错误（降级到 DB）
```

### 4.4 获取回复列表 (`GetReplyList`)

与评论列表相同的缓存模式，差异：
- 查询条件：`parent_id = commentId`（而非 `post_id AND parent_id=0`）
- 仅按时间排序（无 "hot" 模式）
- 回复 ZSet key：`schill:comment:replies:{rootId}:list`

---

## 五、Kafka 事件驱动架构

### 5.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                       comment-rpc 进程                          │
│                                                                 │
│  ┌──────────────┐          ┌──────────────────────────────────┐ │
│  │ gRPC Server  │          │    CommentConsumer               │ │
│  │ (Create/     │          │    (同一进程，不同 goroutine)       │ │
│  │  Delete/     │          │                                  │ │
│  │  Vote)       │          │  ┌────────────────────────────┐  │ │
│  └──────┬───────┘          │  │ ConsumeClaim loop          │  │ │
│         │ 异步发送          │  │  消费 3 个 topic:           │  │ │
│         ▼                  │  │  - comment-create          │  │ │
│  ┌──────────────┐          │  │  - comment-deleted         │  │ │
│  │ KafkaProducer│─────────→│  │  - comment-vote            │  │ │
│  │ (Async)      │  Broker  │  └────────────────────────────┘  │ │
│  └──────────────┘          └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**生产者配置（common/kafka/producer.go）**：
- 类型：`AsyncProducer`
- `RequiredAcks`: `WaitForLocal`（仅 Leader 确认，非 WaitForAll）
- 压缩：Snappy
- Flush: 10ms / 100 条消息
- `Idempotent`: false（未开启幂等生产者）
- `MaxOpenRequests`: 20

**消费者配置**：
- `Rebalance.Strategy`: `BalanceStrategyRange`（按分区范围分配，非 RoundRobin）
- `Offsets.Initial`: `OffsetOldest`（从最早未消费位置开始）

### 5.2 事件消息格式

每条 Kafka 消息使用 **EventEnvelope** 包装：

```json
{
  "event_id": "comment.created.cache_sync:{commentId}:{timestamp}",
  "event_type": "comment.created.cache_sync",
  "aggregate_type": "comment",
  "aggregate_id": "{commentId}",
  "producer": "comment-rpc",
  "schema_version": 1,
  "occurred_at": 1717401600000,
  "data": { /* CommentCreateEvent / VoteEvent 等 */ }
}
```

### 5.3 三个消费者处理逻辑

#### 5.3.1 `handleCommentCreateEvent` — 缓存预热

```
收到 CommentCreateEvent
    │
    ├─→ Redis HMSet commentInfoHash（评论元数据）
    ├─→ Redis Set commentContentKey（评论内容）
    │
    ├─→ 如果 parentId==0（根评论）：
    │   └─→ ZAdd postCommentsZSet (score=created_at)
    │       └─→ IncrBy postCommentCount
    │
    └─→ 如果 parentId>0（回复）：
        ├─→ ZAdd replyListZSet (score=created_at)
        └─→ HIncrBy parentInfoHash.reply_count +1
        └─→ IncrBy replyCount +1
```

#### 5.3.2 `handleCommentDeletedEvent` — 软删除 + 缓存清理

```
收到 CommentDeletedMessage
    │
    ├─→ DB 事务：
    │   ├─→ UPDATE comment SET deleted_at=NOW(), status=3
    │   └─→ UPDATE parent.comment SET reply_count = reply_count - 1
    │
    ├─→ Redis DEL commentInfoKey, commentContentKey
    ├─→ Redis ZRem postCommentsZSet (移除 ID)
    ├─→ Redis ZRem hotCommentsZSet (移除 ID)
    ├─→ Redis IncrBy postCommentCount -1
    │
    └─→ 如果 parentId>0：
        ├─→ Redis ZRem replyListZSet
        ├─→ Redis HIncrBy parentInfoHash.reply_count -1
        └─→ Redis IncrBy replyCount -1
```

#### 5.3.3 `handleVoteEvent` — 异步落库

```
收到 VoteEvent
    │
    ├─→ DB 查询 comment + 现有 vote
    ├─→ DB 事务（2 个独立事务！）：
    │   事务1: 更新 comment 计数器 + 创建/更新/删除 vote 记录
    │   事务2: 更新 comment_stat 表
    │
    └─→ 注意：两次独立事务非原子！
```

### 5.4 幂等性保证

```go
// comment_consumer.go
func (c *CommentConsumer) skipIfProcessed(envelope *mq.EventEnvelope) bool {
    key := mq.BuildIdempotencyKey(c.config.KqConsumerConf.Group, envelope)
    // key = "schill:mq:consume:comment-consumer-group:{eventID}"
    ok, err := c.redis.SetNX(ctx, key, "1", mq.DefaultEventTTL) // 24h
    return !ok  // SetNX 返回 false = key 已存在 = 已处理过
}
```

**原理**：利用 Redis `SetNX`（SET if Not eXists）的原子性，以 `group + eventID` 为 key 实现分布式幂等。TTL 为 24 小时。

### 5.5 重试与死信队列

```
消息处理失败
    │
    ├─→ 检查 x-retry-count header
    │   ├─→ < 3: 重新发送到原 topic（header +1）
    │   └─→ >= 3: 发送到 DLQ (comment-dlq)
    │             携带 x-source-topic + x-original-offset
    │
    └─→ 无论成功/失败/重试/DLQ，都 MarkMessage（不阻塞消费）
```

### 5.6 Kafka Producer 的双重实现

项目存在两套 Kafka Producer 实现：

| 实现 | 位置 | 特点 | comment 使用 |
|------|------|------|-------------|
| `common/mq.Producer` | `/workspace/common/mq/producer.go` | `RequiredAcks=WaitForLocal`, 不返回 success | ❌ |
| `common/kafka.Producer` | `/workspace/common/kafka/producer.go` | 支持 Sync/Async, `WaitForAll`(sync)/`WaitForLocal`(async) | ✅ 使用 Async |

comment service 使用的是 `common/kafka.Producer`（async 模式）。`common/mq.Producer` 未被 comment 使用但代码仍存在。

---

## 六、Redis 缓存架构详解

### 6.1 Redis Key 空间设计

| Key 模式 | 类型 | 用途 | TTL |
|----------|------|------|-----|
| `schill:comment:info:{id}` | **Hash** | 评论元数据（id, post_id, user_id, like_count 等） | 600s (10min) |
| `schill:comment:content:{id}` | **String** | 评论内容文本 | 600s |
| `schill:post:comments:{postId}:list` | **ZSet** | 帖子评论列表（time 排序，score=created_at） | 1200s (physical TTL) |
| `schill:post:comments:{postId}:hot` | **ZSet** | 帖子评论列表（hot 排序，score=热度公式） | 1200s |
| `schill:post:comments:meta:{postId}:{sort}` | **String(JSON)** | 缓存元数据 Entry（fresh marker） | logical=600s / physical=1200s |
| `schill:comment:replies:{rootId}:list` | **ZSet** | 某评论的回复列表（time 排序） | 1200s |
| `schill:comment:replies:meta:{rootId}` | **String(JSON)** | 回复缓存元数据 | 600s / 1200s |
| `schill:comment:vote:{commentId}:user:{userId}` | **String** | 用户投票状态（"1"/"2"） | 30 天 |
| `schill:user:vote:count:{userId}:{date}` | **String** | 用户当日投票计数 | 24h |
| `schill:post:comment_count:{postId}` | **String(int64)** | 帖子根评论总数（go-zero cache） | 600s |
| `schill:comment:reply_count:{rootId}` | **String(int64)** | 某评论回复总数 | 600s |
| `schill:comment:lock:post:{postId}:{sort}` | **String** | 缓存重建分布式锁 | 10s |
| `schill:mq:consume:{group}:{eventId}` | **String** | Kafka 幂等标记 | 24h |

### 6.2 缓存策略核心机制

#### 6.2.1 逻辑 TTL vs 物理 TTL（防缓存雪崩）

```
逻辑 TTL (Logical TTL): 600s
  └── Entry.ExpiresAt 字段，用于 IsFresh() 判断

物理 TTL (Physical TTL): 1200s (2x logical)
  └── Redis Key 的实际过期时间

策略：
  - 逻辑 TTL 内：直接使用缓存
  - 逻辑 TTL 过期但物理 TTL 未过期：使用当前缓存 + 异步刷新
  - 物理 TTL 过期：缓存完全失效，触发同步重建
```

**优势**：避免在缓存大批量同时过期时，所有请求同时打到 DB（缓存雪崩）。

#### 6.2.2 SingleFlight 防缓存击穿

```go
var commentListGroup cacheprotect.Group  // 进程级别

func (l *GetCommentListLogic) ensureCommentCache(postId, sortType) {
    flightKey := "comment:list:{postId}:{sortType}"
    commentListGroup.Do(flightKey, func() {
        // 同一进程内，多个并发请求只会执行一次
        // 其他请求等待并共享结果
    })
}
```

**原理**：`syncx.SingleFlight` 保证同一 key 的并发请求中只有一个真正执行，其他等待共享结果。类似 Go 官方 `golang.org/x/sync/singleflight`。

#### 6.2.3 分布式锁防多实例重建

```go
func rebuildCommentCache(postId, sortType) {
    lockKey := "schill:comment:lock:post:{postId}:{sortType}"
    acquired := TryLock(lockKey, 10s)  // Redis SetNX

    if !acquired {
        // 等 1 秒（20次 × 50ms），轮询检查是否已重建
        WaitFor(ctx, 20, 50ms, func() bool {
            return cacheIsFresh()
        })
    }

    // 获得锁：执行 DB 查询 + ZAdd + StoreMarker
    defer ReleaseLock(lockKey)
}
```

**流程**：
1. 多个 comment-rpc 实例同时发现缓存过期
2. 只有一个实例能获取到分布式锁（Redis SetNX）
3. 其他实例进入 `WaitFor` 轮询，检查缓存是否已被重建
4. 获得锁的实例重建完成后释放锁
5. 等待的实例检测到缓存新鲜，直接使用

#### 6.2.4 Pipeline 批量读取

```go
func batchGetCommentInfo(ids []uint64) {
    pipe := Redis.Pipeline()
    // 批量发送 HGetAll 命令（非逐个发送）
    for _, id := range ids {
        cmds[id] = pipe.HGetAll(ctx, "schill:comment:info:{id}")
    }
    pipe.Exec()  // 一次性发送所有命令，减少 RTT
}
```

**原理**：Redis Pipeline 将多个命令打包在一次网络往返中发送，避免 N 次 RTT。注意不是事务（无原子性保证），只是批量传输优化。

#### 6.2.5 回源填缓存（Cache-Aside 模式）

```
1. 读 Redis → miss
2. 读 DB → hit
3. 写 Redis (HMSet/Set) + Expire
4. 返回数据

注意：步骤 2-3 不在锁保护内，可能存在短时间的重复 DB 查询。
     但 SingleFlight 已大幅减少这种情况。
```

### 6.3 缓存失效策略

| 操作 | 失效范围 | 方式 |
|------|---------|------|
| 创建评论 | 帖子评论 ZSet × 2 (time + hot) + meta × 2 + count | Redis DEL |
| 创建回复 | 父评论回复 ZSet + meta + replyCount | Redis DEL |
| 删除评论 | 同上 + 评论 info hash + content | Redis DEL |
| 投票 | 仅更新 hot ZSet 中该评论的 score（`ZAdd` 覆盖，不失效全列表） | Redis ZAdd |
| 自然过期 | 所有缓存 Key 有 TTL | 自动 |

**投票缓存的精细化处理**：投票时不是简单失效整个 hot 列表，而是用 `ZAdd` 更新单个评论的 score（热度值），避免了全量重建。

### 6.4 缓存一致性分析

| 场景 | 一致性窗口 | 影响 |
|------|-----------|------|
| 创建评论 → 列表可见 | 0（缓存已失效，下次请求重建） | 无 |
| 投票 → 列表热度变化 | 0（ZAdd 即时更新 score） | 无 |
| 投票 → 点赞数变化 | 0（Lua 原子更新 info hash） | 无 |
| 删除 → 列表移除 | 0（ZRem 即时移除） | 无 |
| 回复 → 父评论 replyCount | 0（Kafka 消费者 HIncrBy） | 无 |
| DB 计数器 vs Redis 计数器 | 取决于 Kafka 消费延迟（通常 <100ms） | 可接受 |
| 多实例缓存不一致 | TTL 窗口内可能不一致（600s） | 低影响（评论列表对实时性要求不高） |

---

## 七、投票系统设计

### 7.1 整体流程

```
VoteComment RPC
    │
    ├─→ 步骤1: 参数校验 (commentId≠0, userId≠0, voteType∈{0,1,2})
    ├─→ 步骤2: 评论存在性校验 (DB SELECT)
    ├─→ 步骤3: 防刷限流 (Redis Incr, 每日上限 200)
    ├─→ 步骤4: 确保 Redis info hash 存在
    ├─→ 步骤5: Lua 脚本原子更新
    │         ├─ 读取旧投票状态
    │         ├─ 计算 like/dislike 增量
    │         ├─ 更新 voteKey (SET/DEL)
    │         └─ 更新 infoHash (HIncrBy like_count/dislike_count)
    ├─→ 步骤6: Kafka 异步发送 VoteEvent（落库）
    ├─→ 步骤7: 更新 hot ZSet 中的 score
    └─→ 返回结果
```

### 7.2 Lua 脚本详解

```lua
-- KEYS[1]: voteKey  (schill:comment:vote:{commentId}:user:{userId})
-- KEYS[2]: infoKey  (schill:comment:info:{commentId})
-- ARGV[1]: newVote ("0"/"1"/"2")
-- ARGV[2]: expire   (30天)

-- 状态转换矩阵：
-- oldVote \ newVote |  0(取消)  |  1(赞)   |  2(踩)
-- ──────────────────┼───────────┼──────────┼──────────
--      0(无)        |  无变化    | like+1   | dislike+1
--      1(赞)        |  like-1   |  无变化   | like-1, dislike+1
--      2(踩)        | dislike-1 | like+1, dislike-1 | 无变化
```

**为什么用 Lua 脚本？** 投票涉及"读-改-写"三个操作（读旧状态 → 计算增量 → 写新状态），必须原子化。Lua 脚本在 Redis 服务端原子执行，避免了并发竞态。

### 7.3 DB 降级路径

当 Redis Lua 脚本执行失败时（如 Redis 不可用），降级到 DB 事务处理：

```go
func voteCommentDB(in) {
    DB.Transaction(func(tx) {
        // 1. 查询现有投票
        // 2. 根据状态转换更新 comment 计数器
        // 3. 创建/更新/删除 comment_vote 记录
        // 4. Save comment
    })
}
```

### 7.4 防刷限流

```go
date := time.Now().Format("20060102")              // 按天分 key
key := "schill:user:vote:count:{userId}:{date}"
count := Redis.Incr(key)                           // 原子递增
if count == 1 { Redis.Expire(key, 24h) }           // 首次设置 TTL
if count > 200 { return ErrTooManyRequests }       // 超过限制拒绝
```

**分析**：简单的计数器限流，非滑动窗口。每日 0 点重置。200 次/天对于正常用户足够，对刷子有一定限制但非银弹。

---

## 八、ServiceContext 依赖关系

### 8.1 初始化流程

```
NewServiceContext(config)
    │
    ├─→ OpenMySQL (GORM, 连接池 100/20)
    ├─→ AutoMigrate (4 张表)
    ├─→ NewClient(Redis) — 独立的 go-redis 客户端
    ├─→ cache.New(go-zero cache) — 复用同一 Redis 实例
    ├─→ NewAsyncProducer(Kafka) — 异步生产者
    ├─→ NewCommentConsumer(Kafka) — 消费者组
    └─→ NewContentCenter(ContentRpc) — 上游 gRPC 客户端
```

### 8.2 两套 Redis 客户端

| 客户端 | 用途 | 使用位置 |
|--------|------|---------|
| `svcCtx.Redis` (`*commonredis.Client`) | 直接 Redis 操作（ZSet、Pipeline、Lua、Hash） | logic 层 + consumer 层 |
| `svcCtx.Cache` (`cache.Cache`) | go-zero 封装的缓存（自动序列化 + singleflight） | logic 层（comment count / reply count） |

**注意**：两个客户端指向同一 Redis 实例，但使用不同连接池。`svcCtx.Cache` 通过 go-zero 的 `cache.CacheConf` 配置，`svcCtx.Redis` 通过 `RedisConf` 配置。存在**连接数翻倍**的风险（各 10 个连接 = 总共 20 个）。

---

## 九、存在的问题

### 9.1 P0 — 严重问题

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| **P0-1** | **CreateComment/DeleteComment/VoteComment 信任客户端 userId** | createcommentlogic.go, deletecommentlogic.go, votecommentlogic.go | 任意用户可冒用他人身份创建评论、删除他人评论、替他人投票 |
| **P0-2** | **DeleteComment 不执行 DB 写操作** | deletecommentlogic.go | 评论删除依赖 Kafka 消费者异步处理，消息丢失则删除失败。用户收到的成功响应是"假成功" |
| **P0-3** | **handleVoteEvent 两次独立事务非原子** | comment_consumer.go:319-387 | 更新 comment 计数器的事务1成功，但更新 comment_stat 的事务2失败时，数据不一致 |

### 9.2 P1 — 高风险问题

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| **P1-1** | **voteCommentDB 未更新 comment_stat** | votecommentlogic.go:208-297 | 降级路径不更新统计表，comment_stat 与 comment 数据不一致 |
| **P1-2** | **缓存 miss 时的 DB 回源无并发控制** | getcommentlistlogic.go:303-334, getreplylistlogic.go:150-180 | 多个请求同时 cache miss 时，都会去 DB 查询相同的 comment IDs，造成 DB 压力 |
| **P1-3** | **Kafka 生产者 WaitForLocal，非 WaitForAll** | config（kafka.NewAsyncProducer） | Leader 宕机时可能丢消息（虽然概率低） |
| **P1-4** | **缓存重建锁超时后无降级** | getcommentlistlogic.go:229 | WaitFor 超时返回错误，调用方直接降级到 DB 查询，但此时可能另一个实例正在重建中（即将完成），造成重复 DB 查询 |
| **P1-5** | **Lua 脚本每次 Eval，未使用 EvalSha** | votecommentlogic.go:146 | 每次投票都传输完整 Lua 脚本（~1.5KB），未利用 Redis 脚本缓存优化带宽 |

### 9.3 P2 — 中等问题

| # | 问题 | 位置 |
|---|------|------|
| **P2-1** | `CreateComment` 事务内查询父评论在事务外（line 50-61），非原子 | createcommentlogic.go |
| **P2-2** | `batchGetCommentInfo` 和 `batchGetCommentContent` 在 GetCommentList 和 GetReplyList 中重复实现 | 两个 logic 文件 |
| **P2-3** | `cacheReplyCount` 使用 `cacheutil.JitterDefault` 做 TTL 抖动，但 `getTotalCommentCount` 使用固定 TTL | 不一致 |
| **P2-4** | 投票限流使用每日计数而非滑动窗口 | votecommentlogic.go:115-130 |
| **P2-5** | `common/mq.Producer` 和 `common/kafka.Producer` 两套实现并存 | common/ |
| **P2-6** | 消费者使用 `BalanceStrategyRange`（按分区范围），不如 RoundRobin 均匀 | comment_consumer.go:37 |
| **P2-7** | `handleCommentDeletedEvent` 未失效 go-zero cache 层的 comment count / reply count | comment_consumer.go:254-303 |
| **P2-8** | `parseInt32` 忽略了 `strconv.ParseInt` 的错误返回值 | votecommentlogic.go:344-360 |
| **P2-9** | 评论创建事务内 `createdAt = time.Now()` 后手动赋值 `CreatedAt`/`UpdatedAt`，与 GORM `autoCreateTime` 标签冲突 | createcommentlogic.go:71 |

---

## 十、技术方案选型对比

### 10.1 评论列表分页方案

| 方案 | 实现方式 | 优点 | 缺点 | 本项目选择 |
|------|---------|------|------|-----------|
| **游标分页 (Cursor)** | `WHERE id < cursor ORDER BY id DESC` | 性能稳定，不受新增数据影响 | 不能跳页 | ✅ **已采用** |
| **偏移分页 (Offset)** | `LIMIT offset, size` | 实现简单，可跳页 | 大 offset 性能差，数据变动时重复/遗漏 | ❌ |
| **Keyset 分页** | `WHERE (score, id) < (lastScore, lastId)` | 支持多字段排序的精确分页 | 实现复杂 | ❌ |
| **时间线分页** | 按时间范围切分 | 适合 Feed 流 | 需要固定时间窗口 | ❌ |

**本项目实现**：使用 Redis ZSet 的 score 作为游标 (`ZRevRangeByScoreWithScores`)，天然支持游标分页。nextCursor 使用最后一条记录的 score 值。

### 10.2 缓存更新策略

| 策略 | 描述 | 一致性 | 复杂度 | 本项目选择 |
|------|------|--------|--------|-----------|
| **Cache-Aside** | 读缓存 → miss 则读 DB 并回填 | 最终一致 | 低 | ✅ **已采用（读路径）** |
| **Write-Through** | 写 DB 同时写缓存 | 强一致（同步） | 中 | ❌ |
| **Write-Behind** | 先写缓存，异步写 DB | 弱一致 | 高 | ❌ |
| **Write-Invalidate** | 写 DB 后删除缓存 | 最终一致 | 低 | ✅ **已采用（写路径）** |

**本项目组合**：读路径 Cache-Aside + 写路径 Write-Invalidate。创建/删除评论后失效相关 ZSet，投票后通过 ZAdd 局部更新而非全量失效。

### 10.3 缓存防击穿方案

| 方案 | 原理 | 适用场景 | 本项目选择 |
|------|------|---------|-----------|
| **SingleFlight** | 合并同一 key 的并发请求 | 读密集型 | ✅ **已采用（进程级）** |
| **分布式锁** | Redis SetNX 互斥 | 多实例部署 | ✅ **已采用（实例级）** |
| **互斥锁 (Mutex)** | `sync.Mutex` | 单实例 | ❌ |
| **永不过期 + 异步刷新** | 物理 TTL 设很长 | 对实时性要求低的场景 | ❌ |
| **提前预加载** | 定时任务提前刷新热点 key | 可预测的热点 | ❌ |

**本项目组合**：SingleFlight（进程级） + 分布式锁（实例级），形成双层防护。

### 10.4 投票计数方案

| 方案 | 原理 | 并发安全 | 性能 | 本项目选择 |
|------|------|---------|------|-----------|
| **Redis Lua 原子更新** | Lua 脚本在 Redis 内原子执行 | ✅ 原子 | ⭐⭐⭐⭐⭐ | ✅ **已采用（主路径）** |
| **DB 事务** | `SELECT ... FOR UPDATE` + 事务 | ✅ 原子 | ⭐⭐ | ✅ **已采用（降级路径）** |
| **Redis HIncrBy** | 直接递增 Hash 字段 | ❌ 需读旧状态 | ⭐⭐⭐⭐⭐ | ❌ |
| **CAS (Compare-And-Swap)** | WATCH + MULTI/EXEC | ✅ 原子（乐观锁） | ⭐⭐⭐⭐ | ❌ |
| **异步队列** | 所有投票入队列，单线程消费 | ✅ 串行 | ⭐⭐⭐ | ❌ |

**Lua vs CAS 对比**：
- **Lua**：服务端原子执行，无重试开销，但脚本需要传输（~1.5KB）
- **CAS (WATCH)**：客户端乐观锁，冲突时需要重试，但不需要传输脚本

**本项目选择 Lua**，因为投票状态转换涉及 4 种 case（无→赞、无→踩、赞→取消、赞→踩 等），CAS 的重试概率较高。

### 10.5 消息队列可靠性方案

| 方案 | 描述 | 可靠性 | 本项目选择 |
|------|------|--------|-----------|
| **At-Least-Once** | 消费者手动提交 offset，处理成功后才 commit | ⭐⭐⭐⭐ | ✅ **已采用** |
| **At-Most-Once** | 自动提交 offset（先 commit 后处理） | ⭐⭐ | ❌ |
| **Exactly-Once** | Kafka 事务 + 幂等生产者 | ⭐⭐⭐⭐⭐ | ❌ |
| **Outbox Pattern** | 本地事务 + 发件箱表 | ⭐⭐⭐⭐⭐ | ❌ |

**本项目**：At-Least-Once（`session.MarkMessage` 在处理后调用） + 消费者幂等（Redis SetNX） + 重试 3 次 + DLQ。但未使用 Kafka 事务或幂等生产者（`Idempotent: false`）。

### 10.6 Kafka Producer Ack 配置

| 配置 | 描述 | 可靠性 | 延迟 | 本项目选择 |
|------|------|--------|------|-----------|
| **WaitForLocal** | 仅 Leader 确认 | ⭐⭐⭐ | 低 | ✅ **Async Producer** |
| **WaitForAll** | 所有 ISR 确认 | ⭐⭐⭐⭐⭐ | 高 | ✅ **Sync Producer** |
| **NoResponse** | 不等待确认 | ⭐ | 最低 | ❌ |

**本项目**：comment 的 Kafka Producer 使用 Async 模式 + `WaitForLocal`。这意味着 Leader 宕机且未同步到 Follower 时可能丢消息。

### 10.7 软删除方案

| 方案 | 描述 | 优点 | 缺点 | 本项目选择 |
|------|------|------|------|-----------|
| **deleted_at 字段** | `WHERE deleted_at IS NULL` | 可恢复 | 所有查询需加条件 | ✅ **已采用** |
| **status 标记** | status=3 表示删除 | 查询简单 | 与业务状态混在一起 | ✅ **已采用（同时使用）** |
| **独立删除表** | 移入 archive 表 | 主表干净 | 查询历史复杂 | ❌ |
| **真删除** | 物理删除 | 空间节省 | 不可恢复 | ❌ |

**本项目**：同时使用 `deleted_at` 和 `status=3` 双标记。

### 10.8 防刷限流方案

| 方案 | 描述 | 精度 | 内存 | 本项目选择 |
|------|------|------|------|-----------|
| **固定窗口计数器** | 每日重置计数器 | 低（边界问题） | 低 | ✅ **已采用** |
| **滑动窗口** | ZSet + 时间戳 | 高 | 中 | ❌ |
| **令牌桶** | 匀速放令牌 | 高 | 低 | ❌ |
| **漏桶** | 恒定速率处理 | 高 | 低 | ❌ |

**本项目**：使用 Redis Incr 实现的固定窗口计数器，每日 200 次上限。存在边界问题（23:59 和 00:01 可以连续投 400 次）。

---

## 十一、改进建议优先级

### P0 — 立即修复

1. **P0-1: 添加 gRPC 认证拦截器**（与 user service 相同的修复方案）
   - 在 `comment.go` 注册 `AuthUnaryInterceptor`
   - 在 `CreateComment/DeleteComment/VoteComment` 中校验 `in.UserId == authUserId`
   - 网关层传播 JWT token 到 gRPC metadata

2. **P0-2: DeleteComment 同步执行软删除**
   - 在 `deletecommentlogic.go` 的事务中直接执行软删除
   - Kafka 消息仅用于缓存清理（幂等，不影响核心数据）

3. **P0-3: 合并 handleVoteEvent 的两个事务**
   - 将 comment_stat 更新合并到 comment + comment_vote 事务中

### P1 — 近期修复

4. **P1-1: voteCommentDB 补全 comment_stat 更新**
5. **P1-2: 缓存回源加 SingleFlight 保护**
6. **P1-3: 评估是否需要提高 Kafka Ack 级别**
7. **P1-4: WaitFor 超时后先查询一次缓存再降级 DB**
8. **P1-5: 使用 SCRIPT LOAD + EvalSha 优化 Lua 脚本执行**

### P2 — 计划优化

9. **P2-1: 将父评论查询移入事务**
10. **P2-2: 提取公共的 batchGetCommentInfo/Content 方法**
11. **P2-3: 统一 TTL 抖动策略**
12. **P2-4: 考虑滑动窗口限流**
13. **P2-5: 清理废弃的 common/mq.Producer**
14. **P2-6: 消费者 Rebalance 策略改为 RoundRobin**
15. **P2-7: 删除时失效 go-zero cache 层**
16. **P2-8: parseInt32 处理错误**
17. **P2-9: 移除重复的 CreatedAt/UpdatedAt 手动赋值**

---

## 十二、服务依赖图

```
                          ┌─────────────────┐
                          │   API Gateway    │
                          │   (HTTP :8888)   │
                          └────────┬────────┘
                                   │ gRPC
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                     comment.rpc (:8083)                          │
│                                                                  │
│  ┌───────────────────┐     ┌──────────────────────────────────┐  │
│  │   gRPC Server     │     │   CommentConsumer                │  │
│  │                   │     │   (Kafka Consumer Group)         │  │
│  │  CreateComment ───┼────→│   consume: comment-create        │  │
│  │  DeleteComment ───┼────→│   consume: comment-deleted       │  │
│  │  GetCommentList   │     │   consume: comment-vote          │  │
│  │  GetReplyList     │     │   produce: comment-dlq (DLQ)     │  │
│  │  VoteComment ─────┼────→│                                  │  │
│  └────────┬──────────┘     └──────────────┬───────────────────┘  │
│           │                               │                      │
│           │                               │                      │
└───────────┼───────────────────────────────┼──────────────────────┘
            │                               │
            ▼                               ▼
    ┌──────────────┐              ┌──────────────────┐
    │   MySQL:3306 │              │   Redis :6379    │
    │   schill DB  │              │                  │
    │   - comment  │              │  - ZSet 分页列表  │
    │   - content  │              │  - Hash 评论信息  │
    │   - vote     │              │  - String 内容    │
    │   - stat     │              │  - 投票状态       │
    └──────────────┘              │  - 限流计数       │
                                  │  - 分布式锁       │
            ┌─────────────────────│  - 幂等标记       │
            │                     └──────────────────┘
            ▼
    ┌──────────────┐
    │ Content RPC  │
    │ (:8082)      │
    │ 验证帖子存在  │
    └──────────────┘
```

---

> **总结**：comment service 是项目中**最复杂**的服务之一，涉及多层缓存（ZSet + Hash + String）、异步事件驱动（Kafka 生产/消费同一进程）、Lua 原子操作、分布式锁 + SingleFlight 双重防击穿。核心架构设计合理，但存在认证缺失和异步删除可靠性等关键问题需要修复。
