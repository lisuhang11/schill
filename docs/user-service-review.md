# User Service 全面技术审查报告

> 审查范围：`/workspace/service/user/` 全部代码 + 公共库依赖
> 审查日期：2026-06-03

---

## 一、服务概览

User Service 是 SChill 社区平台的**用户中心微服务**，提供用户注册、登录认证、资料管理、统计查询等功能，并通过 Kafka 消费事件实时更新用户统计。

| 属性 | 值 |
|------|-----|
| 框架 | go-zero v1.10.0 + gRPC |
| 端口 | gRPC :8080 |
| 数据库 | MySQL 8.0（3 张表） |
| 缓存 | Redis（go-zero cache + go-redis/v9） |
| 消息队列 | Kafka（4 个 Consumer Group） |
| 认证方案 | JWT（HS256） + Token Version |
| 密码哈希 | bcrypt（cost=12） |

---

## 二、数据模型

### 2.1 表结构关系

```
user (1) ──── (1) user_profile
  │
  └── (1) user_stat
```

**三表拆分理由**：`user_profile` 存储低频更新的扩展信息，`user_stat` 存储高频更新的计数数据，避免热点更新影响核心用户查询。

### 2.2 user 表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `BIGINT UNSIGNED AUTO_INCREMENT` | 主键 |
| `username` | `VARCHAR(32)` | 用户名，唯一索引 |
| `phone` | `VARCHAR(16)` | 手机号，唯一索引（可空） |
| `email` | `VARCHAR(64)` | 邮箱，唯一索引（可空） |
| `password_hash` | `VARCHAR(255)` | bcrypt 哈希 |
| `avatar` | `VARCHAR(255)` | 头像 URL |
| `status` | `TINYINT` | 1=正常，2=禁言，3=冻结 |
| `is_admin` | `TINYINT` | 0=否，1=是 |
| `last_login_time` | `DATETIME(3)` | 最后登录时间 |
| `deleted_at` | `DATETIME(3)` | 软删除（gorm.DeletedAt） |

### 2.3 user_profile 表

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | `BIGINT UNSIGNED` | 关联用户，唯一索引 |
| `gender` | `TINYINT` | 0=未知，1=男，2=女 |
| `birthday` | `DATE` | 生日 |
| `signature` | `VARCHAR(255)` | 个性签名 |
| `location` | `VARCHAR(64)` | 所在地 |
| `website` | `VARCHAR(255)` | 个人网站 |
| `company` | `VARCHAR(64)` | 公司 |
| `job_title` | `VARCHAR(64)` | 职位 |
| `education` | `VARCHAR(64)` | 教育背景 |

### 2.4 user_stat 表

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | `BIGINT UNSIGNED` | 关联用户，唯一索引 |
| `post_count` | `INT UNSIGNED` | 发帖数 |
| `comment_count` | `INT UNSIGNED` | 评论数 |
| `follower_count` | `INT UNSIGNED` | 粉丝数 |
| `following_count` | `INT UNSIGNED` | 关注数 |
| `like_count` | `INT UNSIGNED` | 获赞总数 |
| `collection_count` | `INT UNSIGNED` | 被收藏总数 |
| `last_active_time` | `BIGINT` | 最后活跃时间（Unix 秒） |

---

## 三、API 接口全览（10 个 RPC）

| RPC 方法 | 请求 | 响应 | 说明 |
|----------|------|------|------|
| `Register` | `username, password` | `userId` | 用户注册 |
| `Login` | `username, password` | `userId, access/refresh token` | 用户登录 |
| `RefreshToken` | `refreshToken` | 新 token 对 | 刷新令牌 |
| `GetUserInfo` | `userId` | `UserInfo + UserProfile + UserStat` | 用户完整信息 |
| `GetUserProfileInfo` | `userId` | `UserProfile` | 用户资料 |
| `GetUserStat` | `userId` | `UserStat` | 用户统计 |
| `BatchGetUserBasicInfo` | `userIds[]` (≤200) | `UserBasicInfo[]` | 批量基础信息 |
| `UpdateUserStatus` | `userId, status` | `success` | 更新用户状态 |
| `UpdateUserProfileInfo` | `userId, UserProfile` | `UserProfile` | 更新个人资料 |
| `UpdateAvatar` | `userId, avatarUrl` | `avatarUrl` | 更新头像 |

