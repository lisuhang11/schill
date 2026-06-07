**# Kafka 面试题深度文档**

\> 基于 SChill 社交平台微服务项目的 Kafka 实战经验整理

\>

\> 技术栈：Go + IBM/sarama v1.47.0 + Kafka 3.9.1 (Kraft 模式)

**---**

**## 目录**

\1. [基础概念篇](#一基础概念篇)

\2. [Producer 篇](#二producer-篇)

\3. [Consumer 篇](#三consumer-篇)

\4. [消息可靠性篇](#四消息可靠性篇)

\5. [幂等性与去重篇](#五幂等性与去重篇)

\6. [消息顺序性篇](#六消息顺序性篇)

\7. [批量处理篇](#七批量处理篇)

\8. [死信队列篇](#八死信队列篇)

\9. [重试机制篇](#九重试机制篇)

\10. [监控与运维篇](#十监控与运维篇)

\11. [CDC 与 Canal 篇](#十一cdc-与-canal-篇)

\12. [架构设计篇](#十二架构设计篇)

\13. [综合场景题](#十三综合场景题)

**---**

**## 一、基础概念篇**

**### Q1: Kafka 的核心架构组件有哪些？各自的作用是什么？**

***\*回答要点：\****

| 组件 | 作用 |

|------|------|

| ***\*Broker\**** | Kafka 服务节点，负责消息存储、读写请求处理 |

| ***\*Topic\**** | 消息的逻辑分类，类似数据库的表 |

| ***\*Partition\**** | Topic 的物理分片，是实现水平扩展和并行消费的基础 |

| ***\*Producer\**** | 消息生产者，将消息发送到指定 Topic |

| ***\*Consumer\**** | 消息消费者，从 Topic 拉取消息 |

| ***\*Consumer Group\**** | 消费者组，组内每个 Partition 只能被一个 Consumer 消费，实现负载均衡 |

| ***\*ZooKeeper / Kraft\**** | 集群元数据管理、Controller 选举（本项目使用 Kraft 模式，无需 ZooKeeper） |

***\*项目实践：\**** 本项目有 15 个 Topic，分布在 6 个微服务中。Producer 和 Consumer 都使用 Sarama 库实现，Kafka 以单节点 Kraft 模式部署在 K8s 中。

**---**

**### Q2: Kafka 为什么快？从存储、网络、消费模型三个角度分析。**

***\*回答要点：\****

\1. ***\*顺序写入（存储）\****：Kafka 采用顺序追加写入磁盘的方式，避免了随机 I/O。现代操作系统的页缓存（Page Cache）机制让顺序写入速度接近内存速度。

\2. ***\*零拷贝（网络）\****：Kafka 使用 `sendfile()` 系统调用，数据从磁盘页缓存直接传输到网卡缓冲区，绕过用户态，减少 CPU 拷贝和上下文切换。

\3. ***\*批量与压缩（消费模型）\****：

   \- Producer 端批量发送（如本项目 `Flush.Messages = 100`）

   \- 使用 Snappy/LZ4 等压缩算法减少网络传输

   \- Consumer 端批量拉取（`Fetch.Min.Bytes`）

***\*项目实践：\**** 本项目 AsyncProducer 配置了 Snappy 压缩、100 条消息批量刷盘、10ms 刷盘间隔。

**---**

**### Q3: 什么是 Consumer Group？Rebalance 机制是什么？**

***\*回答要点：\****

***\*Consumer Group：\****

\- 同一 Group 内的 Consumer 共同消费一个 Topic，每个 Partition 只分配给一个 Consumer

\- 不同 Group 独立消费，互不影响（类似发布-订阅模式）

***\*Rebalance（再平衡）触发条件：\****

\- Consumer 加入或退出 Group

\- Topic Partition 数量变化

\- Consumer 长时间未发送心跳（`session.timeout.ms`）

***\*Rebalance 过程（Eager 协议）：\****

\1. Coordinator 通知所有 Consumer 停止消费（Revoke）

\2. 重新分配 Partition

\3. Consumer 重新开始消费（Assign）

***\*Rebalance 策略：\****

\- ***\*Range\****（默认）：按 Topic 粒度分配，可能导致倾斜

\- ***\*RoundRobin\****：轮询分配，更均匀

\- ***\*Sticky\****：尽量保持原有分配不变

\- ***\*Cooperative Sticky\****：增量 Rebalance，减少暂停时间

***\*项目实践：\**** 本项目中 `common/kafka/consumer.go` 使用 `BalanceStrategyRoundRobin`，comment-rpc 的自定义消费者使用 `BalanceStrategyRange`。

**---**

**### Q4: Topic 和 Partition 的设计原则是什么？**

***\*回答要点：\****

\1. ***\*Partition 数量的考量：\****

   \- Partition 越多，并行度越高，吞吐量越大

   \- 但每个 Partition 会占用 Broker 的文件句柄和内存

   \- Partition 数量应 >= Consumer 数量（同一 Group 内）

\2. ***\*Partition 数量的确定：\****

   \- 预估吞吐量：`所需吞吐量 / 单 Partition 吞吐量`

   \- 考虑未来扩容：预留 20%~50% 余量

   \- 上限：每个 Broker 不超过 4000 个 Partition（含副本）

\3. ***\*Partition 数量只能增加，不能减少\****（Kafka 不支持减少 Partition）

\4. ***\*Key 与 Partition 的映射：\**** `hash(key) % partition_count`，相同 Key 的消息进入同一 Partition，保证顺序。

***\*项目实践：\**** 本项目多数 Topic 使用默认 Partition（1~3 个），与 K8s 单节点 Kraft 部署（Replication Factor = 1）相匹配。生产环境应根据吞吐量调整。

**---**

**### Q5: Kafka 中 Zookeeper 和 Kraft 分别是什么？有什么区别？**

***\*回答要点：\****

| 对比维度 | ZooKeeper 模式 | Kraft 模式 |

|---------|---------------|-----------|

| 元数据存储 | 外部 ZooKeeper 集群 | Kafka 内部 Raft 协议 |

| 部署复杂度 | 需要额外部署 ZK 集群 | 仅部署 Kafka |

| 运维成本 | 维护两套系统 | 单一系统 |

| Controller 选举 | 依赖 ZK | Raft 共识 |

| 成熟度 | 久经考验 | Kafka 2.8+ 支持，3.3+ 生产可用 |

| 性能 | Controller 故障恢复较慢 | 元数据操作更快 |

***\*项目实践：\**** 本项目使用 Kafka 3.9.1 Kraft 模式部署，`infra-kafka.yaml` 中配置了 `KAFKA_NODE_ID: 1` 和 `controller.quorum.voters`。

**---**

**## 二、Producer 篇**

**### Q6: Kafka Producer 的发送流程是怎样的？**

***\*回答要点：\****

\```

消息 → 拦截器 → 序列化器 → 分区器 → RecordAccumulator（缓冲区）

  → Sender 线程 → Broker → 返回 ACK

\```

***\*核心步骤：\****

\1. ***\*拦截器（Interceptor）\****：消息发送前的预处理（可选）

\2. ***\*序列化器（Serializer）\****：将 Java 对象/Go struct 转为字节数组

\3. ***\*分区器（Partitioner）\****：决定消息发往哪个 Partition

   \- 有 Key：`hash(key) % partition_count`

   \- 无 Key：轮询或随机

\4. ***\*RecordAccumulator\****：消息在内存中的缓冲区，每个 Partition 一个 `ProducerBatch`

\5. ***\*Sender 线程\****：独立线程，满足 `batch.size` 或 `linger.ms` 时批量发送

\6. ***\*Broker 响应\****：根据 `acks` 配置决定何时返回确认

***\*项目实践：\**** Sarama 库的 AsyncProducer 内部也有类似的缓冲和批量发送机制。本项目配置 `Flush.Messages = 100`、`Flush.Frequency = 10ms`。

**---**

**### Q7: Kafka Producer 的 acks 参数有哪几种？如何选择？**

***\*回答要点：\****

| acks 值 | 含义 | 可靠性 | 延迟 | 适用场景 |

|---------|------|--------|------|---------|

| ***\*0\**** | 不等待任何确认 | 最低（可能丢消息） | 最低 | 日志收集、监控数据 |

| ***\*1\**** | Leader 写入成功即返回 | 中等（Leader 宕机可能丢） | 中等 | 一般业务，允许少量丢失 |

| ***\*-1 (all)\**** | 所有 ISR 副本都写入成功 | 最高 | 最高 | 金融交易、订单、支付 |

***\*项目实践：\****

\- `SyncProducer`：`RequiredAcks = WaitForAll`（`acks=all`），用于关键业务（content-rpc、relation-rpc）

\- `AsyncProducer`：`RequiredAcks = WaitForLocal`（`acks=1`），用于高吞吐场景（interaction-rpc、comment-rpc）

***\*面试追问：\**** acks=all 一定能保证不丢消息吗？

\- 不一定。如果 ISR 中只有 Leader 存活（`min.insync.replicas=1`），Leader 写入成功就返回了，此时若 Leader 宕机且消息未同步到 Follower，消息仍会丢失。

\- 解决：设置 `min.insync.replicas >= 2`，同时 `acks=all`。

**---**

**### Q8: Sync Producer 和 Async Producer 的区别？各自适用什么场景？**

***\*回答要点：\****

| 维度 | Sync Producer | Async Producer |

|------|--------------|----------------|

| 发送方式 | 阻塞等待 Broker 确认 | 非阻塞，后台异步发送 |

| 吞吐量 | 较低（受 RTT 限制） | 高（批量+异步） |

| 延迟 | 单条延迟高 | 单条延迟低 |

| 可靠性 | 可立即获知结果 | 通过回调/Channel 获知 |

| 实现方式 | `SendMessage()` 返回响应 | 写入 Input Channel，Success/Error Channel 通知 |

***\*适用场景：\****

\- ***\*Sync\****：对可靠性要求高、需要立即确认结果、吞吐量不大的场景

\- ***\*Async\****：高吞吐量、对延迟敏感、允许异步处理结果的场景

***\*项目实践：\****

\```go

// common/kafka/producer.go

// SyncProducer：acks=all，5次重试

func NewSyncProducer(brokers []string, opts ...SyncProducerOption) (*SyncProducer, error) {

​    config.Producer.RequiredAcks = sarama.WaitForAll

​    config.Producer.Retry.Max = 5

​    config.Producer.Return.Successes = true  // 同步需要返回

}

// AsyncProducer：acks=1，Snappy 压缩，批量发送

func NewAsyncProducer(brokers []string, opts ...AsyncProducerOption) (*AsyncProducer, error) {

​    config.Producer.RequiredAcks = sarama.WaitForLocal

​    config.Producer.Retry.Max = 3

​    config.Producer.Compression = sarama.CompressionSnappy

​    config.Producer.Flush.Messages = 100

​    config.Producer.Flush.Frequency = 10 * time.Millisecond

}

\```

**---**

**### Q9: Producer 端如何实现消息压缩？有哪些压缩算法可选？**

***\*回答要点：\****

| 压缩算法 | 压缩率 | CPU 开销 | 适用场景 |

|---------|--------|---------|---------|

| ***\*None\**** | 0% | 无 | 延迟极度敏感 |

| ***\*Snappy\**** | 中等 | 低 | 平衡延迟与吞吐 |

| ***\*LZ4\**** | 中等 | 极低 | 高吞吐，低 CPU |

| ***\*GZIP\**** | 高 | 高 | 带宽敏感 |

| ***\*Zstd\**** | 最高 | 中等 | 压缩率优先（2.1+） |

***\*项目实践：\**** 本项目 AsyncProducer 使用 `sarama.CompressionSnappy`，在延迟和压缩率之间取得平衡。

**---**

**### Q10: Producer 如何选择 Partition？自定义分区策略如何实现？**

***\*回答要点：\****

***\*默认分区策略：\****

\1. 如果消息指定了 Partition → 直接使用

\2. 如果指定了 Key → `murmur2(key) % partition_count`

\3. 无 Key → 粘性分区（Sticky Partitioner），将消息批量发往同一 Partition，减少网络请求

***\*为什么要指定 Key：\****

\- 保证相同 Key 的消息进入同一 Partition（有序性）

\- 关联数据聚合处理

***\*项目实践：\****

\```go

// 基于聚合 ID 设置 Key，保证同一实体的消息有序

producer.SendEvent(

​    "post-created",

​    fmt.Sprintf("%d", postID),  // Key = postID

​    "post.created", "content-rpc", "post", fmt.Sprintf("%d", postID),

​    payload,

)

// relation-rpc：使用复合 Key

func buildRelationKey(userID, targetUserID int64) string {

​    return fmt.Sprintf("%d:%d", userID, targetUserID)

}

\```

**---**

**### Q11: Producer 端有哪些性能优化手段？**

***\*回答要点：\****

\1. ***\*批量发送\****：`batch.size`（如 16KB）、`linger.ms`（如 10ms）

\2. ***\*压缩\****：Snappy/LZ4 减少网络传输

\3. ***\*缓冲区\****：`buffer.memory`（默认 32MB），避免阻塞

\4. ***\*异步发送\****：不阻塞业务线程

\5. ***\*连接复用\****：Sarama Client 共享 TCP 连接

\6. ***\*****`max.in.flight.requests.per.connection`*****\***：调大可提高吞吐，但需考虑顺序性（设为 1 可保证严格顺序）

***\*项目实践：\****

\```go

// AsyncProducer 的性能配置

config.Producer.Flush.Messages = 100    // 100条批量刷盘

config.Producer.Flush.Frequency = *10ms*  // 10ms 定时刷盘

config.Producer.Compression = sarama.CompressionSnappy

config.Producer.MaxMessageBytes = *1MB*   // 最大消息 1MB

\```

**---**

**## 三、Consumer 篇**

**### Q12: Consumer 的消费模型是怎样的？Push vs Pull？**

***\*回答要点：\****

Kafka 采用 ***\*Pull 模型\****，Consumer 主动从 Broker 拉取数据。

***\*为什么是 Pull 而非 Push？\****

\1. ***\*消费速率由 Consumer 控制\****：避免 Push 模式下 Consumer 被压垮

\2. ***\*批量拉取\****：Consumer 可以一次拉取多条消息，提高吞吐

\3. ***\*简化 Broker\****：Broker 不需要维护 Consumer 状态和推送队列

***\*劣势：\**** 如果 Broker 没有数据，Consumer 会空轮询。通过 `fetch.min.bytes` 和 `fetch.max.wait.ms` 优化（长轮询）。

***\*项目实践：\****

\```go

// common/kafka/consumer.go 中的消费循环

func (c *Consumer) Start() {

​    for {

​        select {

​        case <-c.ctx.Done():

​            return

​        default:

​            err := c.group.Consume(c.ctx, c.topics, c.handler)

​            if err != nil {

​                logx.Errorf("Consumer error: %v", err)

​                time.Sleep(time.Second)  // 重连间隔

​            }

​        }

​    }

}

\```

**---**

**### Q13: Consumer 的 Offset 管理有哪几种方式？自动提交和手动提交的区别？**

***\*回答要点：\****

| 方式 | 含义 | 优点 | 缺点 |

|------|------|------|------|

| ***\*自动提交\**** | Consumer 定期自动提交 Offset | 简单 | 可能重复消费、可能丢消息 |

| ***\*手动同步提交\**** | 业务处理后调用 `CommitSync()` | 可控 | 阻塞，影响吞吐 |

| ***\*手动异步提交\**** | 业务处理后调用 `CommitAsync()` | 高吞吐 | 提交失败无法重试 |

***\*提交时机的影响：\****

\- ***\*处理前提交（At-most-once）\****：可能丢消息

\- ***\*处理后提交（At-least-once）\****：可能重复消费

\- ***\*事务提交（Exactly-once）\****：配合幂等性实现精确一次

***\*项目实践：\**** 本项目使用 Sarama 默认行为——***\*消息处理成功后自动标记\****（处理后提交），实现 At-least-once 语义。结合 Redis 幂等性去重实现有效 Exactly-once。

\```go

// 如果 handler 返回 error，消息不会被标记

if err := c.handler.Handle(ctx, msg); err != nil {

​    logx.Errorf("handler error: %v", err)

​    return err  // 不提交 offset，重启后重新消费

}

\```

**---**

**### Q14: Consumer 的** **`session.timeout.ms`** **和** **`max.poll.interval.ms`** **有什么区别？**

***\*回答要点：\****

| 参数 | 含义 | 超时后果 | 默认值 |

|------|------|---------|--------|

| `session.timeout.ms` | Consumer 与 Coordinator 的心跳超时 | 触发 Rebalance，踢出 Group | 45s |

| `max.poll.interval.ms` | 两次 `poll()` 之间的最大间隔 | 触发 Rebalance，认为 Consumer 已死 | 5min |

***\*常见问题：\**** 如果消息处理逻辑耗时超过 `max.poll.interval.ms`，Consumer 会被踢出 Group，导致无限 Rebalance。

***\*解决方案：\****

\1. 增大 `max.poll.interval.ms`

\2. 缩短消息处理时间（异步化、批量优化）

\3. 降低单次拉取的消息数（`max.poll.records`）

**---**

**### Q15: 如何解决 Consumer 消费积压（Lag）问题？**

***\*回答要点：\****

***\*排查步骤：\****

\1. 确认积压量：`kafka-consumer-groups --describe`

\2. 定位瓶颈：是 Producer 写入太快还是 Consumer 处理太慢？

***\*解决方案：\****

| 方案 | 说明 | 适用场景 |

|------|------|---------|

| ***\*增加 Consumer 实例\**** | 扩容同 Group Consumer | Consumer 处理慢 |

| ***\*增加 Partition\**** | 提高并行度上限 | Partition 不足 |

| ***\*优化处理逻辑\**** | 减少 DB 查询、增加缓存、批量写入 | Consumer 处理逻辑重 |

| ***\*异步化处理\**** | 消息投递到内部队列，Worker 池异步处理 | 允许最终一致性 |

| ***\*限流 Producer\**** | 控制生产速度 | 下游确实无法承受 |

| ***\*跳过非关键消息\**** | 降级策略 | 临时应对突发流量 |

***\*项目实践：\**** 本项目的 `common/kafka/idempotency.go` 中有 Consumer Lag 监控：

\```go

// 记录 Consumer Lag 日志（慢消费告警）

func logConsumerLag(partition int32, highWaterMark, offset int64) {

​    lag := highWaterMark - offset

​    if lag > 30 {  // 超过30秒视为慢消费

​        logx.Slowf("Consumer lag is high, partition: %d, lag: %d", partition, lag)

​    }

}

\```

**---**

**### Q16: Consumer Rebalance 的痛点是什么？如何优化？**

***\*回答要点：\****

***\*痛点：\****

\1. ***\*Stop-The-World\****：Eager 协议下，Rebalance 期间所有 Consumer 停止消费

\2. ***\*重复消费\****：Rebalance 前已处理但未提交的消息会被重新消费

\3. ***\*频繁 Rebalance\****：心跳超时或处理超时导致恶性循环

***\*优化方案：\****

\1. ***\*使用 Cooperative Sticky 协议\****（Kafka 2.4+）：增量 Rebalance，不需要全部暂停

\2. ***\*合理设置超时参数\****：

   \- `session.timeout.ms` 不要太小（网络抖动导致误判）

   \- `max.poll.interval.ms` 根据实际处理时间调整

\3. ***\*Consumer 优雅退出\****：应用关闭时主动 `LeaveGroup()`

\4. ***\*避免长事务\****：消息处理时间控制在 `max.poll.interval.ms` 以内

***\*项目实践：\**** 本项目的 Consumer 在收到 `ctx.Done()` 信号后，会等待当前消息处理完成再退出，实现优雅关闭：

\```go

func (c *Consumer) Stop() {

​    c.cancel()  // 触发 ctx.Done()

​    c.group.Close()

}

\```

**---**

**## 四、消息可靠性篇**

**### Q17: Kafka 如何保证消息不丢失？从 Producer、Broker、Consumer 三端分析。**

***\*回答要点：\****

***\*Producer 端：\****

\- `acks = all`（或 -1）：等待所有 ISR 副本确认

\- `retries > 0`：发送失败自动重试（如本项目 SyncProducer 的 5 次）

\- `max.in.flight.requests.per.connection = 1`：保证重试时的顺序性

\- 处理回调中的异常（AsyncProducer 的 Error Channel）

***\*Broker 端：\****

\- `replication.factor >= 3`：多副本冗余

\- `min.insync.replicas >= 2`：最小同步副本数

\- `unclean.leader.election.enable = false`：不允许非 ISR 副本竞选 Leader

\- 日志刷盘策略：`log.flush.interval.messages` 和 `log.flush.interval.ms`

***\*Consumer 端：\****

\- 手动提交 Offset（处理成功后再提交）

\- 消费逻辑需要幂等处理（防止重复消费导致的数据错误）

\- 先处理业务，再提交 Offset

***\*项目实践：\****

| 层次 | 本项目措施 |

|------|-----------|

| Producer | SyncProducer: acks=all, retries=5; AsyncProducer: acks=1, retries=3 |

| Producer 发送 | 异步 goroutine + 错误日志，不阻塞业务 |

| Consumer Offset | 处理后标记（at-least-once） |

| Consumer 幂等 | Redis SETNX 去重 |

| 优雅关闭 | ctx.Done() 信号 + group.Close() |

**---**

**### Q18: 什么是 ISR？什么情况下会出现 ISR 收缩？**

***\*回答要点：\****

***\*ISR（In-Sync Replicas）：\**** 与 Leader 保持同步的副本集合。Follower 需要满足：

\- 与 Leader 的落后时间不超过 `replica.lag.time.max.ms`（默认 30s）

\- 与 Leader 的消息数量差距不超过 `replica.lag.max.messages`（已废弃）

***\*ISR 收缩场景：\****

\1. Follower 宕机或网络故障

\2. Follower 消费速度跟不上 Leader（高吞吐场景）

\3. Follower GC 停顿导致同步中断

***\*影响：\****

\- 如果 `min.insync.replicas > 当前ISR数量`，Producer 写入被拒绝（`NotEnoughReplicas`）

\- 可用副本减少，容灾能力下降

***\*项目实践：\**** 本项目使用单节点 Kraft 部署，`replication.factor = 1`，ISR 始终为 1。生产环境应至少配置 3 副本。

**---**

**### Q19: Kafka 的幂等 Producer 和事务 Producer 是什么？有什么区别？**

***\*回答要点：\****

***\*幂等 Producer（****`enable.idempotence = true`****）：\****

\- 保证单 Partition 内的精确一次语义

\- 原理：Producer 分配 PID（Producer ID），每条消息带 Sequence Number，Broker 根据 `<PID, TopicPartition, SeqNum>` 去重

\- 限制：仅单 Partition、单会话

***\*事务 Producer（****`transactional.id`****）：\****

\- 保证跨 Partition、跨会话的精确一次语义

\- 原理：基于幂等 Producer + 事务协调器（Transaction Coordinator），两阶段提交

\- 使用场景：`consume-process-produce`（如消费 A Topic，处理后发到 B Topic，整体原子）

***\*对比：\****

| 维度 | 幂等 Producer | 事务 Producer |

|------|-------------|-------------|

| 作用范围 | 单 Partition | 跨 Partition、跨 Topic |

| 去重周期 | 单 Producer 会话 | 跨 Producer 会话（基于 transactional.id） |

| 性能开销 | 低 | 中等（需要事务协调） |

| 配置复杂度 | 低 | 高 |

***\*项目实践：\**** 本项目未使用 Kafka 原生的幂等/事务 Producer，而是通过 ***\*应用层 EventEnvelope + Redis SETNX\**** 实现业务级幂等。

**---**

**### Q20: Kafka 的高水位（HW）和 LEO 是什么？它们如何影响消费？**

***\*回答要点：\****

\- ***\*LEO（Log End Offset）\****：每个 Partition 的下一条待写入消息的 Offset

\- ***\*HW（High Watermark）\****：Consumer 可见的最大 Offset，等于所有 ISR 副本 LEO 的最小值

***\*消费可见性：\****

\```

Consumer 只能消费到 HW 之前的消息。

HW 之后的消息（即使已写入 Leader）对 Consumer 不可见。

\```

***\*HW 的作用：\****

\1. 防止 Consumer 消费到未完全同步的数据（Leader 宕机后可能丢失）

\2. 保证副本间的数据一致性

***\*示例：\****

\```

Leader:   [0] [1] [2] [3] [4]   LEO=5

Follower: [0] [1] [2] [3]       LEO=4

​                        ↑

​                       HW=4（所有 ISR 副本 LEO 的最小值）

Consumer 最多消费到 offset=4

\```

**---**

**## 五、幂等性与去重篇**

**### Q21: 在 Kafka 消费端，为什么需要做幂等处理？有哪些实现方式？**

***\*回答要点：\****

***\*为什么需要幂等：\****

\- Kafka 默认提供 ***\*At-least-once\**** 语义（消息可能被重复消费）

\- 重复消费场景：Rebalance 后重复消费、Consumer 崩溃重启、网络重试

\- 如果业务操作非幂等（如「增加计数」+1），重复消费会导致数据错误

***\*实现方式：\****

| 方式 | 原理 | 优点 | 缺点 |

|------|------|------|------|

| ***\*数据库唯一键\**** | INSERT ... ON DUPLICATE KEY UPDATE | 简单可靠 | 仅适用插入场景 |

| ***\*Redis SETNX\**** | 用消息 ID 做 Key，SETNX 防重 | 高性能 | 需要维护 Redis |

| ***\*数据库去重表\**** | 维护已处理消息 ID 表 | 强一致 | 性能较低 |

| ***\*消息状态机\**** | 基于业务状态判断是否已处理 | 无额外存储 | 场景受限 |

| ***\*Kafka 幂等 Producer\**** | Broker 端去重 | 原生支持 | 仅单 Partition |

***\*项目实践：\**** 本项目统一使用 ***\*Redis SETNX\**** 方案：

\```go

// common/kafka/idempotency.go

func (s *RedisIdempotencyStore) MarkProcessed(ctx context.Context, eventID string) (bool, error) {

​    // SETNX: 返回 true 表示首次处理，false 表示已处理过

​    result, err := s.client.SetNX(ctx, eventID, "1", DefaultEventTTL).Result()

​    return result, err  // true = 可以继续处理

}

// DefaultEventTTL = 24h，防止 Redis 内存无限增长

\```

***\*幂等 Key 的生成：\****

\```go

func BuildIdempotencyKey(eventType, aggregateID string, occurredAt int64) string {

​    return fmt.Sprintf("idempotent:%s:%s:%d", eventType, aggregateID, occurredAt)

}

\```

**---**

**### Q22: Redis SETNX 做幂等有什么问题？如何优化？**

***\*回答要点：\****

***\*潜在问题：\****

\1. ***\*Redis 不可用\**** → 降级策略是什么？

\2. ***\*TTL 过期\**** → 24 小时后的重复消息会穿透

\3. ***\*SETNX 成功但业务失败\**** → 消息被标记为已处理，实际未处理

\4. ***\*Redis 主从切换丢数据\**** → 已标记的消息可能丢失

***\*优化方案：\****

\1. ***\*降级策略\****：Redis 不可用时，使用 Noop 存储或限流

   \```go

   // 本项目提供了 NoopIdempotencyStore，不做去重

   \```

\2. ***\*先业务后标记\****：先执行数据库操作，成功后再 SETNX

   \- 风险：业务成功但 SETNX 失败 → 下次重复执行

   \- 缓解：业务逻辑本身幂等（如 `UPDATE SET count = count + 1 WHERE id = ?`）

\3. ***\*持久化标记\****：使用数据库去重表作为最终兜底

\4. ***\*Redis AOF + 持久化\****：配置 `appendfsync everysec` 减少丢数据风险

***\*项目实践：\****

\```go

// common/kafka/consumer.go 中的处理流程

func (c *Consumer) consumeMessage(msg *sarama.ConsumerMessage) error {

​    // 1. 解析 EventEnvelope

​    envelope, err := DecodeEnvelope(msg.Value)

​    

​    // 2. 幂等检查

​    if c.idempotency != nil {

​        isNew, err := c.idempotency.MarkProcessed(ctx, envelope.EventID)

​        if !isNew {

​            return nil  // 已处理，跳过

​        }

​    }

​    

​    // 3. 执行业务逻辑

​    return c.handler.Handle(ctx, envelope)

}

\```

**---**

**### Q23: 本项目中的 EventEnvelope 设计解决了什么问题？**

***\*回答要点：\****

\```go

type EventEnvelope struct {

​    EventID       string          // 全局唯一事件 ID

​    EventType     string          // 事件类型

​    AggregateType string          // 聚合类型

​    AggregateID   string          // 聚合 ID

​    Producer      string          // 生产者标识

​    TraceID       string          // 链路追踪 ID

​    SchemaVersion int             // Schema 版本

​    OccurredAt    int64           // 事件时间戳

​    RetryCount    int             // 重试计数

​    Data          json.RawMessage // 实际业务数据

}

\```

***\*解决的问题：\****

| 字段 | 解决的问题 |

|------|-----------|

| ***\*EventID\**** | 幂等去重 Key 来源 |

| ***\*EventType + AggregateType\**** | 消费者路由分发 |

| ***\*AggregateID\**** | Partition Key 来源（保证同一聚合的顺序） |

| ***\*Producer\**** | 可追溯消息来源 |

| ***\*TraceID\**** | 分布式链路追踪 |

| ***\*OccurredAt\**** | 事件时序判断、延迟监控 |

| ***\*RetryCount\**** | 重试控制、DLQ 触发 |

| ***\*Data (json.RawMessage)\**** | 延迟解析，Consumer 按需反序列化 |

***\*核心价值：\****

\1. ***\*自描述消息\****：Consumer 不需要额外上下文就能理解消息

\2. ***\*可追溯\****：EventID + Producer + TraceID 形成完整的调用链路

\3. ***\*版本兼容\****：SchemaVersion 支持消息格式演进

**---**

**## 六、消息顺序性篇**

**### Q24: Kafka 如何保证消息的顺序性？**

***\*回答要点：\****

***\*Kafka 的顺序保证：\****

\- ***\*单 Partition 内有序\****：同一 Partition 内消息按写入顺序排列

\- ***\*跨 Partition 无序\****：不同 Partition 之间没有顺序保证

***\*保证顺序的方法：\****

\1. ***\*使用相同的 Key\****：相同 Key 的消息进入同一 Partition

\2. ***\*单 Partition Topic\****：所有消息天然有序（但牺牲并行度）

\3. ***\*Producer 端** **`max.in.flight.requests.per.connection = 1`*****\***：防止重试打乱顺序

***\*项目实践：\****

\```go

// 基于 AggregateID 设置 Key，保证同一实体的消息有序

// 例如：同一帖子的 star → unstar → star 事件按顺序消费

producer.SendEvent(

​    "post-star",

​    fmt.Sprintf("%d", postID),  // Key = postID，进入同一 Partition

​    "post.starred", "interaction-rpc", "post", fmt.Sprintf("%d", postID),

​    payload,

)

\```

***\*需要注意：\**** 如果 Producer 配置了 `retries > 0` 且 `max.in.flight > 1`，重试可能导致消息乱序。需要设为 `max.in.flight = 1` 来保证严格顺序。

**---**

**### Q25: 什么场景下需要严格的消息顺序？什么场景下不需要？**

***\*回答要点：\****

***\*需要严格顺序的场景：\****

\- 状态机转换（如订单状态：待支付 → 已支付 → 已发货）

\- 增量更新（如 `count++` 需要知道当前值）

\- 事件溯源（Event Sourcing）中的聚合重建

***\*不需要严格顺序的场景：\****

\- 独立实体的操作（不同帖子的 star 操作互不影响）

\- 幂等更新（如覆盖式更新 `UPDATE SET status = 'active'`）

\- 最终一致性场景（如缓存刷新、搜索索引更新）

***\*项目实践分析：\****

| 场景 | 是否需顺序 | 原因 |

|------|-----------|------|

| post-star / post-unstar | ***\*需要\**** | star→unstar→star 需按序处理，否则计数错误 |

| post-created → user post_count++ | ***\*需要\**** | 同一用户 post 计数依赖顺序 |

| comment-created → post comment_count++ | ***\*需要\**** | 计数累加依赖顺序 |

| Canal → ES 同步 | ***\*不需要\**** | 覆盖式更新，幂等 |

***\*本项目的处理方式：\****

\- interaction-rpc 消费者使用 ***\*Map 批量去重\****，在批处理窗口内按 Key 去重，保证同一 (postID, userID) 的最新状态生效

\- 通过 `WHERE count > 0` 和 `GREATEST(count-1, 0)` 防护计数异常

**---**

**## 七、批量处理篇**

**### Q26: Kafka 消费端的批量处理有哪些优势和挑战？**

***\*回答要点：\****

***\*优势：\****

\1. ***\*减少 DB 操作次数\****：批量 INSERT 比逐条 INSERT 性能高数十倍

\2. ***\*减少网络开销\****：一次请求处理多条消息

\3. ***\*提高吞吐量\****：减少 Consumer 的 `poll()` 频率

\4. ***\*事务性\****：整个批次可以包装在一个 DB 事务中

***\*挑战：\****

\1. ***\*消息积压风险\****：等待凑批增加端到端延迟

\2. ***\*部分失败处理\****：批次中某条消息失败，整个批次如何回滚？

\3. ***\*内存占用\****：缓冲消息消耗内存

\4. ***\*Offset 提交时机\****：批次处理成功才能提交 Offset

***\*项目实践：\**** 本项目有两种批量实现：

***\*方案一：通用 BatchConsumer\****

\```go

// common/kafka/batch_consumer.go

type BatchConsumer struct {

​    maxSize      int           // 默认 100

​    flushTimeout time.Duration // 默认 500ms

​    buffer       []BatchMessage

}

// 凑批逻辑：数量达标或超时触发

func (bc *BatchConsumer) addToBatch(msg BatchMessage) {

​    bc.buffer = append(bc.buffer, msg)

​    if len(bc.buffer) >= bc.maxSize {

​        bc.flush()

​    }

}

\```

***\*方案二：interaction-rpc 自定义批处理\****

\```go

// 内存 Map 去重 + 批量写入

starMap := make(map[string]*model.PostStar)       // Key = "postID:userID"

unstarUserPostIds := make(map[int64][]int64)      // 待删除集合

// 触发条件：100条 或 500ms

if len(starMap) >= 100 || time.Since(lastFlush) > *500ms* {

​    // 批量 Create + 批量 Delete

​    starDAO.BatchCreate(ctx, stars)

​    starDAO.Delete(ctx, userID, postIDs)

}

\```

**---**

**### Q27: 批量处理失败时如何保证数据不丢？**

***\*回答要点：\****

***\*本项目方案：DLQ + 全部回退\****

\```go

// common/kafka/batch_consumer.go

func (bc *BatchConsumer) flush() error {

​    if err := bc.handler.HandleBatch(ctx, bc.buffer); err != nil {

​        // 整个批次失败 → 全部发送到 DLQ

​        for _, msg := range bc.buffer {

​            dlqMsg := &sarama.ProducerMessage{

​                Topic: bc.dlqTopic,

​                Key:   msg.Key,

​                Value: msg.Value,

​                Headers: []sarama.RecordHeader{

​                    {Key: []byte("x-source-topic"), Value: []byte(msg.Topic)},

​                    {Key: []byte("x-original-offset"), Value: []byte(fmt.Sprintf("%d", msg.Offset))},

​                    {Key: []byte("x-error"), Value: []byte(err.Error())},

​                },

​            }

​            bc.dlqProducer.SendMessage(dlqMsg)

​        }

​        // 不提交 Offset，让 Consumer 重新消费

​        return err

​    }

​    return nil

}

\```

***\*其他常见策略：\****

\- ***\*逐条回退\****：只把失败的记录重试，成功的正常提交

\- ***\*部分成功提交\****：记录成功和失败的 Offset，只提交成功部分

\- ***\*业务级补偿\****：失败时执行补偿事务

**---**

**## 八、死信队列篇**

**### Q28: 什么是死信队列（DLQ）？什么时候需要 DLQ？**

***\*回答要点：\****

***\*死信队列（Dead Letter Queue）：\**** 用于存放无法被正常消费的消息的特殊 Topic。

***\*需要 DLQ 的场景：\****

\1. ***\*消息格式错误\****：反序列化失败，永远无法成功消费

\2. ***\*重试次数耗尽\****：业务处理反复失败

\3. ***\*业务规则拒绝\****：消息不符合业务规则（如引用已删除的实体）

\4. ***\*批量处理失败\****：整个批次回退

***\*DLQ 的设计原则：\****

\1. 保留原始消息内容和元数据

\2. 记录失败原因

\3. 支持人工介入和重放

\4. 设置监控告警

***\*项目实践：\**** 本项目有两个 DLQ：

| DLQ Topic | 来源 | 触发条件 |

|-----------|------|---------|

| `interaction-dlq` | interaction-rpc 批量处理失败 | 整个批次写入 DB 失败 |

| `comment-dlq` | comment-rpc 消费重试耗尽 | 重试 3 次后仍失败 |

\```go

// comment-rpc 的 DLQ 逻辑

func (c *CommentConsumer) handleWithRetry(msg *sarama.ConsumerMessage) error {

​    retryCount := getRetryCount(msg.Headers)

​    

​    if err := c.handleMessage(msg); err != nil {

​        if retryCount >= c.maxRetryCount {  // 达到最大重试次数

​            return c.sendToDLQ(msg, err)     // 发送到 DLQ

​        }

​        return c.republishWithRetry(msg, retryCount+1)  // 重新发布，增加重试计数

​    }

​    return nil

}

\```

**---**

**### Q29: DLQ 中的消息如何处理？**

***\*回答要点：\****

***\*处理流程：\****

\1. ***\*监控告警\****：DLQ 有消息进入时触发告警

\2. ***\*问题排查\****：根据 `x-error` Header 定位根因

\3. ***\*修复后重放\****：修复代码或数据后，将 DLQ 消息重新投递到原始 Topic

\4. ***\*人工确认丢弃\****：对于确实无用的消息，确认后丢弃

***\*重放工具设计：\****

\```go

// 概念性代码：DLQ 重放工具

func ReplayDLQ(dlqTopic, sourceTopic string) {

​    consumer := NewConsumer(dlqTopic)

​    producer := NewProducer(sourceTopic)

​    

​    for msg := range consumer.Messages() {

​        sourceTopic := getHeader(msg, "x-source-topic")

​        // 重新投递到原始 Topic

​        producer.Send(sourceTopic, msg.Key, msg.Value)

​    }

}

\```

**---**

**## 九、重试机制篇**

**### Q30: Kafka 消费端有哪些重试策略？各自优劣？**

***\*回答要点：\****

***\*策略一：不提交 Offset，利用 Rebalance 重试\****

\- 原理：Handler 返回 error，不标记消息，下次 Rebalance 后重新消费

\- 优点：实现简单

\- 缺点：重试间隔不确定，可能阻塞整个 Partition

***\*策略二：内部重试队列 + 退避\****

\- 原理：失败消息放入内部重试队列，定时重试

\- 优点：不阻塞正常消费

\- 缺点：增加内存开销，实现复杂

***\*策略三：重新投递到原 Topic（本项目方案）\****

\- 原理：消费失败后，将消息重新发到同一 Topic（带重试计数 Header）

\- 优点：利用 Kafka 的持久化能力，不丢消息

\- 缺点：可能改变消息顺序

***\*策略四：重试 Topic（多级延迟队列）\****

\- 原理：多个延迟 Topic（如 retry-1s, retry-10s, retry-1m），消息逐级投递

\- 优点：精确控制重试间隔

\- 缺点：Topic 数量多，维护成本高

***\*项目实践（comment-rpc 的策略三实现）：\****

\```go

const maxRetryCount = 3

func (c *CommentConsumer) republishWithRetry(msg *sarama.ConsumerMessage, retryCount int) error {

​    newMsg := &sarama.ProducerMessage{

​        Topic: msg.Topic,  // 投递到原 Topic

​        Key:   sarama.ByteEncoder(msg.Key),

​        Value: sarama.ByteEncoder(msg.Value),

​        Headers: append(msg.Headers, sarama.RecordHeader{

​            Key:   []byte("x-retry-count"),

​            Value: []byte(strconv.Itoa(retryCount)),

​        }),

​    }

​    _, _, err := c.producer.SendMessage(newMsg)

​    return err

}

\```

**---**

**### Q31: 重试时如何避免消息乱序？**

***\*回答要点：\****

***\*问题：\**** 如果使用「重新投递到原 Topic」策略，消息 A（处理中）→ 消息 B（处理成功）→ 消息 A 重试投递，A 会排在 B 后面，破坏了顺序。

***\*解决方案：\****

\1. ***\*****`max.in.flight.requests.per.connection = 1`*****\***：严格串行发送

\2. ***\*单 Partition + 阻塞重试\****：当前消息失败，不拉取下一条，阻塞重试直到成功或达到上限

\3. ***\*业务层版本号 / 序列号\****：消费时检查消息的版本号，丢弃过期消息

   \```go

   func handleMessage(msg) error {

​       currentVersion := getCurrentVersion(msg.AggregateID)

​       if msg.Version < currentVersion {

​           return nil  // 旧消息，跳过

​       }

​       // 处理...

   }

   \```

\4. ***\*使用幂等操作\****：即使乱序，结果也正确（如覆盖式更新而非增量更新）

***\*项目实践：\**** 本项目 interaction-rpc 使用 ***\*Map 去重\**** 来缓解乱序问题——在批处理窗口内，同一 Key 只保留最新状态：

\```go

// 后到的操作覆盖先到的

starMap[key] = &model.PostStar{PostID: postID, UserID: userID}

// 如果 star 和 unstar 乱序到达，窗口内最后一条决定最终状态

\```

**---**

**## 十、监控与运维篇**

**### Q32: Kafka 需要监控哪些关键指标？**

***\*回答要点：\****

***\*Producer 端：\****

\- `record-send-rate`：发送速率

\- `record-error-rate`：错误率

\- `record-retry-rate`：重试率

\- `request-latency-avg`：请求延迟

\- `buffer-available-bytes`：缓冲区可用大小（满了会阻塞）

***\*Broker 端：\****

\- `UnderReplicatedPartitions`：未完全复制的 Partition 数

\- `ActiveControllerCount`：Controller 存活数（应该为 1）

\- `OfflinePartitionsCount`：离线 Partition 数

\- `BytesInPerSec / BytesOutPerSec`：进出流量

\- `RequestQueueSize`：请求队列大小

***\*Consumer 端：\****

\- `records-lag-max`：最大消费 Lag

\- `records-consumed-rate`：消费速率

\- `fetch-rate`：拉取频率

\- `rebalance-rate`：Rebalance 频率

***\*项目实践：\**** 本项目在 Consumer 中实现了 Lag 监控：

\```go

func logConsumerLag(partition int32, highWaterMark, offset int64) {

​    lag := highWaterMark - offset

​    if lag > 30 {

​        logx.Slowf("[KAFKA] Consumer lag is high, partition: %d, lag: %d", partition, lag)

​    }

}

\```

**---**

**### Q33: Kafka 集群如何做容量规划？**

***\*回答要点：\****

***\*关键参数估算：\****

\1. ***\*磁盘容量：\****

   \```

   所需磁盘 = 日均消息量 × 平均消息大小 × 保留天数 × 副本数 × 1.3（余量）

   \```

   例如：日均 1 亿条 × 1KB × 7 天 × 3 副本 = 约 2.1TB

\2. ***\*网络带宽：\****

   \```

   带宽 = 峰值吞吐量 × 平均消息大小 × 副本因子 × 2（进出）

   \```

\3. ***\*Broker 数量：\****

   \- 根据 Partition 数量：每个 Broker 建议不超过 4000 Partition

   \- 根据磁盘容量：每个 Broker 不超过 4TB（便于迁移恢复）

   \- 根据吞吐量：单个 Broker 可达 100MB/s+

\4. ***\*内存：\****

   \- Page Cache 尽可能大（Kafka 重度依赖 OS 页缓存）

   \- JVM Heap 不需要太大（6GB 通常足够）

***\*项目实践：\**** 本项目 K8s 部署中，Kafka 配置的资源限制：

\```yaml

\# infra-kafka.yaml

resources:

  requests:

​    memory: "512Mi"

​    cpu: "250m"

  limits:

​    memory: "1Gi"

​    cpu: "500m"

\```

**---**

**### Q34: Kafka 如何进行数据迁移和分区重分配？**

***\*回答要点：\****

***\*常见场景：\****

\1. 新增 Broker 节点，需要重新分配 Partition

\2. Broker 退役，需要迁移数据

\3. 负载不均衡，需要均衡 Partition 分布

***\*操作工具：\**** `kafka-reassign-partitions.sh`

***\*步骤：\****

\1. 生成迁移计划 JSON

\2. 执行迁移（可限速，避免影响在线服务）

\3. 验证迁移结果

***\*注意事项：\****

\- 迁移过程会增加网络和磁盘 I/O

\- 建议设置限速（`--throttle`）

\- 使用 `--verify` 确认迁移完成

\- 迁移期间不进行其他运维操作

**---**

**## 十一、CDC 与 Canal 篇**

**### Q35: 什么是 CDC？Canal 是如何工作的？**

***\*回答要点：\****

***\*CDC（Change Data Capture）：\**** 捕获数据库的变更事件，同步到其他系统。

***\*Canal 工作原理：\****

\1. Canal 伪装成 MySQL Slave，向 Master 发送 dump 协议

\2. Master 推送 binlog 事件给 Canal

\3. Canal 解析 binlog，提取 INSERT/UPDATE/DELETE 事件

\4. Canal 将事件投递到 Kafka

***\*项目实践架构：\****

\```

MySQL ──binlog──▶ Canal Server ──canal_topic──▶ canal Service ──▶ Elasticsearch

\```

\```go

// canal Service 消费 Kafka 消息并同步到 ES

func (c *KafkaConsumer) consumeMessage(msg *sarama.ConsumerMessage) error {

​    var canalMsg CanalMessage

​    json.Unmarshal(msg.Value, &canalMsg)

​    

​    // 批量积累

​    c.buffer = append(c.buffer, canalMsg)

​    if len(c.buffer) >= 100 || time.Since(c.lastFlush) > time.Second {

​        c.syncHandler.HandleMessages(ctx, c.buffer)  // 批量同步 ES

​        c.buffer = c.buffer[:0]

​    }

​    return nil

}

\```

**---**

**### Q36: Canal + Kafka 做数据同步有哪些注意事项？**

***\*回答要点：\****

\1. ***\*消息顺序性\****：同一行数据的变更必须顺序处理（使用主键作为 Kafka Key）

\2. ***\*数据一致性\****：全量 + 增量的数据对账

   \- 定期对比 MySQL 和 ES 的数据

   \- 处理 binlog 丢失或解析失败的情况

\3. ***\*DDL 变更处理\****：表结构变更（加字段等）需要 Canal 和消费者同步更新

\4. ***\*大事务处理\****：大批量 UPDATE 会产生大量 binlog，可能导致 Kafka 消息积压

\5. ***\*幂等性\****：ES 同步应使用覆盖式更新（`_doc` 或 `_update`），天然幂等

\6. ***\*延迟监控\****：binlog 产生到 ES 索引完成的端到端延迟

***\*项目实践：\**** 本项目 Canal 消费者使用批量处理（100 条或 1 秒），解码失败的消息标记后跳过（不阻塞消费）：

\```go

if err := json.Unmarshal(msg.Value, &canalMsg); err != nil {

​    session.MarkMessage(msg, "")  // 跳过无法解析的消息

​    return nil

}

\```

**---**

**## 十二、架构设计篇**

**### Q37: 本项目中 Kafka 的消息流转架构是怎样的？画出整体架构图。**

***\*回答要点：\****

\```

​                          ┌──────────────────────────────────────────────┐

​                          │                 Kafka Cluster                │

​                          │                                              │

┌──────────┐   post-created   ┌──────────────────┐   post-star    ┌───────────────┐

│content-rpc│────────────────▶│                  │◀───────────────│interaction-rpc│

│(Producer) │   post-deleted  │                  │  post-unstar   │ (Producer+    │

└──────────┘                  │                  │  post-collect   │  Consumer)    │

​                              │    15 Topics     │  post-uncollect │               │

┌──────────┐  user-followed   │                  │                 └───────────────┘

│relation- │────────────────▶│                  │

│rpc       │  user-unfollowed │                  │  comment-create  ┌───────────┐

└──────────┘                  │                  │◀─────────────────│comment-rpc│

​                              │                  │  comment-created  │(Producer+ │

┌──────────┐  post-created    │                  │  comment-deleted  │ Consumer) │

│ user-rpc │◀─────────────────│                  │  comment-vote     └───────────┘

│(Consumer)│  post-deleted    │                  │

│          │  user-followed   │                  │

│          │  user-unfollowed │                  │

└──────────┘                  └──────────────────┘

​                                     ▲

​                                     │ canal_topic

​                              ┌──────┴──────┐

​                              │ Canal Server │

​                              └──────┬──────┘

​                                     │ binlog

​                              ┌──────┴──────┐

​                              │    MySQL     │

​                              └─────────────┘

\```

***\*6 个微服务的角色：\****

| 服务 | 角色 | 生产的 Topic | 消费的 Topic |

|------|------|-------------|-------------|

| content-rpc | Producer + Consumer | post-created, post-deleted | 6 个 stats 相关 Topic |

| user-rpc | Consumer | - | post-created/deleted, user-followed/unfollowed |

| relation-rpc | Producer | user-followed, user-unfollowed | - |

| interaction-rpc | Producer + Consumer | post-star/unstar/collect/uncollect | 同左（自产自消） |

| comment-rpc | Producer + Consumer | comment-create/created/deleted/vote | 同左（自产自消） |

| canal | Consumer | - | canal_topic（MySQL binlog → ES） |

**---**

**### Q38: 为什么 interaction-rpc 和 comment-rpc 采用「自产自消」的模式？**

***\*回答要点：\****

***\*「自产自消」：\**** 同一个服务既是消息生产者也是消费者。

***\*设计目的：\****

\1. ***\*异步解耦\****：API 请求快速返回，持久化操作异步完成

   \```

   用户请求 → API 返回 "success" → Kafka → 消费者写入 DB

   \```

\2. ***\*削峰填谷\****：高并发场景下，Kafka 作为缓冲层，消费者按自身能力消费

\3. ***\*批量优化\****：消费者端可以攒批写入 DB，提升吞吐量

   \```go

   // 单条 INSERT 变成批量 INSERT

   starDAO.BatchCreate(ctx, stars)  // 一次写入 100 条

   \```

\4. ***\*容错与重试\****：DB 暂时不可用时，消息在 Kafka 中持久化，恢复后继续消费

***\*权衡：\****

\- 优点：高吞吐、低延迟、解耦

\- 缺点：数据一致性延迟（最终一致性）、增加了运维复杂度

**---**

**### Q39: 为什么 content-rpc 和 user-rpc 的消费者使用了不同的 Consumer Group？**

***\*回答要点：\****

***\*背景：\**** content-rpc 消费 6 个 Topic 共用 `post-comment-count-consumer-group`；user-rpc 消费 4 个 Topic 各自使用独立 Group（如 `user-post-count-consumer-group-created`）。

***\*原因分析：\****

\1. ***\*content-rpc 共用 Group\****：这 6 个 Topic 的处理逻辑相似（都是更新 post 表的统计字段），共享 Group 可以复用 Consumer 实例，减少资源开销。

\2. ***\*user-rpc 独立 Group\****：每个 Topic 对应不同的用户统计字段（post_count、following_count、follower_count），处理逻辑差异较大，独立 Group 便于：

   \- 独立扩缩容：某个 Topic 流量大时，单独增加 Consumer

   \- 故障隔离：一个 Consumer 出问题不影响其他

   \- 独立监控：分别监控每个 Topic 的消费 Lag

***\*设计原则：\****

\- 处理逻辑相似 → 共用 Group

\- 处理逻辑独立 → 独立 Group

\- 需要独立扩缩容 → 独立 Group

**---**

**### Q40: 本项目的事件驱动架构（EDA）有什么优缺点？**

***\*回答要点：\****

***\*优点：\****

\1. ***\*松耦合\****：服务之间不直接调用，通过事件异步通信

   \- 例如：content-rpc 发帖后不直接调用 user-rpc，而是发送 `post-created` 事件

\2. ***\*可扩展性\****：新增消费者不影响现有服务

   \- 例如：未来新增「推荐服务」只需订阅 `post-created` 事件

\3. ***\*容错性\****：服务暂时不可用时，事件在 Kafka 中持久化

   \- 例如：user-rpc 重启期间，`post-created` 事件积压，恢复后继续消费

\4. ***\*追溯性\****：EventID + TraceID 实现完整的调用链追踪

***\*缺点：\****

\1. ***\*最终一致性\****：数据不是实时一致

   \- 例如：帖子创建后，用户 post_count 可能有短暂延迟

\2. ***\*复杂性增加\****：

   \- 调试困难：需要关联多个服务的日志

   \- 需要处理消息丢失、重复、乱序

\3. ***\*运维成本\****：

   \- 需要维护 Kafka 集群

   \- 需要监控 DLQ、Consumer Lag

\4. ***\*数据一致性保障\****：

   \- 需要幂等性设计

   \- 需要补偿机制

**---**

**## 十三、综合场景题**

**### Q41: 如果用户快速 star → unstar → star 同一个帖子，系统如何保证最终计数正确？**

***\*回答要点：\****

***\*场景分析：\**** 用户快速操作，Kafka 中可能乱序到达：star → star → unstar。

***\*本项目的多层保障：\****

***\*第一层：Partition Key 保证顺序\****

\```go

// 使用 postID 作为 Key，同一帖子的消息进入同一 Partition

producer.SendEvent("post-star", fmt.Sprintf("%d", postID), ...)

producer.SendEvent("post-unstar", fmt.Sprintf("%d", postID), ...)

\```

***\*第二层：Map 批处理去重（interaction-rpc）\****

\```go

// 在批处理窗口内（100条 或 500ms），同一 Key 只保留最新状态

starMap[key] = &model.PostStar{PostID: postID, UserID: userID}

// 如果 unstar 后到，会从 starMap 删除，加入 unstarUserPostIds

\```

***\*第三层：数据库层面的幂等防护\****

\```go

// 取消 star 时，WHERE 条件保证不会删除不存在的记录

starDAO.Delete(ctx, userID, postID)  // DELETE 天然幂等

// 计数更新时使用 GREATEST 防止负数

UPDATE post SET upvote_count = GREATEST(upvote_count - 1, 0) WHERE id = ?

\```

***\*最终结果：\**** 无论消息如何乱序到达，通过 Partition 有序 + Map 去重 + 幂等 DB 操作，保证最终计数与用户意图一致。

**---**

**### Q42: 设计一个支持百万 QPS 的点赞系统，Kafka 如何发挥作用？**

***\*回答要点：\****

***\*架构设计：\****

\```

用户请求 → API Gateway → 点赞服务 → Kafka → 计数服务 → Redis → 定时刷 DB

\```

***\*分层设计：\****

\1. ***\*接入层\****：限流 + 降级，防止突发流量打垮服务

\2. ***\*Kafka 缓冲层\****：

   \- 使用 AsyncProducer（本项目方案），低延迟发送

   \- Partition 按 postID 哈希，保证同一帖子计数有序

   \- Snappy 压缩减少网络传输

\3. ***\*消费层（计数聚合）\****：

   \- 本地内存聚合：先在内存中聚合计数，定时批量写入 Redis

   \- 例如：100 条 star 消息 → 内存 `count++` → 每 100ms 刷一次 Redis

\4. ***\*存储层\****：

   \- Redis：实时计数（`INCR post:123:upvote_count`）

   \- DB：定时从 Redis 同步（每分钟一次）

   \- ES：通过 Canal 或异步同步，支持搜索

***\*关键优化：\****

\- Kafka Partition 数量 = 预期 QPS / 单 Partition 吞吐（约 10w/s）

\- Consumer 使用批量处理 + 内存聚合

\- 计数操作先写 Redis（O(1)），异步同步到 DB

***\*项目实践经验：\****

\- AsyncProducer: acks=1, Snappy, 100 条批量

\- BatchConsumer: 100 条或 500ms 触发

\- DB 批量写入: `BatchCreate` 而非逐条 INSERT

**---**

**### Q43: 如果 Kafka 集群完全不可用，系统如何降级？**

***\*回答要点：\****

***\*降级策略（按优先级）：\****

***\*Level 1：降级为同步写入\****

\```go

func (l *StarPostLogic) StarPost(req *StarPostReq) (*StarPostResp, error) {

​    // 尝试 Kafka 发送

​    if err := l.producer.SendEvent(...); err != nil {

​        // 降级：直接写入 DB

​        logx.Errorf("Kafka unavailable, fallback to DB: %v", err)

​        return l.starDAO.Create(ctx, &PostStar{...})

​    }

​    return &StarPostResp{}, nil

}

\```

***\*Level 2：本地缓冲 + 重试\****

\- 消息写入本地文件或内存队列

\- 定时重试发送到 Kafka

\- 风险：进程重启丢失

***\*Level 3：功能降级\****

\- 非核心功能直接关闭（如 ES 索引同步）

\- 核心功能返回成功（异步操作允许丢失）

***\*Level 4：限流 + 熔断\****

\- 限制 API 请求速率

\- 熔断器打开时直接拒绝请求

***\*监控告警：\****

\- Kafka 不可用立即触发 P0 告警

\- 降级期间密切监控 DB 压力

**---**

**### Q44: 项目中为什么有些服务用 SyncProducer 而有些用 AsyncProducer？依据是什么？**

***\*回答要点：\****

| 服务 | Producer 类型 | 生产的 Topic | 选择原因 |

|------|-------------|-------------|---------|

| content-rpc | ***\*SyncProducer\**** | post-created, post-deleted | 帖子创建/删除是关键操作，需要保证消息可靠送达 |

| relation-rpc | ***\*SyncProducer\**** | user-followed, user-unfollowed | 关注关系是核心数据，需要可靠 |

| interaction-rpc | ***\*AsyncProducer\**** | post-star/unstar/collect/uncollect | 高频操作（点赞/收藏），追求低延迟和高吞吐 |

| comment-rpc | ***\*AsyncProducer\**** | comment-create/created/deleted/vote | 评论相关操作频率高，吞吐优先 |

***\*选择依据：\****

\1. ***\*业务重要性\****：核心数据变更 → Sync；辅助统计 → Async

\2. ***\*请求频率\****：低频关键 → Sync；高频 → Async

\3. ***\*延迟要求\****：对延迟敏感 → Async（不阻塞请求线程）

\4. ***\*数据丢失容忍度\****：不可丢 → Sync（acks=all）；可容忍少量丢失 → Async（acks=1）

***\*代码实现对比：\****

\```go

// SyncProducer: 适用于关键业务

config.Producer.RequiredAcks = sarama.WaitForAll   // 所有 ISR 确认

config.Producer.Retry.Max = 5                      // 最多重试 5 次

config.Producer.Return.Successes = true            // 需要同步等待结果

// AsyncProducer: 适用于高吞吐场景

config.Producer.RequiredAcks = sarama.WaitForLocal  // Leader 确认即可

config.Producer.Retry.Max = 3                       // 最多重试 3 次

config.Producer.Compression = sarama.CompressionSnappy  // 压缩

\```

**---**

**### Q45: 本项目使用 Redis SETNX 做幂等，如果 Redis 不可用怎么办？**

***\*回答要点：\****

***\*问题分析：\**** Redis SETNX 是消费端幂等性的核心依赖。如果 Redis 不可用：

\1. `MarkProcessed()` 返回 error

\2. 消费者无法判断消息是否已处理

\3. 可能造成重复消费或消息丢失

***\*本项目的应对：\****

***\*方案一：NoopIdempotencyStore（无去重模式）\****

\```go

// common/kafka/idempotency.go

type NoopIdempotencyStore struct{}

func (n *NoopIdempotencyStore) MarkProcessed(ctx context.Context, eventID string) (bool, error) {

​    return true, nil  // 总是返回「可以处理」

}

\```

***\*方案二：业务逻辑本身幂等（最佳实践）\****

\```go

// DELETE 操作天然幂等

db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&PostStar{})

// INSERT 使用 ON DUPLICATE KEY UPDATE

db.Clauses(clause.OnConflict{

​    Columns:   []clause.Column{{Name: "user_id"}, {Name: "post_id"}},

​    DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),

}).Create(&PostStar{})

// UPDATE 使用条件更新 + GREATEST 防护

db.Exec("UPDATE post SET upvote_count = GREATEST(upvote_count - 1, 0) WHERE id = ?", postID)

\```

***\*设计建议：\****

\1. ***\*幂等性作为最后防线\****：Redis SETNX 是性能优化（避免执行重复业务逻辑），而非唯一保障

\2. ***\*业务操作本身应尽可能幂等\****：DELETE、UPSERT、条件 UPDATE

\3. ***\*多层防护\****：Redis 去重 + 业务幂等 + DB 约束

**---**

**## 附录：本项目 Kafka 核心配置速查表**

**### Producer 配置**

| 配置项 | SyncProducer | AsyncProducer |

|--------|-------------|---------------|

| RequiredAcks | WaitForAll | WaitForLocal |

| Retry.Max | 5 | 3 |

| Compression | None | Snappy |

| Flush.Messages | - | 100 |

| Flush.Frequency | - | 10ms |

| MaxMessageBytes | - | 1MB (1048576) |

| Return.Successes | true | - |

**### Consumer 配置**

| 配置项 | 值 | 说明 |

|--------|-----|------|

| Version | sarama.V2_8_0_0 | Kafka 协议版本 |

| Consumer.Offsets.Initial | OffsetOldest / OffsetNewest | 首次消费起始位置 |

| Consumer.Group.Rebalance.Strategy | RoundRobin / Range | 分区分配策略 |

**### 服务维度速查**

| 服务 | Producer 类型 | 生产 Topic 数 | 消费 Topic 数 | Consumer Group 数 | 幂等 | DLQ |

|------|-------------|-------------|-------------|-------------------|------|-----|

| content-rpc | SyncProducer | 2 | 6 | 1 | Redis SETNX | ❌ |

| user-rpc | - | 0 | 4 | 4 | Redis SETNX | ❌ |

| relation-rpc | SyncProducer | 2 | 0 | - | - | ❌ |

| interaction-rpc | AsyncProducer | 4 | 4 | 1 | Redis SETNX | ✅ |

| comment-rpc | AsyncProducer | 4 | 3 | 1 | 手动 SETNX | ✅ |

| canal | - | 0 | 1 | 1 | ❌ | ❌ |

**---**

\> ***\*文档说明：\**** 本文档基于 SChill 项目的实际 Kafka 代码编写，所有代码示例均来自项目源码。问题设计涵盖基础概念、生产实践、架构设计和故障处理，适合准备 Kafka 相关面试的初中高级工程师参考。