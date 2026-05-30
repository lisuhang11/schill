
# 社交内容系统架构升级 - 技术面试题

## 面试官背景
P8+ 架构师，负责大厂社交产品后端架构设计与演进

---

## 目录
- [维度一：服务拆分与边界治理](#维度一服务拆分与边界治理)
- [维度二：批量接口设计及 N+1 消除](#维度二批量接口设计及-n1-消除)
- [维度三：搜索索引模型重构及零停机回填](#维度三搜索索引模型重构及零停机回填)
- [维度四：异步消息可靠性](#维度四异步消息可靠性)
- [维度五：缓存治理](#维度五缓存治理)
- [维度六：消费者生命周期与优雅上下线](#维度六消费者生命周期与优雅上下线)
- [维度七：安全基线](#维度七安全基线)
- [维度八：可观测性与回归测试策略](#维度八可观测性与回归测试策略)
- [维度九：性能优化与容量规划](#维度九性能优化与容量规划)
- [维度十：架构演进与技术债务](#维度十架构演进与技术债务)

---

## 维度一：服务拆分与边界治理

### 问题1.1
当前系统存在跨服务直查他域数据库的耦合问题，作为 Tech Lead，你会如何设计服务拆分方案？请从业务领域、技术实现、演进策略三个层面展开。

**考察点：DDD领域建模能力、服务拆分方法论、架构演进策略**

#### 参考答案：

**领域划分（DDD）：**

| 服务 | 核心职责 | 聚合根 | 提供的能力 |
|------|----------|--------|------------|
| 用户服务 | 用户信息、认证、权限 | User | 用户CRUD、登录、权限校验 |
| 内容服务 | 帖子、话题管理 | Post | 帖子CRUD、话题管理、内容检索 |
| 评论服务 | 评论、回复管理 | Comment | 评论CRUD、评论列表 |
| 互动服务 | 点赞、收藏、分享 | Like/Collect | 互动操作、状态查询 |
| 关系服务 | 关注、粉丝 | Follow | 关系管理、关系查询 |
| 搜索服务 | 全文检索 | - | 聚合搜索、索引同步 |

**互动服务独立的理由：**
- 互动是高频变更操作，与内容/评论的读多写少特征不同
- 互动逻辑独立演化，经常需要新增互动类型（比如"踩"、"转发"）
- 互动数据量大，可以独立存储和扩展

**演进策略：**
1. **识别阶段**：通过代码分析梳理所有跨库调用，建立依赖图谱
2. **隔离阶段**：先在被调用方提供 RPC 接口，调用方从直查改 RPC（不改代码结构）
3. **解耦阶段**：数据库权限收回，彻底禁止跨库访问
4. **优化阶段**：对聚合查询场景引入数据冗余或视图服务

---

### 问题1.2
你提到了按业务领域拆分，那么用户、内容、评论、互动这几个核心域，你会如何划分服务边界？特别是"互动"这个概念，你认为应该独立成服务还是分散在各个服务中？请说明理由。

**考察点：领域划分的深度思考、聚合根识别能力**

#### 参考答案：

**服务边界划分原则：**
1. **单一职责**：每个服务只负责一个核心业务领域
2. **高内聚低耦合**：服务内部高度聚合，服务之间松耦合
3. **独立部署**：每个服务可以独立部署和扩展
4. **数据隔离**：每个服务拥有自己的数据存储

**互动服务应该独立，理由：**
1. **变更频率不同**：互动逻辑（点赞、收藏）比内容、评论更频繁变化
2. **访问特征不同**：互动是写多读少，内容是读多写少
3. **技术栈可以差异化**：互动可以用 Redis 做计数，内容用 MySQL+ES
4. **业务演进独立**：未来可能新增互动类型，独立服务更容易扩展

**边界划分示例：**
- 内容服务：帖子、话题、内容审核
- 评论服务：评论、回复、评论审核
- 互动服务：点赞、收藏、分享、打赏
- 用户服务：用户信息、认证、权限、关注关系

---

### 问题1.3
在服务拆分过程中，必然会遇到"跨域数据关联"问题（比如帖子列表需要展示作者头像、昵称），你会如何解决？请对比不同方案的优劣势。

**考察点：数据冗余策略、数据一致性权衡、API设计能力**

#### 参考答案：

| 方案 | 适用场景 | 优点 | 缺点 |
|------|----------|------|------|
| 服务端聚合（API Composition） | 实时性要求高、数据量小 | 逻辑简单、数据一致 | 有RPC开销、故障扩散 |
| 数据冗余（CQRS） | 读多写少、可接受最终一致 | 查询性能高、无RPC依赖 | 存储成本、一致性复杂 |
| 物化视图（Materialized View） | 复杂聚合查询 | 查询高效 | 刷新延迟、存储成本 |
| 宽表设计（ES/HBase） | 搜索/分析场景 | 灵活性高 | 数据同步成本 |

**推荐策略：**
- 核心实时数据：服务端聚合 + 批量接口 + 多级缓存
- 非核心数据：通过消息异步冗余到读服务
- 搜索/分析场景：ES 宽表

**具体实现（帖子列表）：**
```go
// 方案1：服务端聚合（实时性高）
func GetPostList(ctx context.Context, req *GetPostListReq) (*GetPostListResp, error) {
    // 1. 从内容服务获取帖子列表
    posts, err := contentRpc.GetPostList(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 2. 批量获取用户信息
    userIds := extractUserIds(posts)
    users, err := userRpc.BatchGetUser(ctx, userIds)
    
    // 3. 批量获取话题信息
    topicIds := extractTopicIds(posts)
    topics, err := topicRpc.BatchGetTopic(ctx, topicIds)
    
    // 4. 聚合返回
    return aggregate(posts, users, topics), nil
}

// 方案2：数据冗余（查询性能高）
// 在内容服务冗余用户快照（昵称、头像），通过消息同步更新
```

---

### 问题1.4
你如何确保在重构过程中服务边界不被突破？有没有什么技术或流程上的保障措施？

**考察点：架构防护意识、工程化落地能力**

#### 参考答案：

**技术保障：**
1. **数据库权限控制**
   - 每个服务只拥有自己数据库的访问权限
   - 禁止创建跨库查询账号
   - 使用数据库审计日志监控跨库访问

2. **静态代码扫描**
   - 禁止直接 import 其他服务的 model 包
   - 禁止在代码中硬编码其他服务的数据库连接
   - 使用 ArchUnit 等工具做架构守护测试

3. **Service Mesh 治理**
   - 服务间调用必须通过 Service Mesh
   - 服务调用授权和审计
   - 流量控制和熔断降级

4. **代码生成规范**
   - goctl 生成的代码自动遵循分层架构
   - API 层 → Logic 层 → Model 层，禁止跨层访问

**流程保障：**
1. **API 变更评审**：任何 API 变更需要架构组评审
2. **定期架构评审**：每季度做一次架构健康度检查
3. **架构债务追踪**：使用 Jira 等工具追踪架构债务
4. **代码 Owner 机制**：每个服务有明确的 Owner，负责边界守护

**架构守护测试示例：**
```go
// ArchUnit 风格的测试
func TestArchitectureRules(t *testing.T) {
    // 规则1：content 服务不能直接依赖 comment 服务的 model
    classes().That().ResideInPackage("content.service").
        Should().NotDependOnClassesThat().ResideInPackage("comment.model")
    
    // 规则2：API 层只能调用 Logic 层，不能直接调用 Model 层
    classes().That().ResideInPackage("*.api").
        Should().OnlyDependOnClassesThat().ResideInPackage("*.logic")
}
```

---

### 问题1.5
假设现在要彻底禁止跨库查询，但历史代码中有大量这类调用。请设计一个可落地的迁移方案，包括但不限于：如何识别风险点、如何灰度验证、回滚策略是什么。

**考察点：风险评估能力、灰度发布策略、回滚预案设计**

#### 参考答案：

**迁移方案：**

**阶段一：风险识别（1-2周）**
1. **代码扫描**：使用静态分析工具找出所有跨库查询
2. **依赖图谱**：绘制服务依赖图，识别关键路径
3. **风险分级**：
   - P0：核心路径（发帖、评论、登录）
   - P1：重要路径（列表查询）
   - P2：非核心路径（运营后台、统计报表）
4. **压测评估**：评估改为 RPC 后的性能影响

**阶段二：接口准备（2-3周）**
1. **RPC 接口开发**：在被调用方提供批量查询接口
2. **兼容层开发**：在调用方开发兼容层，支持"直查"和"RPC"双模式
3. **配置开关**：使用配置中心动态切换模式
4. **数据对比**：双跑模式下对比两种方式的返回结果

**阶段三：灰度迁移（3-4周）**
1. **非核心优先**：先迁移 P2 场景
2. **小流量灰度**：按用户灰度（1% → 10% → 50%）
3. **监控验证**：重点监控错误率、延迟、耗时
4. **全量切换**：确认无误后全量切换 RPC 模式

**阶段四：彻底清理（1周）**
1. **删除兼容代码**：移除直查代码路径
2. **收回数据库权限**：撤销跨库访问权限
3. **总结沉淀**：文档化迁移经验

**回滚策略：**
1. **配置回滚**：通过配置中心一键切回直查模式
2. **灰度回滚**：先回滚部分流量，确认后全量回滚
3. **数据回滚**：如涉及数据变更，提前准备回滚脚本
4. **快速发布**：准备好回滚版本的镜像，支持快速发布

---

### 问题1.6
如果某个关键业务路径在切换后出现性能问题（比如 P99 上涨 50%），你会如何快速定位和解决？

**考察点：性能优化思路、故障排查能力**

#### 参考答案：

**排查步骤：**

**第一步：快速止血（5分钟内）**
- 立即切回旧模式（配置开关）
- 观察指标是否恢复
- 记录故障现场

**第二步：问题定位（30分钟内）**
1. **查看链路追踪**
   - 确认是哪个 RPC 调用慢
   - 看每个环节的耗时分布

2. **查看依赖服务**
   - 被调用方的负载（CPU、内存）
   - 数据库慢查询
   - 网络延迟

3. **查看应用日志**
   - 错误日志
   - 业务日志中的耗时统计

**第三步：根因分析（1小时内）**
可能的原因：
- **网络延迟**：服务间跨机房调用
- **批量不够**：改成 RPC 后没有批量，还是 N+1
- **连接池满**：RPC 连接池配置太小
- **序列化开销**：使用了低效的序列化方式
- **缓存失效**：依赖服务的缓存命中率低

**第四步：解决优化**
1. **短期方案（快速恢复）**
   - 增加 RPC 连接池
   - 启用客户端缓存
   - 服务同机房部署

2. **长期方案（根治问题）**
   - 优化批量接口设计
   - 引入数据冗余（CQRS）
   - 增加多级缓存
   - 优化序列化协议（从 JSON 改 Protobuf）

---

### 问题1.7
在迁移过程中，如何保证新旧逻辑的数据一致性？特别是涉及写入操作的场景。

**考察点：双写策略、数据补偿机制、一致性保障**

#### 参考答案：**

**一致性保障方案：**

**场景一：只读操作**
- 方案：双读对比
- 实现：同时调用旧逻辑和新逻辑，对比结果是否一致
- 不一致告警，人工介入

**场景二：写入操作**
- 方案：双写 + 最终一致
- 实现：
  1. 先写旧库（保证可用性）
  2. 再写新库（异步）
  3. 消息队列补偿（新库写入失败则重试）
  4. 定时对账任务（对比新旧数据差异）

**双写实现示例：**
```go
func MigratedCreateComment(ctx context.Context, req *CreateCommentReq) error {
    // 1. 先写旧库（保证核心路径可用）
    oldErr := oldLogic.CreateComment(ctx, req)
    if oldErr != nil {
        return oldErr
    }
    
    // 2. 异步写新库（不阻塞主流程）
    go func() {
        newErr := newLogic.CreateComment(ctx, req)
        if newErr != nil {
            // 写入失败，发消息补偿
            mq.Send(CompensationMessage{
                Type: "CreateComment",
                Data: req,
            })
        }
    }()
    
    return nil
}

// 补偿消费者
func (c *CompensationConsumer) Consume(msg *Message) error {
    switch msg.Type {
    case "CreateComment":
        return newLogic.CreateComment(ctx, msg.Data)
    // ... 其他补偿
    }
}

// 定时对账任务
func ReconciliationJob(ctx context.Context) {
    // 对比最近1小时的新旧数据差异
    diffs := compareOldAndNew(lastHour)
    for _, diff := range diffs {
        // 修复不一致
        fixInconsistency(diff)
    }
}
```

---

## 维度二：批量接口设计及 N+1 消除

### 问题2.1
当前多个读链路存在 N+1 RPC 问题，请系统地说明你会如何从设计层面、框架层面、最佳实践层面来解决这类问题。

**考察点：性能优化方法论、批量接口设计能力、框架抽象思维**

#### 参考答案：

**设计层面：**
1. **提供批量接口**：每个单查接口都要有对应的批量接口
2. **服务端聚合**：在上层服务做数据聚合，减少前端请求
3. **数据冗余**：将关联数据冗余到读服务（CQRS）
4. **图查询引擎**：复杂场景引入 GraphQL 或 DQL

**框架层面：**
1. **DataLoader 模式**：自动合并请求，批量加载
2. **请求批处理**：框架层自动收集一段时间内的请求，批量执行
3. **客户端缓存**：对热点数据做客户端本地缓存
4. **预加载机制**：在查询主数据时，预加载关联数据

**最佳实践层面：**
1. **代码审查**：Review 时重点关注 N+1 问题
2. **性能测试**：压测时检查 RPC 调用次数
3. **监控告警**：监控单接口的 RPC 调用量，超过阈值告警
4. **文档规范**：API 文档中明确推荐使用批量接口

---

### 问题2.2
你提到了批量接口，那具体到"获取帖子列表"这个场景，帖子需要关联用户信息、评论数、点赞状态等数据，你会如何设计这个聚合接口？请详细说明接口参数、返回结构、实现逻辑。

**考察点：API设计细节、数据聚合策略、性能考量**

#### 参考答案：

**接口设计：**

```protobuf
// 请求
message GetPostFeedReq {
  uint64 user_id = 1; // 当前用户ID（用于获取互动状态）
  int64 cursor = 2; // 游标分页
  int32 page_size = 3;
  repeated uint64 post_ids = 4; // 补充场景：已知ID列表
  string feed_type = 5; // "recommend" | "follow" | "hot"
}

// 帖子聚合项
message PostFeedItem {
  PostInfo post = 1;
  UserSnapshot author = 2; // 用户快照（冗余关键信息）
  repeated TopicSnapshot topics = 3; // 话题快照
  PostStats stats = 4; // 统计数据
  InteractionState interaction = 5; // 当前用户的互动状态
  repeated CommentPreview comments = 6; // 热门评论预览
}

// 响应
message GetPostFeedResp {
  repeated PostFeedItem items = 1;
  int64 next_cursor = 2;
  bool has_more = 3;
}

// 用户快照（只冗余列表页需要的字段）
message UserSnapshot {
  uint64 id = 1;
  string nickname = 2;
  string avatar = 3;
}

// 话题快照
message TopicSnapshot {
  uint64 id = 1;
  string name = 2;
}

// 统计数据
message PostStats {
  int32 comment_count = 1;
  int32 like_count = 2;
  int32 view_count = 3;
  int32 share_count = 4;
}

// 互动状态
message InteractionState {
  bool is_liked = 1;
  bool is_collected = 2;
}

// 评论预览
message CommentPreview {
  uint64 id = 1;
  string content = 2;
  UserSnapshot author = 3;
  int32 like_count = 4;
}
```

**实现逻辑：**

```go
func (l *GetPostFeedLogic) GetPostFeed(req *types.GetPostFeedReq) (*types.GetPostFeedResp, error) {
    // 1. 获取帖子列表
    posts, err := l.contentRpc.GetPostList(l.ctx, &content.GetPostListReq{
        Cursor: req.Cursor,
        PageSize: req.PageSize,
    })
    if err != nil {
        return nil, err
    }
    
    if len(posts.List) == 0 {
        return &types.GetPostFeedResp{Items: []types.PostFeedItem{}}, nil
    }
    
    // 2. 提取需要批量查询的ID
    postIds := extractPostIds(posts.List)
    userIds := extractUserIds(posts.List)
    topicIds := extractTopicIds(posts.List)
    
    // 3. 并行批量查询（使用 errgroup）
    var wg errgroup.Group
    var users map[uint64]*types.UserSnapshot
    var topics map[uint64]*types.TopicSnapshot
    var stats map[uint64]*types.PostStats
    var interactions map[uint64]*types.InteractionState
    var comments map[uint64][]*types.CommentPreview
    
    wg.Go(func() error {
        var err error
        users, err = l.userRpc.BatchGetUser(l.ctx, userIds)
        return err
    })
    
    wg.Go(func() error {
        var err error
        topics, err = l.topicRpc.BatchGetTopic(l.ctx, topicIds)
        return err
    })
    
    wg.Go(func() error {
        var err error
        stats, err = l.contentRpc.BatchGetPostStats(l.ctx, postIds)
        return err
    })
    
    wg.Go(func() error {
        var err error
        interactions, err = l.interactionRpc.BatchCheckInteraction(l.ctx, &interaction.BatchCheckReq{
            UserId: req.UserId,
            PostIds: postIds,
        })
        return err
    })
    
    wg.Go(func() error {
        var err error
        comments, err = l.commentRpc.BatchGetHotComments(l.ctx, postIds)
        return err
    })
    
    if err := wg.Wait(); err != nil {
        return nil, err
    }
    
    // 4. 聚合结果
    items := make([]types.PostFeedItem, 0, len(posts.List))
    for _, post := range posts.List {
        item := types.PostFeedItem{
            Post: post,
            Author: users[post.UserId],
            Topics: topicsForPost(post, topics),
            Stats: stats[post.Id],
            Interaction: interactions[post.Id],
            Comments: comments[post.Id],
        }
        items = append(items, item)
    }
    
    return &types.GetPostFeedResp{
        Items: items,
        NextCursor: posts.NextCursor,
        HasMore: posts.HasMore,
    }, nil
}
```

---

### 问题2.3
批量接口往往面临"数据稀疏"问题（比如请求 100 个 ID，实际只有 50 个存在），你如何处理这种情况？这对缓存策略有什么影响？

**考察点：边界情况处理、缓存设计能力**

#### 参考答案：

**数据稀疏的处理策略：**

**方案1：返回 Map（推荐）**
```protobuf
message BatchGetPostResp {
  map<uint64, PostInfo> posts = 1; // key是postId，不存在的key不返回
}
```
- 优点：调用方可以快速判断某个 ID 是否存在
- 缺点：Protobuf 的 map 在某些语言中使用不太方便

**方案2：返回列表 + 状态**
```protobuf
message PostResult {
  uint64 id = 1;
  bool exists = 2;
  PostInfo post = 3; // exists=true时才有值
}

message BatchGetPostResp {
  repeated PostResult results = 1;
}
```
- 优点：顺序与请求一致，调用方容易处理
- 缺点：数据量稍大

**方案3：只返回存在的，忽略不存在的**
```protobuf
message BatchGetPostResp {
  repeated PostInfo posts = 1; // 只返回存在的，顺序不保证
}
```
- 优点：数据量最小
- 缺点：调用方需要自己做 ID 匹配

**缓存策略影响：**

**问题：不存在的 ID 要不要缓存？**
- **要缓存**，原因：
  1. 防止缓存穿透（攻击者大量请求不存在的 ID）
  2. 减少数据库压力
  3. 性能更好

**缓存不存在数据的方案：**
```go
func (l *BatchGetPostLogic) BatchGetPost(req *types.BatchGetPostReq) (*types.BatchGetPostResp, error) {
    result := make(map[uint64]*types.PostInfo)
    missedIds := make([]uint64, 0)
    
    // 1. 批量查缓存
    cached, err := l.cache.MGet(req.Ids)
    if err != nil {
        return nil, err
    }
    
    // 2. 分离命中和未命中
    for _, id := range req.Ids {
        if val, ok := cached[id]; ok {
            if val == "null" {
                // 缓存了不存在的标记
                continue
            }
            result[id] = deserialize(val)
        } else {
            missedIds = append(missedIds, id)
        }
    }
    
    // 3. 未命中的回源
    if len(missedIds) > 0 {
        dbResult, err := l.model.BatchFind(missedIds)
        if err != nil {
            return nil, err
        }
        
        // 构建存在的 ID 集合
        existsSet := make(map[uint64]bool)
        for _, post := range dbResult {
            result[post.Id] = post
            existsSet[post.Id] = true
            // 缓存存在的数据
            l.cache.SetEX(fmt.Sprintf("post:%d", post.Id), serialize(post), time.Hour)
        }
        
        // 缓存不存在的数据（短 TTL）
        for _, id := range missedIds {
            if !existsSet[id] {
                l.cache.SetEX(fmt.Sprintf("post:%d", id), "null", 5*time.Minute)
            }
        }
    }
    
    return &types.BatchGetPostResp{Posts: result}, nil
}
```

---

### 问题2.4
在 go-zero 框架下，你有没有想过做一些框架层面的增强，来自动避免 N+1 问题？比如某种拦截器或者代码生成层面的优化？

**考察点：框架理解深度、工程化创新能力**

#### 参考答案：

**框架增强方案：**

**方案1：DataLoader 封装（推荐）**

```go
// DataLoader 批量加载器
type DataLoader[K comparable, V any] struct {
    batchFn func([]K) (map[K]V, error)
    maxBatch int
    timeout time.Duration
    mu sync.Mutex
    pending []K
    results chan map[K]V
}

func NewDataLoader[K comparable, V any](
    batchFn func([]K) (map[K]V, error),
    opts ...Option,
) *DataLoader[K, V] {
    dl := &DataLoader[K, V]{
        batchFn: batchFn,
        maxBatch: 100,
        timeout: 100 * time.Millisecond,
    }
    for _, opt := range opts {
        opt(dl)
    }
    return dl
}

func (dl *DataLoader[K, V]) Load(ctx context.Context, key K) (V, error) {
    dl.mu.Lock()
    dl.pending = append(dl.pending, key)
    
    // 如果达到批量大小，立即执行
    if len(dl.pending) >= dl.maxBatch {
        keys := dl.pending
        dl.pending = nil
        dl.mu.Unlock()
        return dl.doLoad(ctx, keys)[key], nil
    }
    
    // 否则等待 timeout 或批量满
    resultCh := make(chan map[K]V, 1)
    dl.results = resultCh
    dl.mu.Unlock()
    
    select {
    case <-time.After(dl.timeout):
        dl.mu.Lock()
        keys := dl.pending
        dl.pending = nil
        dl.mu.Unlock()
        results := dl.doLoad(ctx, keys)
        return results[key], nil
    case results := <-resultCh:
        return results[key], nil
    case <-ctx.Done():
        return *new(V), ctx.Err()
    }
}

func (dl *DataLoader[K, V]) doLoad(ctx context.Context, keys []K) map[K]V {
    results, _ := dl.batchFn(keys)
    return results
}

// 使用示例
func (l *GetPostListLogic) GetPostList(req *types.GetPostListReq) {
    // 创建用户 DataLoader
    userLoader := NewDataLoader(func(ids []uint64) (map[uint64]*types.User, error) {
        return l.userRpc.BatchGetUser(l.ctx, ids)
    })
    
    for _, post := range posts {
        // 单个加载，内部会自动批量
        user, _ := userLoader.Load(l.ctx, post.UserId)
        // ...
    }
}
```

**方案2：goctl 代码生成增强**

修改 goctl，在生成 rpc 代码时，自动为每个单查接口生成批量接口：

```go
// 自动生成的批量接口
func (l *BatchGetUserLogic) BatchGetUser(req *types.BatchGetUserReq) (*types.BatchGetUserResp, error) {
    // 自动生成的批量查询逻辑
}
```

**方案3：静态分析工具**

开发一个 linter，检查代码中的 N+1 问题：

```go
// Bad：循环内调用 RPC
for _, id := range ids {
    user, _ := userRpc.GetUser(id) // N+1 问题！
}

// Good：使用批量接口
users, _ := userRpc.BatchGetUser(ids)
```

---

## 维度三：搜索索引模型重构及零停机回填

### 问题3.1
当前 ES 索引字段不全，需要二次补字段，效率低下。请设计一个新的索引模型，并说明如何零停机地完成全量数据回填和切换。

**考察点：ES建模能力、数据迁移方案、高可用保障**

#### 参考答案：

**ES 索引模型设计：**

```json
{
  "mappings": {
    "properties": {
      "post_id": {"type": "long"},
      "user_id": {"type": "long"},
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {"type": "keyword"}
        }
      },
      "content": {"type": "text", "analyzer": "ik_max_word"},
      "cover": {"type": "keyword", "index": false},
      "author": {
        "properties": {
          "id": {"type": "long"},
          "nickname": {"type": "text"},
          "avatar": {"type": "keyword", "index": false}
        }
      },
      "topics": {"type": "keyword"},
      "tags": {"type": "keyword"},
      "stats": {
        "properties": {
          "comment_count": {"type": "integer"},
          "like_count": {"type": "integer"},
          "view_count": {"type": "integer"}
        }
      },
      "visibility": {"type": "integer"},
      "is_top": {"type": "boolean"},
      "is_essence": {"type": "boolean"},
      "created_at": {"type": "date"},
      "updated_at": {"type": "date"},
      "hot_score": {"type": "float"} // 用于排序的热度分
    }
  },
  "settings": {
    "number_of_shards": 6,
    "number_of_replicas": 2,
    "refresh_interval": "30s" // 写入优化
  }
}
```

**设计要点：**
- 用户、话题、统计数据冗余进索引，无需二次查询
- 关键词用 keyword，文本用 text+ik
- 不需要检索的字段设为 index: false
- title 同时建 keyword 子字段，用于精确匹配和排序

**零停机回填方案：**

**架构：**
```
MySQL → Canal → Kafka → Indexer → ES
              ↓
         回填任务
```

**步骤：**
1. **创建新索引**：建新索引 post_v2，保留旧索引 post_v1
2. **开启双写**：新写入同时进 v1 和 v2（保证增量不丢）
3. **全量回填**：启动回填任务，从 MySQL 读取历史数据写入 v2
4. **数据校验**：回填完成后，抽样验证数据正确性
5. **灰度切换**：索引别名 post_alias 先切 1% 流量到 v2
6. **全量切换**：确认无误后全量切换到 v2
7. **下线旧索引**：观察稳定后，下线 v1 索引

---

### 问题3.2
你提到了"索引别名"做零停机切换，那如果在回填过程中，数据还在持续写入，如何保证新索引的数据是最新的？

**考察点：增量同步策略、数据一致性保障**

#### 参考答案：

**增量同步 + 全量回填方案：**

**方案：**

```
┌─────────────────────────────────────────────────────────┐
│                     时间线                                │
├─────────────┬───────────────────────────────────────────┤
│  T0         │  创建索引 v2，开启双写                       │
│             │  增量写入：同时写 v1 和 v2                  │
├─────────────┼───────────────────────────────────────────┤
│  T1 ~ T2    │  全量回填历史数据到 v2                      │
│             │  （回填范围：T0 之前的数据）                 │
├─────────────┼───────────────────────────────────────────┤
│  T2         │  回填完成，校验数据                         │
├─────────────┼───────────────────────────────────────────┤
│  T3         │  切换别名，流量切到 v2                      │
└─────────────┴───────────────────────────────────────────┘
```

**关键实现：**

**1. 基于时间点的全量回填**
```go
func BackfillTask(ctx context.Context, startTime time.Time) error {
    // 只回填 startTime 之前的数据
    batchSize := 1000
    lastId := uint64(0)
    
    for {
        // 范围查询：id > lastId AND created_at < startTime
        posts, err := db.Query(ctx, `
            SELECT * FROM posts 
            WHERE id > ? AND created_at < ? 
            ORDER BY id LIMIT ?
        `, lastId, startTime, batchSize)
        
        if len(posts) == 0 {
            break
        }
        
        // 写入 ES
        if err := bulkWriteToES(ctx, posts); err != nil {
            return err
        }
        
        lastId = posts[len(posts)-1].Id
        // 记录进度，支持断点续传
        saveProgress(lastId)
    }
    return nil
}
```

**2. 增量同步（Canal + Kafka）**
```go
func (c *CanalConsumer) Consume(msg *CanalMessage) error {
    switch msg.EventType {
    case "INSERT":
        return indexer.Index(msg.Post)
    case "UPDATE":
        return indexer.Update(msg.Post)
    case "DELETE":
        return indexer.Delete(msg.PostId)
    }
    return nil
}
```

**3. 数据校验（回填后）**
```go
func ValidateData(ctx context.Context, startTime time.Time) error {
    // 随机抽样校验
    sampleSize := 1000
    sampleIds := randomSampleIds(sampleSize)
    
    for _, id := range sampleIds {
        // 从 MySQL 读取
        dbPost, _ := db.Get(ctx, id)
        // 从 ES 读取
        esPost, _ := es.Get(ctx, id)
        
        // 对比数据（注意：created_at >= startTime 的可能有延迟）
        if dbPost.CreatedAt.Before(startTime) {
            if !equals(dbPost, esPost) {
                return errors.New("数据不一致")
            }
        }
    }
    return nil
}
```

**4. 冲突解决（时间戳为最后胜利）**
```go
// 索引文档时带上时间戳
doc := map[string]interface{}{
    "post_id": post.Id,
    "updated_at": post.UpdatedAt,
    // ...
}

// 使用 version_type=external_gte，只更新更新的数据
es.Update(ctx, doc, elasticsearch.WithVersionType("external_gte"))
```

---

### 问题3.3
全量回填通常耗时较长，你会如何优化回填速度？同时又如何避免对线上数据库造成过大压力？

**考察点：性能优化、限流降级、资源隔离**

#### 参考答案：

**回填优化策略：**

**一、速度优化**

**1. 并发回填**
```go
// 按 ID 范围分片，并发回填
shards := 10
var wg sync.WaitGroup

for i := 0; i < shards; i++ {
    wg.Add(1)
    go func(shardId int) {
        defer wg.Done()
        // 只回填 id % shards == shardId 的数据
        backfillShard(ctx, shardId, shards)
    }(i)
}
wg.Wait()
```

**2. 批量写入**
```go
// ES Bulk API，每次批量写入 500-1000 条
bulkRequest := es.Bulk()
for _, post := range posts {
    bulkRequest.Add(elasticsearch.NewBulkIndexRequest().
        Index("post_v2").
        Doc(post))
}
if _, err := bulkRequest.Do(ctx); err != nil {
    // 处理错误
}
```

**3. ES 写入优化**
```json
{
  "settings": {
    "refresh_interval": "-1", // 回填期间关闭刷新
    "number_of_replicas": 0, // 回填期间关闭副本
    "index.translog.durability": "async"
  }
}
```
回填完成后再改回来：
```json
{
  "settings": {
    "refresh_interval": "30s",
    "number_of_replicas": 2
  }
}
```

**二、压力控制**

**1. 数据库限流**
```go
// 使用令牌桶限流
limiter := rate.NewLimiter(1000, 1000) // 每秒 1000 次

for {
    if !limiter.Allow() {
        time.Sleep(10 * time.Millisecond)
        continue
    }
    // 查询数据库
}
```

**2. 按时间切片**
```go
// 每次只回填 1 小时的数据
startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
endTime := startTime.Add(time.Hour)

for endTime.Before(time.Now()) {
    backfillRange(ctx, startTime, endTime)
    startTime = endTime
    endTime = endTime.Add(time.Hour)
}
```

**3. 低峰期执行**
```go
// 只在凌晨 2-6 点执行回填
func canRun() bool {
    hour := time.Now().Hour()
    return hour >= 2 && hour < 6
}
```

**4. 资源隔离**
- 回填任务独立部署
- 使用单独的数据库从库查询
- ES 节点标记为 warm 节点，不影响热数据

---

### 问题3.4
假设新索引上线后，发现搜索效果不符合预期，你会如何回滚？回滚过程中如何保证这段时间的写入不丢失？

**考察点：回滚预案、数据补偿、风险意识**

#### 参考答案：

**回滚方案：**

**步骤一：快速回滚**
1. **切换索引别名**：将 post_alias 立即切回 post_v1
2. **验证回滚成功**：确认搜索请求都到 v1
3. **保留双写**：新写入仍然同时写 v1 和 v2（防止 v2 数据丢失）

**步骤二：数据补偿**
```go
// 记录切换时间点
rollbackTime := time.Now()

// 回滚期间，v2 仍然接收写入（通过双写）
// 等问题修复后，可以：
// 1. 从 rollbackTime 开始重新回填到 v2
// 2. 或者直接从 Kafka 消息回放

// 消息回放补偿
func ReplayMessages(startTime time.Time) error {
    // 从 Kafka 消费 startTime 之后的消息
    consumer := kafka.NewConsumer(startTime)
    for {
        msg, err := consumer.ReadMessage()
        if err != nil {
            break
        }
        // 重新写入 v2
        indexer.Index(msg.Post)
    }
    return nil
}
```

**步骤三：修复问题**
- 分析搜索效果不好的原因（分词问题？字段缺失？评分问题？）
- 修复索引模型或查询逻辑
- 重新创建 post_v3 索引

**步骤四：重新上线**
- 重复之前的流程：双写 → 回填 → 校验 → 切换

**回滚前的准备：**
1. **保留旧索引**：不要急于删除 v1，至少保留 7 天
2. **记录切换时间**：精确记录每次切换的时间点
3. **消息持久化**：Kafka 消息保留足够长时间（比如 3 天）
4. **回滚脚本**：提前写好一键回滚脚本

---

### 问题3.5
搜索服务经常面临"召回率"和"精确率"的权衡，在社交内容场景中，你如何设计搜索策略？

**考察点：搜索策略、排序算法、业务理解**

#### 参考答案：

**搜索策略设计：**

**一、多阶段召回**

```
用户查询 → 粗召回（快速筛选候选集）→ 精排（机器学习排序）→ 重排（业务规则）
```

**1. 粗召回（Elasticsearch）**
- 标题匹配（高权重）
- 内容匹配（中等权重）
- 话题匹配（中等权重）
- 召回 1000-2000 个候选

**2. 精排（机器学习模型）**
- 特征：文本匹配度、点击率、互动率、发布时间、作者质量
- 模型：LR / GBDT / 深度学习
- 排序输出 Top 100

**3. 重排（业务规则）**
- 多样性：同一作者的帖子不要太集中
- 新鲜度：新帖子适当提权
- 质量：优质内容（加精、置顶）优先

**二、冷热数据分离**
```
热数据（最近7天）→ 快索引（内存优化）
冷数据（7天前）  → 慢索引（磁盘优化）
```

**三、索引分层策略**
```
索引 A：核心字段（标题、作者、时间）- 小而快
索引 B：全量字段（包含内容）- 大而全
```
查询时先查 A，拿到 ID 后再查 B 补全字段

**四、业务场景定制**

| 场景 | 召回策略 | 排序策略 |
|------|----------|----------|
| 搜索框 | 文本相关性为主 | 相关性 + 热度 |
| 话题页 | 精确匹配话题 | 时间倒序 + 互动 |
| 个人主页 | 精确匹配作者 | 时间倒序 |
| 推荐流 | 多路召回 | 机器学习模型 |

---

## 维度四：异步消息可靠性

### 问题4.1
当前 Kafka 消息缺乏统一 envelope、幂等、重试和 DLQ 机制。请设计一套完整的消息中间件解决方案。

**考察点：消息系统设计能力、可靠性保障机制、可观测性**

#### 参考答案：

**消息架构设计：**

```
生产者 → Kafka Topic → 消费者 → 业务处理
                        ↓
                    失败重试（3次）
                        ↓
                    失败 → DLQ → 告警 → 人工处理
```

**核心组件：**
1. **统一 Envelope**：消息统一格式
2. **幂等处理器**：保证消息不重复处理
3. **重试机制**：指数退避重试
4. **DLQ 处理**：死信队列处理流程
5. **可观测性**：监控、链路追踪

---

### 问题4.2
你如何设计消息的 Envelope 结构？需要包含哪些元数据字段？为什么？

**考察点：协议设计能力、可扩展性考虑**

#### 参考答案：

**Envelope 设计：**

```protobuf
syntax = "proto3";

package message;

option go_package = "message";

// 消息信封
message MessageEnvelope {
  // 基础元数据
  string id = 1; // 消息唯一ID（UUID），用于幂等
  string type = 2; // 消息类型（如 "post.created"）
  string source = 3; // 来源服务（如 "content-service"）
  int64 created_at = 4; // 创建时间（毫秒时间戳）
  int64 occurred_at = 5; // 业务发生时间
  
  // 重试元数据
  int32 retry_count = 6; // 已重试次数
  int32 max_retry = 7; // 最大重试次数
  int64 next_retry_at = 8; // 下次重试时间
  
  // 链路追踪
  string trace_id = 9; // 链路追踪ID
  string span_id = 10; // Span ID
  
  // 业务数据
  string payload_schema = 11; // Payload Schema（如 "v1"）
  bytes payload = 12; // 业务数据（Protobuf/JSON）
  
  // 扩展元数据（预留）
  map<string, string> metadata = 13;
}
```

**各字段设计理由：**

| 字段 | 理由 |
|------|------|
| id | 幂等校验的关键，必须全局唯一 |
| type | 路由到不同的消费者处理器 |
| source | 便于排查问题，知道消息来源 |
| created_at/occurred_at | 业务时间可能和发送时间不同 |
| retry_count/max_retry | 控制重试次数，避免无限重试 |
| next_retry_at | 延迟重试，避免立即重试压垮下游 |
| trace_id/span_id | 可观测性，链路追踪 |
| payload_schema | 支持 Payload 版本升级 |
| metadata | 灵活扩展，不用改 Envelope 结构 |

**使用示例：**

```go
// 生产者发送消息
func SendPostCreatedMessage(ctx context.Context, post *Post) error {
    payload, _ := proto.Marshal(&PostCreatedEvent{Post: post})
    
    msg := &MessageEnvelope{
        Id: uuid.New().String(),
        Type: "post.created",
        Source: "content-service",
        CreatedAt: time.Now().UnixMilli(),
        OccurredAt: post.CreatedAt.UnixMilli(),
        MaxRetry: 3,
        TraceId: trace.FromContext(ctx).TraceId(),
        SpanId: trace.FromContext(ctx).SpanId(),
        PayloadSchema: "v1",
        Payload: payload,
    }
    
    return kafkaProducer.Send(ctx, "post-events", msg)
}

// 消费者处理消息
func (c *Consumer) HandleMessage(ctx context.Context, msg *MessageEnvelope) error {
    // 1. 恢复链路追踪
    ctx = trace.WithContext(ctx, msg.TraceId, msg.SpanId)
    
    // 2. 幂等校验
    if c.idempotentChecker.HasProcessed(ctx, msg.Id) {
        return nil // 已处理，直接返回
    }
    
    // 3. 路由到对应的处理器
    switch msg.Type {
    case "post.created":
        var event PostCreatedEvent
        proto.Unmarshal(msg.Payload, &event)
        return c.handlePostCreated(ctx, &event)
    // ...
    }
    
    return nil
}
```

---

### 问题4.3
幂等性是消息系统的核心难点，请列举至少三种不同的幂等实现方案，并分别说明它们的适用场景、优劣势。

**考察点：技术选型能力、场景化思考**

#### 参考答案：

**幂等实现方案对比：**

| 方案 | 实现方式 | 优点 | 缺点 | 适用场景 |
|------|----------|------|------|----------|
| 数据库唯一键 | 利用业务唯一约束建索引 | 简单可靠，天然幂等 | 依赖数据库、有额外写入开销 | 业务本身有唯一键 |
| 幂等表 | 单独建表记录已处理消息 | 通用性强，不侵入业务 | 额外存储、需清理历史数据 | 通用场景 |
| Redis去重 | SETNX或记录处理状态 | 高性能，低延迟 | 可能丢数据、需考虑持久化 | 高吞吐、可接受小概率重复 |
| 状态机校验 | 校验当前状态是否合法 | 业务语义化、一致性强 | 需设计状态流转 | 有状态的业务 |
| 乐观锁 | 版本号机制 | 性能好、不阻塞 | 冲突时需重试 | 更新操作 |

**方案详解：**

**方案1：数据库唯一键（推荐用于有业务唯一键的场景）**
```sql
-- 例如：点赞表，(user_id, post_id) 是唯一键
CREATE TABLE likes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT,
    post_id BIGINT,
    created_at DATETIME,
    UNIQUE KEY uk_user_post (user_id, post_id)
);
```
```go
// 插入时，如果重复会报错，捕获错误即可
func LikePost(ctx context.Context, userId, postId uint64) error {
    _, err := db.Exec(ctx, `
        INSERT INTO likes (user_id, post_id, created_at) 
        VALUES (?, ?, NOW())
    `, userId, postId)
    
    if err != nil && mysqlErrCode(err) == 1062 { // 唯一键冲突
        return nil // 已点赞，幂等返回成功
    }
    return err
}
```

**方案2：幂等表（通用方案）**
```sql
CREATE TABLE idempotent_records (
    id VARCHAR(64) PRIMARY KEY, -- 消息ID
    handler VARCHAR(64), -- 处理器名称
    created_at DATETIME,
    KEY idx_created_at (created_at) -- 用于清理历史数据
);
```
```go
func (c *IdempotentChecker) CheckAndMark(ctx context.Context, msgId string, handler string) (bool, error) {
    // 尝试插入
    _, err := db.Exec(ctx, `
        INSERT INTO idempotent_records (id, handler, created_at) 
        VALUES (?, ?, NOW())
    `, msgId, handler)
    
    if err != nil && mysqlErrCode(err) == 1062 {
        return true, nil // 已处理
    }
    return false, err
}

// 使用
func (c *Consumer) HandleMessage(ctx context.Context, msg *MessageEnvelope) error {
    processed, err := c.idempotentChecker.CheckAndMark(ctx, msg.Id, "comment-consumer")
    if err != nil {
        return err
    }
    if processed {
        return nil
    }
    // 处理业务
    return c.doHandle(ctx, msg)
}
```

**方案3：Redis 去重（高性能场景）**
```go
func (c *RedisIdempotentChecker) HasProcessed(ctx context.Context, msgId string) (bool, error) {
    // SETNX：设置成功返回 1（未处理），返回 0（已处理）
    key := fmt.Sprintf("idempotent:%s", msgId)
    result, err := redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
    if err != nil {
        return false, err
    }
    return !result, nil // result=true 表示未处理，返回 false
}
```

**方案4：状态机校验（有状态业务）**
```go
type OrderState string

const (
    OrderStateCreated OrderState = "created"
    OrderStatePaid OrderState = "paid"
    OrderStateShipped OrderState = "shipped"
)

// 状态流转规则
var stateTransitions = map[OrderState][]OrderState{
    OrderStateCreated: {OrderStatePaid},
    OrderStatePaid: {OrderStateShipped},
}

func (s *OrderService) PayOrder(ctx context.Context, orderId uint64) error {
    order, _ := s.GetOrder(ctx, orderId)
    
    // 状态校验
    if order.State != OrderStateCreated {
        return nil // 已经支付过，幂等返回
    }
    
    // 更新状态
    return s.UpdateOrderState(ctx, orderId, OrderStatePaid)
}
```

---

### 问题4.4
请设计一个 DLQ（死信队列）的处理流程。消息进入 DLQ 后，人工介入修复后，如何重新投递？重新投递时如何保证不影响正常流量？

**考察点：异常处理机制、操作流程设计**

#### 参考答案：

**DLQ 处理流程设计：**

```
正常消费 → 消费失败 → 重试（指数退避，3次） → 超过重试次数 → 进入 DLQ
                                                                        ↓
                                                                  告警通知
                                                                        ↓
                                                                  人工介入
                                                                        ↓
                                                                  修复数据/代码
                                                                        ↓
                                                                  重新投递（补发队列）
```

**架构图：**

```
┌─────────────┐
│   主 Topic   │
└──────┬──────┘
       │
       ▼
┌──────────────────┐
│   消费者         │─┐
└──────┬───────────┘ │
       │             │ 重试
       ▼             │
┌───────────────┐    │
│   重试 Topic  │◄───┘
└───────┬───────┘
        │
        ▼ (超过重试次数)
┌───────────────┐
│    DLQ Topic  │
└───────┬───────┘
        │
        ▼
┌──────────────────┐
│   DLQ 消费者     │ ──→ 告警（邮件/钉钉/PagerDuty）
└───────┬──────────┘
        │
        ▼
┌──────────────────┐
│   管理后台       │ ──→ 查看、修复、重发
└───────┬──────────┘
        │
        ▼
┌───────────────┐
│   补发 Topic  │
└───────┬───────┘
        │
        ▼
┌──────────────────┐
│   补发消费者     │ ──→ 限流、独立部署
└──────────────────┘
```

**实现代码：**

**1. 消费者主逻辑**
```go
func (c *Consumer) HandleMessage(ctx context.Context, msg *MessageEnvelope) error {
    // 处理消息
    err := c.doHandle(ctx, msg)
    if err == nil {
        return nil // 处理成功
    }
    
    // 处理失败，判断是否重试
    if msg.RetryCount < msg.MaxRetry && c.isRetriable(err) {
        // 发送到重试 Topic（延迟消息）
        msg.RetryCount++
        msg.NextRetryAt = time.Now().Add(c.backoff(msg.RetryCount)).UnixMilli()
        return c.retryProducer.Send(ctx, msg)
    }
    
    // 不可重试或超过重试次数，发送到 DLQ
    msg.Metadata["error"] = err.Error()
    return c.dlqProducer.Send(ctx, msg)
}

// 指数退避
func (c *Consumer) backoff(retryCount int32) time.Duration {
    return time.Duration(math.Pow(2, float64(retryCount))) * time.Second
}

// 判断是否可重试
func (c *Consumer) isRetriable(err error) bool {
    switch err.(type) {
    case *network.Error:
        return true // 网络错误可重试
    case *db.DeadlockError:
        return true // 死锁可重试
    default:
        return false // 业务错误不可重试
    }
}
```

**2. DLQ 消息存储和管理**
```sql
CREATE TABLE dlq_messages (
    id VARCHAR(64) PRIMARY KEY,
    topic VARCHAR(64),
    message_type VARCHAR(64),
    payload BLOB,
    error_msg TEXT,
    created_at DATETIME,
    processed_at DATETIME NULL,
    status ENUM('pending', 'processing', 'fixed', 'ignored'),
    INDEX idx_status_created (status, created_at)
);
```

**3. 重新投递**
```go
func (c *DLQService) Redeliver(ctx context.Context, msgId string) error {
    // 从数据库读取 DLQ 消息
    msg, _ := c.GetDLQMessage(ctx, msgId)
    
    // 标记为处理中
    c.UpdateStatus(ctx, msgId, "processing")
    
    // 发送到补发 Topic（不是主 Topic！）
    err := c.republishProducer.Send(ctx, "dlq-republish", msg)
    if err != nil {
        c.UpdateStatus(ctx, msgId, "pending")
        return err
    }
    
    return nil
}

// 补发消费者（独立部署，限流）
func (c *RepublishConsumer) HandleMessage(ctx context.Context, msg *MessageEnvelope) error {
    // 限流：每秒最多处理 10 条
    if !c.limiter.Allow() {
        time.Sleep(100 * time.Millisecond)
        return errors.New("rate limited") // 会重试
    }
    
    // 调用正常的业务处理逻辑
    err := c.normalHandler.HandleMessage(ctx, msg)
    if err == nil {
        // 处理成功，标记 DLQ 消息为已修复
        c.dlqService.UpdateStatus(ctx, msg.Id, "fixed")
    }
    return err
}
```

**4. 告警规则**
- DLQ 消息数 > 0，立即告警
- DLQ 消息数 > 100，升级告警
- 每小时汇总 DLQ 报告

---

### 问题4.5
你考虑过"事件溯源"吗？在这个社交内容系统中，哪些场景适合引入事件溯源？它会带来什么额外的复杂度？

**考察点：架构模式理解、业务场景匹配能力**

#### 参考答案：

**事件溯源（Event Sourcing）简介：**

不直接保存当前状态，而是保存所有状态变更的事件。当前状态通过重放事件得到。

```
传统方式：
用户表 → {id: 1, nickname: "张三", level: 5}

事件溯源：
事件表 → [
    {type: "UserCreated", data: {id: 1, nickname: "张三"}},
    {type: "LevelUp", data: {userId: 1, newLevel: 2}},
    {type: "LevelUp", data: {userId: 1, newLevel: 3}},
    {type: "NicknameUpdated", data: {userId: 1, newNickname: "李四"}},
    {type: "LevelUp", data: {userId: 1, newLevel: 4}},
    {type: "LevelUp", data: {userId: 1, newLevel: 5}}
]
```

**适合引入事件溯源的场景：**

**1. 积分/成长值系统（推荐）**
- 需要审计：每一笔积分变更都可追溯
- 需要回滚：误操作可以回滚到之前的状态
- 需要对账：与第三方支付系统对账

**2. 互动行为（点赞、收藏）**
- 需要统计：用户的点赞历史
- 需要恢复：误操作可以恢复
- 需要分析：分析用户的互动行为模式

**3. 内容审核流程**
- 需要查看审核历史：谁审核的、什么时候审核的
- 需要回滚：误审可以恢复
- 需要统计：审核人员的工作统计

**不适合的场景：**
- 帖子内容（频繁更新，重放成本高）
- 用户基本信息（隐私敏感，不需要全量历史）

**社交系统中的事件示例：**

```protobuf
// 事件定义
message UserLevelUpEvent {
    uint64 user_id = 1;
    int32 old_level = 2;
    int32 new_level = 3;
    string reason = 4; // "posted" | "liked" | "admin"
    int64 occurred_at = 5;
}

message PostLikedEvent {
    uint64 user_id = 1;
    uint64 post_id = 2;
    int64 occurred_at = 3;
}

message PostUnlikedEvent {
    uint64 user_id = 1;
    uint64 post_id = 2;
    int64 occurred_at = 3;
}
```

**事件溯源带来的复杂度：**

| 复杂度 | 说明 | 应对方案 |
|--------|------|----------|
| 存储成本 | 事件数据会越来越大 | 定期归档、快照 |
| 性能问题 | 每次读取都要重放事件 | 快照 + CQRS |
| 事件版本 | 事件结构变化如何兼容 | 事件升级器、版本号 |
| 乱序问题 | 事件可能不是按顺序到达 | 基于时间戳合并、幂等 |
| 调试困难 | 问题排查需要看事件流 | 事件可视化工具 |

**实现示例（积分系统）：**

```go
// 事件存储
type EventStore struct {
    db *sql.DB
}

func (es *EventStore) AppendEvent(ctx context.Context, event interface{}) error {
    // 序列化事件
    payload, _ := json.Marshal(event)
    
    // 存储事件
    _, err := es.db.Exec(ctx, `
        INSERT INTO events (user_id, type, payload, occurred_at)
        VALUES (?, ?, ?, ?)
    `, event.UserId, event.Type, payload, event.OccurredAt)
    
    return err
}

// 重放事件，构建当前状态
func (es *EventStore) ReplayUserState(ctx context.Context, userId uint64) (*UserState, error) {
    events, _ := es.GetEvents(ctx, userId)
    
    state := &UserState{UserId: userId}
    for _, event := range events {
        state.Apply(event) // 应用事件
    }
    return state, nil
}

// 快照（优化性能）
func (es *EventStore) SaveSnapshot(ctx context.Context, userId uint64, state *UserState) error {
    // 定期保存快照
    // 重放时：快照 + 之后的事件
}
```

---

## 维度五：缓存治理

### 问题5.1
当前缓存使用零散，无热点保护和回源限流。请设计一套完整的缓存治理方案，包括缓存使用规范、防击穿/穿透/雪崩策略、一致性保障机制。

**考察点：缓存设计方法论、风险防控能力、一致性权衡**

#### 参考答案：

**缓存治理方案：**

**一、缓存使用规范**

**1. Key 设计规范**
```go
// 格式：{service}:{entity}:{id}:{version}
// 示例：
key := fmt.Sprintf("content:post:%d:v1", postId)           // 帖子
key := fmt.Sprintf("user:info:%d:v1", userId)               // 用户信息
key := fmt.Sprintf("interaction:like:%d:%d", userId, postId) // 点赞状态
```

**2. Value 序列化**
- 推荐：Protobuf（性能好、二进制小）
- 备选：MessagePack
- 不推荐：JSON（性能差、体积大）

**3. 过期时间设置**
```go
// 基础数据（用户信息）：1小时
userCacheTTL = 1 * time.Hour

// 内容数据（帖子）：30分钟
postCacheTTL = 30 * time.Minute

// 统计数据（点赞数）：5分钟
statsCacheTTL = 5 * time.Minute

// 配置数据：10分钟，后台变更主动刷新
configCacheTTL = 10 * time.Minute
```

**4. 缓存命名空间**
- 不同环境使用不同的 Redis DB 或 Key 前缀
- 测试环境：`test:` 前缀
- 生产环境：`prod:` 前缀

**二、防缓存问题**

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 缓存穿透 | 查询不存在的数据 | 缓存空值、布隆过滤器 |
| 缓存击穿 | 热点 Key 失效 | 互斥锁、永不过期、预热 |
| 缓存雪崩 | 大量 Key 同时失效 | 过期时间加随机值、多级缓存 |

**三、一致性保障策略**

推荐：Cache Aside Pattern（旁路缓存）

**写操作：先删缓存，再更新数据库**
```go
func (s *PostService) UpdatePost(ctx context.Context, post *Post) error {
    // 1. 先删缓存
    cacheKey := fmt.Sprintf("content:post:%d:v1", post.Id)
    s.cache.Del(ctx, cacheKey)
    
    // 2. 再更新数据库
    if err := s.db.Update(ctx, post); err != nil {
        return err
    }
    
    // 3. 异步延迟双删（防止并发问题）
    go func() {
        time.Sleep(500 * time.Millisecond)
        s.cache.Del(ctx, cacheKey)
    }()
    
    return nil
}
```

**读操作：先读缓存，没命中回源，回源后写缓存**
```go
func (s *PostService) GetPost(ctx context.Context, postId uint64) (*Post, error) {
    cacheKey := fmt.Sprintf("content:post:%d:v1", postId)
    
    // 1. 读缓存
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil && cached != "" {
        if cached == "null" {
            return nil, errors.New("post not found")
        }
        return deserializePost(cached), nil
    }
    
    // 2. 回源（加分布式锁防止缓存击穿）
    lockKey := fmt.Sprintf("lock:post:%d", postId)
    lock := s.distributedLock.Acquire(ctx, lockKey, 5*time.Second)
    defer lock.Release()
    
    // 再次检查缓存（DCL）
    cached, err = s.cache.Get(ctx, cacheKey)
    if err == nil && cached != "" {
        if cached == "null" {
            return nil, errors.New("post not found")
        }
        return deserializePost(cached), nil
    }
    
    // 3. 查数据库
    post, err := s.db.FindById(ctx, postId)
    if err != nil {
        return nil, err
    }
    
    // 4. 写缓存（空结果也缓存）
    if post == nil {
        s.cache.SetEX(ctx, cacheKey, "null", 1*time.Minute)
        return nil, errors.New("post not found")
    }
    
    s.cache.SetEX(ctx, cacheKey, serializePost(post), 30*time.Minute)
    return post, nil
}
```

**四、回源限流**
```go
// 使用令牌桶限流，保护数据库
limiter := rate.NewLimiter(100, 100) // 每秒 100 次回源

func (s *PostService) GetPost(ctx context.Context, postId uint64) (*Post, error) {
    // ... 读缓存 ...
    
    if cacheMiss {
        // 限流检查
        if !limiter.Allow() {
            // 触发限流，返回降级数据
            return s.getFallbackPost(postId), nil
        }
        // ... 回源 ...
    }
}
```

---

### 问题5.2
热点 Key 是社交系统的常见问题（比如某个热门帖子），你如何发现和治理热点 Key？请分别从"事前预防"、"事中发现"、"事后治理"三个阶段说明。

**考察点：热点治理经验、全流程思维**

#### 参考答案：

**热点 Key 治理方案：**

**一、事前预防**

**1. 热点 Key 预判与预热**
```go
// 运营后台设置的热门帖子，提前预热到缓存
func PreheatHotPosts(ctx context.Context, postIds []uint64) error {
    for _, id := range postIds {
        post, _ := db.Get(ctx, id)
        cache.SetEX(ctx, key, serialize(post), 1*time.Hour)
    }
    return nil
}
```

**2. 本地缓存 + 分布式缓存二级缓存**
```go
type MultiLevelCache struct {
    local *bigcache.BigCache // 本地缓存
    remote *redis.Client     // 远程缓存
}

func (c *MultiLevelCache) Get(ctx context.Context, key string) (string