---

## 四、核心业务流程

### 4.1 注册流程

```
客户端                  user-rpc                    MySQL
  │                       │                          │
  │──RegisterReq──────────▶                          │
  │                       │                          │
  │              ① validatePassword()                │
  │              ② bcrypt(password, cost=12)         │
  │                       │                          │
  │                       │──BEGIN TRANSACTION──────▶│
  │                       │──INSERT user─────────────▶│
  │                       │  (ON CONFLICT username    │
  │                       │   DO NOTHING)            │
  │                       │  RowsAffected==0?        │
  │                       │  → ErrUsernameExists     │
  │                       │──INSERT user_profile─────▶│
  │                       │──INSERT user_stat────────▶│
  │                       │──COMMIT──────────────────▶│
  │                       │                          │
  │◀──RegisterResp────────│                          │
```

**关键设计点**：
- 使用 `ON CONFLICT ... DO NOTHING` 实现原子化的用户名唯一性检查
- 事务内一次性创建三表记录，保证数据一致性
- 默认头像 URL 硬编码：`http://localhost:9000/user-avatar/user_default_avatar.png`

**用户名唯一性方案选型对比**：

| 方案 | 实现 | 优点 | 缺点 |
|------|------|------|------|
| **当前方案：ON CONFLICT DO NOTHING** | `tx.Clauses(clause.OnConflict{...DoNothing: true}).Create(user)` | 单次 DB 操作，原子性 | 依赖特定 GORM 子句；软删除场景下唯一索引可能包含已删除记录 |
| SELECT + INSERT | 先查是否存在，再插入 | 逻辑直观 | 两次 DB 调用，非原子，有并发竞争 |
| 分布式锁 | `SETNX user:register:{username}` 锁 + 插入 | 跨实例安全 | 引入 Redis 依赖，锁超时处理复杂 |
| 唯一索引 + 捕获错误 | 直接 INSERT，捕获 `Duplicate entry` 错误 | 一次操作，纯 DB 保证 | 需要区分"冲突"和"其他错误" |

**分析**：当前方案是较好的选择。注意 `ON CONFLICT` 依赖唯一索引 `uk_username`，且唯一索引包含 `deleted_at` 字段的话，软删除后同名用户无法重新注册——这是一个**潜在的 bug**，取决于索引是否包含 `deleted_at`。

---

### 4.2 登录流程

```
客户端                  user-rpc                    MySQL          Redis
  │                       │                          │              │
  │──LoginReq─────────────▶                          │              │
  │                       │                          │              │
  │              ① SELECT user WHERE username=?      │              │
  │                       │─────────────────────────▶│              │
  │                       │◀──── user record ────────│              │
  │                       │                          │              │
  │              ② 检查 deleted_at                    │              │
  │              ③ bcrypt.CompareHashAndPassword()   │              │
  │              ④ 检查 status != 1                   │              │
  │                       │                          │              │
  │              ⑤ UPDATE last_login_time            │              │
  │              ⑥ 查询/创建 user_stat               │              │
  │                 (更新 last_active_time)           │              │
  │                       │                          │              │
  │              ⑦ 清除用户缓存                       │──────────────▶│
  │              ⑧ 获取/初始化 tokenVersion           │──────────────▶│
  │              ⑨ 生成 Access Token (JWT)            │              │
  │              ⑩ 生成 Refresh Token (JWT)           │              │
  │                       │                          │              │
  │◀──LoginResp───────────│                          │              │
```

**关键设计点**：
- **密码验证后统一错误信息**：无论用户名不存在、密码错误还是已删除，都返回 `ErrInvalidCredentials`，防止用户枚举攻击
- **last_active_time 更新采用 UPSERT 模式**：先查 `user_stat`，存在则 UPDATE，不存在则 CREATE
- **tokenVersion 机制**：首次登录时用 `user.CreatedAt.Unix()` 作为初始版本号

