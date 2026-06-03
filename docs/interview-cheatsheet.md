# SChill 项目面试速查表

## 一句话概括

**SChill** 是一个基于 go-zero 的 Go 微服务内容社区平台，采用 gRPC + Kafka + ES + Canal 实现事件驱动、CQRS 读写分离、最终一致性的架构。

---

## 架构关键词

| 概念 | 项目实现 |
|------|---------|
| 微服务框架 | go-zero (RPC + HTTP Gateway + 服务治理) |
| 通信协议 | gRPC (Protobuf) 内部通信，REST (HTTP) 对外 |
| 服务发现 | Etcd |
| 事件驱动 | Kafka (7 个业务 topic + DLQ) |
| CQRS | MySQL 写，ES 读（搜索），Canal CDC 同步 |
| 缓存策略 | Redis 多层缓存 + 本地内存缓存 + 缓存防护 |
| 最终一致性 | Kafka 事件异步更新跨服务计数器 |
| 认证鉴权 | JWT (Access + Refresh Token) |
| 部署方案 | Docker Compose / K8s (Kustomize) |
| 可观测性 | OpenTelemetry + Zipkin + Prometheus + Pyroscope |

---

## 面试常见问题速答

### Q1: 为什么用 go-zero？

- 开箱即用的微服务全家桶：RPC 框架 + HTTP 网关 + 服务注册发现 + 负载均衡 + 熔断限流 + 链路追踪
- 通过 goctl 从 proto 自动生成代码，开发效率高
- 内置中间件生态（breaker、limiter、timeout、trace）
- 适合中小团队快速搭建微服务体系

### Q2: 为什么用 gRPC 而不是 REST 做内部通信？

- Protobuf 二进制序列化，比 JSON 体积小 3-10x，解析快
- 基于 HTTP/2，支持多路复用、双向流
- 强类型契约（proto 文件），自动生成客户端/服务端代码
- 内置超时、重试、负载均衡（go-zero 封装）

### Q3: Feed 流怎么实现的？

Feed 服务聚合多服务数据：
1. 从 Content 服务获取帖子列表
2. 从 User 服务获取作者信息
3. 从 Interaction 服务获取当前用户点赞/收藏状态
4. 从 Relation 服务获取关注状态
5. 组装返回给客户端

优化：Redis 缓存聚合结果（10-30s TTL），本地内存缓存热点数据。

### Q4: 帖子点赞数/评论数怎么保证一致性？

**不保证强一致性，采用最终一致性。完整链路：**

```
用户点 Star
  → Redis Lua 原子执行 SADD + INCR（同步，< 1ms 返回）
  → go func() 异步发一条 Kafka 消息（不阻塞 HTTP 响应）
  → Kafka AsyncProducer 层：Snappy 压缩 + 10ms/100条 攒批刷新
  → Content 服务逐条消费，执行 UPDATE post SET upvote_count = upvote_count + 1
  → 读帖子时优先读 Redis 计数，降级读 MySQL
```

**关于压缩和攒批：**

| 问题 | 答案 |
|------|------|
| Kafka 是否压缩？ | **是** — AsyncProducer 使用 Snappy 压缩，减少网络带宽 |
| 是否逐条发？ | **应用层是逐条发**（每个操作一个 `SendEvent`），但 **Sarama 客户端层会攒批**（每 10ms 或攒够 100 条 flush 一次） |
| 需要自己弄时间窗口吗？ | **不需要** — Sarama AsyncProducer 自带 `Flush.Frequency=10ms` + `Flush.Messages=100`，网络层自动批量化 |
| 消费端是否批量？ | Content 服务 **逐条消费逐条 UPDATE**（每次一条 SQL）；Interaction 服务自身消费者做了**内存攒批**（100 条/500ms 批量 INSERT） |

**为什么不做应用层攒批（比如 5s 窗口合并）？**
- **实时性要求**：计数更新需要尽快反映到前端，5s 延迟太慢
- **已经够用**：AsyncProducer 的 10ms 攒批 + Snappy 压缩已经大幅降低了网络开销
- **计数操作是幂等的**：`upvote_count = upvote_count + 1` 不怕重复，不需要合并
- **如果真做合并**：需要在 Interaction 服务加本地 Map 聚合 `{post_id: delta}`，定时 flush → 引入内存开销和丢失风险，收益不大

### Q5: 搜索怎么实现？

**CQRS 模式：**
- MySQL 是写模型，存储完整业务数据
- Elasticsearch 是读模型，存储搜索优化的文档
- Canal 伪装成 MySQL Slave 捕获 binlog，发到 Kafka
- canal-sync 服务消费 Kafka 消息，同步到 ES
- 延迟通常在 100ms-1s

### Q6: 怎么防止缓存穿透/击穿/雪崩？

| 问题 | 方案 |
|------|------|
| 穿透 | 缓存空值（nil marker），短 TTL |
| 击穿 | 分布式锁（cacheprotect.TryLock），同一时刻只有一个线程重建缓存 |
| 雪崩 | TTL 加随机 jitter（cacheutil），避免同时过期 |