---

### 4.3 Token 刷新流程

```
客户端                  user-rpc                    MySQL          Redis
  │                       │                          │              │
  │──RefreshTokenReq──────▶                          │              │
  │                       │                          │              │
  │              ① jwt.ParseRefreshToken()           │              │
  │              ② SELECT user WHERE id=?             │              │
  │              ③ 检查 deleted_at + status           │              │
  │              ④ 校验 tokenVersion                  │──────────────▶│
  │                 (Redis GET token_version:{uid})   │              │
  │              ⑤ 生成新 Access Token                │              │
  │              ⑥ 生成新 Refresh Token               │              │
  │                       │                          │              │
  │◀──RefreshTokenResp────│                          │              │
```

**关键设计点**：
- 刷新时返回新的 Access Token **和** Refresh Token（滑动过期策略）
- tokenVersion 校验：若 token 中的版本与 Redis 中不一致，拒绝刷新（实现令牌吊销）
- 版本为 0 时回退到 `user.CreatedAt.Unix()`

---

### 4.4 用户信息查询（GetUserInfo）

采用 **LEFT JOIN 三表一次查询**：

```sql
SELECT user.*, profile.*, stat.*
FROM user
LEFT JOIN user_profile profile ON user.id = profile.user_id
LEFT JOIN user_stat stat ON user.id = stat.user_id
WHERE user.id = ? AND user.deleted_at IS NULL
```

查询结果通过 `go-zero Cache.TakeCtx()` 缓存 1 小时。

**多表查询方案对比**：

| 方案 | 实现 | DB 调用次数 | 优点 | 缺点 |
|------|------|------------|------|------|
| **当前：LEFT JOIN** | 一次 JOIN 查询 | 1 | 单次 DB 往返，缓存粒度统一 | SQL 复杂，JOIN 结果集可能较大 |
| 三次独立查询 | 分别查 user, profile, stat | 3 | 代码简单，每表独立缓存 | 三次 DB 往返，需要合并逻辑 |
| 先查 user，再并发查其余 | goroutine 并行 | 2（并发） | 较快，可独立缓存 | 需要 errgroup 管理并发 |
| 查询后异步补全 | 先返回 user，异步加载 profile+stat | 1 + 异步 | 首屏最快 | 需要前端配合二次加载，体验复杂 |

**分析**：当前 JOIN 方案适合小数据量场景。高并发下，若三表有独立缓存需求，推荐方案 3（并发独立查询）。

---

### 4.5 批量查询（BatchGetUserBasicInfo）

```
① 去重 userIDs (uniqueUint64s)
② 限制最大 200 个
③ 逐个检查 Redis 缓存（cache-aside）
   ├─ 命中 → 加入结果 map
   └─ 未命中 → 加入 missedIDs
④ 对 missedIDs 批量查询 DB: WHERE id IN (...)
⑤ 写回缓存（包括空值缓存防穿透）
⑥ 按原始顺序返回结果
```

**批量缓存方案对比**：

| 方案 | 缓存粒度 | Redis 调用 | 优点 | 缺点 |
|------|---------|-----------|------|------|
| **当前：逐个 GetCtx** | per-user | N 次 | 最大缓存复用 | N 次网络往返，延迟累加 |
| Pipeline GET | per-user | 1 次 | 单次往返 | 需要 Pipeline 支持，go-zero Cache 不支持 |
| MGET 批量读取 | per-user | 1 次 | 最快 | 缓存结构需统一为 string，不兼容复杂类型 |
| Hash 聚合存储 | per-batch | 1 次 HGETALL | 单次读写 | 缓存粒度粗，任意用户更新导致全量失效 |

**分析**：当前逐个 GET 的方案在 200 个 ID 时会有 200 次 Redis 往返（约 20-40ms），可优化为 Pipeline 或 MGET。但 go-zero Cache 封装不支持 Pipeline，改造需要绕过 `Cache` 接口直接操作底层 Redis。

---

### 4.6 用户资料更新（UpdateUserProfileInfo）

采用 **部分更新（map-based UPDATE）**：

```go
updates := make(map[string]interface{})
if in.UserProfile.Gender != nil { updates["gender"] = *in.UserProfile.Gender }
// ... 逐字段检查
if len(updates) > 0 {
    db.Model(&profile).Updates(updates)
    invalidateUserCaches(...)
}
```

**部分更新方案对比**：

| 方案 | 实现 | SQL | 优点 | 缺点 |
|------|------|-----|------|------|
| **当前：map[string]interface{}** | GORM `Updates(map)` | `UPDATE SET col1=v1, col2=v2` | 不更新零值字段，不触发 Hook | map 类型不安全 |
| struct 更新 | GORM `Updates(struct)` | 同上 | 类型安全 | 零值字段也会更新，会触发 Hook |
| `Select().Updates()` | 指定字段名列表 | `UPDATE SET col1=v1` | 精确控制 | 需要手动维护字段列表 |
| 全量覆盖 | `Save()` | `UPDATE SET all_columns` | 最简单 | 并发写可能覆盖，非请求字段会被清空 |

**分析**：当前方案合理。`map[string]interface{}` 不触发 GORM Hook，`Updates(struct)` 会触发——此处选择正确。但 `map[string]interface{}` 丢失类型检查，建议用泛型或字段函数封装。

---

## 五、认证与安全

### 5.1 JWT 方案

| 属性 | 配置 |
|------|------|
| 签名算法 | HS256（对称加密） |
| Access Token TTL | 86400s（24 小时） |
| Refresh Token TTL | 604800s（7 天） |
| Token 载体 | `userId` + `tokenVersion` |

**JWT 签名算法对比**：

| 算法 | 类型 | 性能 | 安全性 | 适用场景 |
|------|------|------|--------|---------|
| **HS256（当前）** | 对称（HMAC-SHA256） | 快 | 依赖 secret 保密 | 单体/小规模微服务 |
| RS256 | 非对称（RSA） | 较慢 | 公私钥分离 | 多服务验证，auth service 签发 |
| ES256 | 非对称（ECDSA） | 快 | 强 | 现代推荐方案 |
| EdDSA | 非对称（Ed25519） | 最快 | 最强 | 最新方案 |

**分析**：HS256 在单服务场景下足够，但 secret 泄露后所有 token 可伪造。若扩展到多服务需要独立验证 token 的场景，应迁移到 RS256/ES256。

### 5.2 Token Version 吊销机制

```
┌──────────────────────────────────────┐
│  登录时：                              │
│    1. Redis GET token_version:{uid}   │
│    2. 若不存在 → version = createdAt   │
│    3. 签发 JWT(uid, version)          │
│                                      │
│  刷新时：                              │
│    1. 解析 JWT → 提取 version          │
│    2. Redis GET token_version:{uid}   │
│    3. version 不一致 → 拒绝刷新         │
│                                      │
│  吊销时（改密/封号）：                   │
│    1. Redis INCR token_version:{uid}  │
│    2. 旧 token 的 version 不再匹配      │
└──────────────────────────────────────┘
```

**令牌吊销方案对比**：

| 方案 | 实现 | 存储开销 | 实时性 | 优点 | 缺点 |
|------|------|---------|--------|------|------|
| **Token Version（当前）** | Redis 存储版本号，JWT 携带 | O(1) per user | 实时 | 简单高效 | 依赖 Redis，version 无过期 |
| 黑名单 | Redis SET 存储已吊销 token | O(N) tokens | 实时 | 精确控制 | 内存随 token 数量增长 |
| 短 TTL + 无吊销 | Access Token 5min | 0 | 最多 5min 延迟 | 零存储 | 不能即时吊销 |
| Refresh Token 轮换 | 每次刷新换新 token 对 | O(1) | 实时 | 无额外存储 | 并发刷新可能出问题 |

**分析**：当前方案较好。Token Version 依赖 Redis——若 Redis 数据丢失，version 回退到 0，已吊销的 token 会重新生效。建议将 version 也持久化到 DB 中作为兜底。

### 5.3 密码安全

```go
// 注册：bcrypt.GenerateFromPassword(password, cost=12)
// 登录：bcrypt.CompareHashAndPassword(hash, password)
```