### Q7: Kafka 怎么保证消息不丢/不重？

**不丢（可靠性）：**
- **Producer 端**：AsyncProducer `RequiredAcks=WaitForAll`，等所有 ISR 副本确认；SyncProducer 同理
- **Broker 端**：默认 replication factor=1（单节点开发环境），生产应 ≥ 3
- **Consumer 端**：手动提交 offset（`session.MarkMessage`），处理成功才标记

**不重（幂等性）：**
- 每条消息有唯一 `EventID`（`Envelope.EventID`）
- Consumer 处理前通过 `Redis SETNX` 标记（key: `schill:idempotent:{groupId}:{eventId}`，TTL 24h）
- ⚠️ 当前项目存在**幂等标记时机问题**：标记在处理前，处理失败则消息丢失（审查中已指出）

**Producer 压缩与攒批：**
| 配置项 | AsyncProducer | SyncProducer |
|--------|--------------|--------------|
| 压缩算法 | **Snappy** | 无 |
| 攒批策略 | 10ms / 100条 | 无（逐条同步等待） |
| 使用方 | interaction, comment | content, relation |

**失败处理：**
- comment consumer：重试 3 次 → DLQ（`comment-dlq`）
- interaction consumer：重试 → DLQ（`interaction-dlq`）
- content consumer（通用框架）：不重试，不标记，等待 rebalance 重消费

### Q8: 分布式事务怎么处理？

**不使用分布式事务，采用 Saga 模式 + 最终一致性：**
1. 本地事务完成（如创建帖子）
2. 发 Kafka 事件
3. 下游服务异步消费（如更新用户发帖数）
4. 失败通过 DLQ + 补偿处理

当前不足：缺少 Transactional Outbox 保证 DB 写入和消息发送的原子性。

### Q9: 服务怎么扩容？

- 无状态服务（gateway、feed、search）：直接增加副本数
- 有状态服务（user、content、comment）：数据库是瓶颈，需要读写分离、分库分表
- Kafka：增加 partition 数（需要提前规划）
- 通过 go-zero + Etcd 自动服务发现，新实例自动注册

### Q10: 项目有什么可以改进的地方？

1. **安全**：密码从 SHA-256 升级到 bcrypt；基础设施加认证；授权从 JWT context 获取用户 ID
2. **可靠性**：Kafka 幂等性标记移到处理成功后；加 Transactional Outbox；分布式锁加所有权验证
3. **性能**：Feed 聚合并行化（errgroup）；加数据库复合索引；解决 N+1 查询
4. **运维**：K8s 用 PVC 替代 hostPath；加 HPA 自动扩缩；完善监控告警

---

## 数据库核心表

```
user ──→ user_profile (1:1)
user ──→ user_stat (1:1)
user ──→ following (1:N, 关注关系)
user ──→ post (1:N)
post ──→ post_content (1:N, 多段内容)
post ──→ post_topic (N:M) ──→ topic
post ──→ post_star (N:M)
post ──→ post_collection (N:M)
post ──→ comment (1:N)
comment ──→ comment (N:1, parent 嵌套)
comment ──→ comment_content (1:1)
comment ──→ comment_vote (N:M)
comment ──→ comment_stat (1:1)
```

---

## Redis Key 设计

```
schill:post:base:{id}          // 帖子基础信息（长 TTL）
schill:post:stats:{id}         // 帖子统计信息（短 TTL）
schill:post:like:set:{id}      // 点赞用户集合
schill:post:like:count:{id}    // 点赞计数
schill:post:collect:set:{id}   // 收藏用户集合
schill:post:collect:count:{id} // 收藏计数
schill:comment:{id}:info       // 评论信息
schill:comment:{id}:content    // 评论内容
schill:comment:list:{postId}   // 帖子评论列表（ZSet）
schill:feed:{type}:user:{uid}:page:{n}:size:{m}  // Feed 缓存
schill:idempotent:{groupId}:{eventId}  // 幂等标记
```

---

## Kafka Topic 设计

| Topic | 生产者 | 消费者 | 用途 |
|-------|--------|--------|------|
| post-created | content | user | 更新发帖数 |
| post-deleted | content | user | 更新发帖数 |
| comment-created | comment | content | 更新评论数 |
| comment-deleted | comment | content | 更新评论数 |
| user-followed | relation | user | 更新关注统计 |
| user-unfollowed | relation | user | 更新关注统计 |
| post-star | interaction | content | 更新点赞数 |
| post-collect | interaction | content | 更新收藏数 |
| canal_topic | Canal | canal-sync | ES 同步 |

---

## 服务端口映射

| 服务 | 端口 | 协议 |
|------|------|------|
| gateway | 8086 | HTTP |
| user-rpc | 8080 | gRPC |
| relation-rpc | 8081 | gRPC |
| content-rpc | 8082 | gRPC |
| comment-rpc | 8083 | gRPC |
| interaction-rpc | 8084 | gRPC |
| feed-rpc | 8087 | gRPC |
| search-rpc | 8088 | gRPC |
| canal-sync | 8089 | (内部) |
| Canal Server | 11111 | MySQL 协议 |