**密码哈希方案对比**：

| 方案 | 成本参数 | 抗 GPU | 内存硬 | 说明 |
|------|---------|--------|--------|------|
| **bcrypt（当前）** | cost=12 (~300ms) | 中等 | 否 | 成熟稳定，Go 标准库 |
| scrypt | N=32768, r=8, p=1 | 高 | 是 | 内存硬，抗 ASIC |
| argon2id | time=1, mem=64MB | 最高 | 是 | 2015 年密码哈希竞赛冠军 |
| PBKDF2-SHA256 | 迭代 100k+ | 低 | 否 | NIST 标准，但 GPU 易破解 |
| SHA-256（无盐） | N/A | 极低 | 否 | ❌ 不安全 |

**分析**：bcrypt cost=12 在 2026 年是合理的（约 300ms 哈希时间）。如需更高安全性，可升级到 argon2id。当前代码有一个设计问题：**`PasswordVerify` 返回 `bool` 而非 `error`**，无法区分"密码不匹配"和"哈希格式错误"，不过在实际使用中影响不大。

### 5.4 密码强度验证

```go
func validatePassword(password string) error {
    // 最少 8 位
    // 至少包含大写字母、小写字母、数字各一个
}
```

**不足**：未检查特殊字符。建议补充特殊字符要求或改用 zxcvbn 等密码强度库。

---

## 六、缓存策略

### 6.1 缓存架构

```
                    ┌─────────────────┐
                    │   go-zero Cache  │
                    │  (Singleflight)  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │     Redis        │
                    └─────────────────┘
```

- **缓存库**：go-zero `cache.Cache`，内置 singleflight 防击穿
- **缓存策略**：Cache-Aside（先查缓存，未命中查 DB 并回写）
- **空值缓存**：TTL 60s（`CacheNullExpire`），防穿透
- **正常缓存 TTL**：3600s（`UserExpire`）

### 6.2 缓存 Key 设计

| Key Pattern | 内容 | TTL |
|-------------|------|-----|
| `schill:user:info:{uid}` | `GetUserInfoResp`（完整信息） | 3600s |
| `schill:user:profile:{uid}` | `GetUserProfileInfoResp` | 3600s |
| `schill:user:stat:{uid}` | `GetUserStatResp` | 3600s |
| `schill:user:info:{uid}:basic` | `UserBasicInfo` | 3600s |
| `schill:user:token_version:{uid}` | token 版本号 | 无过期 |

### 6.3 缓存失效策略

```go
func invalidateUserCaches(ctx, svcCtx, userID) {
    Del("schill:user:info:{uid}")
    Del("schill:user:profile:{uid}")
    Del("schill:user:stat:{uid}")
    Del("schill:user:info:{uid}:basic")
}
```

**失效策略对比**：

| 策略 | 实现 | 一致性 | 复杂度 |
|------|------|--------|--------|
| **当前：直接删除（Cache-Aside）** | 写 DB 后删缓存 | 最终一致 | 低 |
| 先删缓存再写 DB | 删缓存→写 DB | 有并发写覆盖风险 | 低 |
| 双删 | 删缓存→写 DB→延迟再删 | 较好 | 中 |
| 订阅 Binlog 异步更新 | Canal→消费→刷新缓存 | 准实时 | 高 |

**分析**：当前"先写 DB 再删缓存"是标准 Cache-Aside 模式。并发场景下存在短暂不一致窗口（写 DB 完成到删缓存之间），但概率低且影响小。`DelCtx` 失败时仅打日志，**未重试**——可能导致缓存脏数据长期存在。

---

## 七、Kafka 事件消费

### 7.1 消费者架构

```
Kafka Topics                    Consumer Groups
─────────────                   ────────────────────────
post-created    ────▶  user-post-count-consumer-group-created
post-deleted    ────▶  user-post-count-consumer-group-deleted
user-followed   ────▶  user-post-count-consumer-group-followed
user-unfollowed ────▶  user-post-count-consumer-group-unfollowed
```

每个 consumer 都配置了 `IdempotencyStore`（Redis SETNX）。

### 7.2 事件处理器

| 事件 | 处理器 | 操作 | 幂等性 |
|------|--------|------|--------|
| `post-created` | `PostCreatedHandler` | `post_count + 1`（无记录则创建） | 天然幂等（创建+1） |
| `post-deleted` | `PostDeletedHandler` | `post_count - 1`（下限 0） | 应用层保护（count>0 才减） |
| `user-followed` | `UserFollowedHandler` | `following_count + 1` + `follower_count + 1` | 无记录时创建 |
| `user-unfollowed` | `UserUnfollowedHandler` | `GREATEST(count - 1, 0)` | SQL 层面保护 |

### 7.3 幂等性实现

```go
// event_helper.go
func skipUserEvent(svcCtx, group, envelope) bool {
    key := "schill:mq:consume:{group}:{eventID}"
    ok, _ := svcCtx.Redis.SetnxExCtx(ctx, key, "1", 24h)
    return !ok  // false = 新事件，true = 重复
}
```

**注意**：user service 使用的是自己实现的 `skipUserEvent`（在 handler 中调用），而非 `common/kafka/consumer.go` 中的 `IdempotencyStore`。两者逻辑一致但代码重复。

### 7.4 Kafka 事件一致性对比

| 方案 | 原理 | 一致性 | 复杂度 | 延迟 |
|------|------|--------|--------|------|
| **当前：异步 Kafka** | 发帖服务写 DB → 发 Kafka → user service 消费 | 最终一致 | 中 | ms 级 |
| Transactional Outbox | DB 事务内写 outbox 表 → 轮询发 Kafka | 最终一致 | 高 | ms~s 级 |
| 同步 RPC | 发帖服务直接调 user service 更新计数 | 强一致 | 低 | 实时 |
| 不维护冗余计数 | 每次查询时 COUNT | 强一致 | 低 | 查询慢 |

**分析**：当前方案在"发帖成功但 Kafka 发送失败"时，计数会不一致。对于用户发帖数这种非关键指标，最终一致性可接受。若需强一致，应使用 Transactional Outbox。

---

## 八、ServiceContext 初始化

```go
func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 初始化 MySQL (GORM + 连接池)
    db := commondb.OpenMySQL(c.Mysql.DataSource, poolConfig)
    
    // 2. AutoMigrate (非生产环境)
    commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate,
        &model.User{}, &model.UserProfile{}, &model.UserStat{})
    
    // 3. 初始化 go-zero Cache (Redis + Singleflight)
    cacheNode := cache.New(c.Cache, syncx.NewSingleFlight(), ...)
    
    // 4. 初始化通用 Redis 客户端
    redisClient := commonredis.NewClient(...)
    
    return &ServiceContext{DB, Cache, Redis, RedisClient}
}
```

**连接池配置**：
- MaxOpenConns: 100
- MaxIdleConns: 20
- ConnMaxLifetime: 600s
- ConnMaxIdleTime: 300s

---

## 九、存在的问题

### 9.1 安全问题

#### 🔴 P0-1: UpdateUserProfileInfo 信任客户端 userId

```go
// updateuserprofileinfologic.go:35-38
userId := in.UserId          // ← 来自客户端！
if userId == 0 {
    userId = in.UserProfile.UserId  // ← 也是客户端！
}
```

**风险**：任何已认证用户可以修改任意用户的资料。虽然 gRPC 服务通常在网关层做鉴权，但 userId 应从 JWT context 中提取而非信任客户端传入。

**修复**：在网关层将 JWT 中的 userId 注入 gRPC metadata，服务端从 context 提取。

#### 🔴 P0-2: UpdateAvatar 同样信任客户端 userId

```go
// updateavatarlogic.go:31
if in.UserId == 0 { ... }  // ← 来自客户端
```

同上。

#### 🟠 P1-1: 密码验证函数返回 bool 而非 error

```go
// crypt.go
func PasswordVerify(hashedPassword, plainPassword string) bool
```

无法区分"密码错误"和"哈希格式损坏"两种失败场景。

#### 🟠 P1-2: 默认头像 URL 硬编码

```go
// registerlogic.go:51
Avatar: "http://localhost:9000/user-avatar/user_default_avatar.png"
```

部署到不同环境时 URL 错误。

### 9.2 可靠性问题

#### 🟠 P1-3: 缓存删除无重试机制

```go
// cache_helper.go:39-41
for _, key := range keysToDelete {
    if err := svcCtx.Cache.DelCtx(ctx, key); err != nil {
        logx.Errorf(...)  // 仅打日志
    }
}
```

Redis 临时不可用时缓存删除失败，导致脏数据残留长达 1 小时。

**修复**：失败时写入"待清理"队列，异步重试。

#### 🟠 P1-4: tokenVersion 仅存 Redis，无 DB 兜底

Redis 数据丢失 → tokenVersion 归零 → 已吊销的 token 重新生效。

**修复**：将 version 同时持久化到 user 表。

#### 🟡 P2-1: login 中的 user_stat 处理逻辑冗余

```go
// loginlogic.go:64-83
// 先查 userStat → 不存在则创建 → 存在则更新
// 但注册时已创建了 user_stat 记录
// 仅防御"注册后手动删除 stat 记录"的边缘情况
```

### 9.3 性能问题

#### 🟡 P2-2: BatchGetUserBasicInfo 逐次 GET 缓存

200 个 ID → 200 次 Redis 往返 → ~20-40ms 纯网络延迟。

**优化**：使用 Redis Pipeline 或 MGET。

#### 🟡 P2-3: GetUserInfo LEFT JOIN 三表

单次查询可接受，但大字段（如 `signature` 可能存长文本）会增加网络传输。

**优化**：考虑按需加载 profile/stat（前端指定需要的模块）。

### 9.4 代码质量问题

#### 🟡 P2-4: proto 状态值与 DB 定义不一致

```protobuf
// user.proto:17
int32 status = 7;  // 注释：0:正常 1:封禁 2:删除
```

```sql
-- db.sql:9
`status` TINYINT ... COMMENT '状态：1正常，2禁言，3冻结'
```

proto 注释和 DB 定义不一致，且实际代码中 `UpdateUserStatus` 接受 `1-3`。

#### 🟡 P2-5: 幂等性逻辑重复实现

`common/kafka/consumer.go` 有 `IdempotencyStore`，`mqs/event_helper.go` 有 `skipUserEvent`。两者功能相同但代码独立。

#### 🟡 P2-6: user_stat 中的 last_active_time 类型混乱

- DB 定义：`DATETIME(3)`（MySQL 时间类型）
- GORM model：`int64`（Unix 时间戳）
- `GetUserInfo`：使用 `UNIX_TIMESTAMP(last_active_time)` 转换
- `Login`：直接赋值 `time.Now().Unix()`

说明 DB 实际存储的是 Unix 时间戳数值而非 DATETIME，与 DDL 定义不一致。

#### 🟡 P2-7: UpdateAvatar 使用 Save 而非 Update

```go
// updateavatarlogic.go:53
l.svcCtx.DB.WithContext(l.ctx).Save(&user)
```

`Save` 会更新所有字段，而 `Update("avatar", url)` 只更新 avatar。在高并发下 `Save` 可能覆盖其他字段的并发更新。

### 9.5 架构问题

#### 🟡 P2-8: 用户统计通过 Kafka 异步更新

发帖数/粉丝数等统计存在短暂不一致（最终一致）。对于粉丝数这种展示型数据可接受，但若用于业务逻辑判断（如"粉丝超过 100 才能发帖"），则需要考虑一致性窗口。

#### 🟡 P2-9: 缺少操作审计

密码修改、状态变更、资料修改等操作无审计日志，不利于安全追溯。

---

## 十、技术方案选型总结

### 10.1 数据库方案

| 决策点 | 当前选择 | 替代方案 | 建议 |
|--------|---------|---------|------|
| ID 生成 | MySQL AUTO_INCREMENT | Snowflake / UUIDv7 | 单机可接受，分布式建议 Snowflake |
| 软删除 | `gorm.DeletedAt` | `*time.Time` 手动过滤 | 当前正确 |
| 表拆分 | 3 表（user/profile/stat） | 单表 / 更多表 | 当前合理 |
| 连接池 | MaxOpen=100, MaxIdle=20 | — | 根据 QPS 调优 |

### 10.2 缓存方案

| 决策点 | 当前选择 | 替代方案 | 建议 |
|--------|---------|---------|------|
| 缓存模式 | Cache-Aside | Read-Through / Write-Through | 当前合理 |
| 防击穿 | go-zero Singleflight | 分布式锁 | 当前合理 |
| 防穿透 | 空值缓存 60s | 布隆过滤器 | 当前够用，大数据量建议布隆 |
| 防雪崩 | 固定 TTL 3600s | 随机 TTL Jitter | **建议添加 Jitter** |
| 缓存粒度 | per-entity | 聚合缓存 | 当前合理 |

### 10.3 认证方案

| 决策点 | 当前选择 | 替代方案 | 建议 |
|--------|---------|---------|------|
| 签名算法 | HS256 | RS256 / ES256 / EdDSA | 多服务建议 ES256 |
| Access Token TTL | 24h | 15min / 1h | **建议缩短到 15-60min** |
| 吊销机制 | Token Version | 黑名单 / 短 TTL | 当前合理，建议加 DB 兜底 |
| 密码哈希 | bcrypt(cost=12) | argon2id / scrypt | 当前合理 |

### 10.4 事件驱动方案

| 决策点 | 当前选择 | 替代方案 | 建议 |
|--------|---------|---------|------|
| 消息队列 | Kafka | RabbitMQ / Redis Stream | 当前合理 |
| 一致性模型 | 最终一致（异步） | 强一致（同步 RPC / Outbox） | 统计类数据最终一致可接受 |
| 幂等性 | Redis SETNX 24h | DB 唯一约束 | 当前合理，建议加 DB 兜底 |
| 消费者分组 | 4 个独立 CG | 单 CG 多 handler | 当前合理，隔离性好 |

---

## 十一、改进建议优先级

### P0（紧急）

1. **修复 UpdateUserProfileInfo 和 UpdateAvatar 的 userId 信任问题**：从 JWT context 提取 userId
2. **统一 proto 状态值与 DB 定义**：避免文档与代码不一致

### P1（重要）

3. **缓存删除添加重试机制**：失败时写入延迟队列
4. **tokenVersion 持久化到 DB**：Redis 丢失时的兜底
5. **Access Token TTL 从 24h 缩短到 1h**：减少 token 泄露窗口
6. **默认头像 URL 配置化**：从配置文件读取

### P2（改进）

7. **BatchGetUserBasicInfo 使用 Pipeline**：减少 Redis 往返
8. **UpdateAvatar 改用 Update 单字段**：避免并发覆盖
9. **统一幂等性实现**：复用 common/kafka 的 IdempotencyStore
10. **添加缓存 TTL Jitter**：防雪崩

---

## 十二、服务依赖图

```
                    ┌──────────────┐
                    │   Gateway     │
                    │  (HTTP API)   │
                    └──────┬───────┘
                           │ gRPC
                    ┌──────▼───────┐
                    │  user-rpc     │
                    │  (:8080)      │
                    └──┬───┬───┬──┘
                       │   │   │
              ┌────────▼┐  │   └──────────────┐
              │  MySQL  │  │                  │
              │ (3表)   │  │           ┌──────▼──────┐
              └─────────┘  │           │    Kafka     │
                           │           │ (4 topics)   │
                    ┌──────▼──────┐   └──────────────┘
                    │    Redis    │
                    │ (缓存+版本)  │
                    └─────────────┘
```

**被依赖方**（调用 user-rpc 的服务）：
- `gateway`：所有用户相关 API 的入口
- `feed-rpc`：批量获取帖子作者基础信息
- `content-rpc`：获取帖子作者信息
- `comment-rpc`：获取评论用户信息
- `relation-rpc`：获取关注/粉丝用户信息

**依赖方**（user-rpc 调用的外部服务）：
- 无（user-rpc 不主动调用其他微服务，仅消费 Kafka 事件）
