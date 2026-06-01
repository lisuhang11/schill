# SChill

SChill 是一个基于 Go 微服务和 Next.js 前端构建的内容社区系统。项目围绕用户、内容发布、评论、关注关系、点赞收藏、信息流和全文搜索等社区核心场景拆分服务，后端采用 go-zero 体系组织 RPC/API 服务，前端采用 Next.js App Router 提供 Web 交互界面。

当前仓库包含完整的后端服务代码、前端代码、数据库建表脚本、Elasticsearch 索引定义、Docker Compose 基础设施编排、Docker Desktop Kubernetes 本地部署清单，以及 OpenSpec 需求/设计文档。

## 项目定位

SChill 面向内容社区、校园论坛、兴趣社区或轻量社交产品一类场景，核心能力包括：

- 用户注册、登录、令牌刷新、用户资料和统计信息维护
- 动态/帖子发布、编辑、删除、详情、列表、话题和标签
- 评论、回复、评论投票、评论计数异步更新
- 用户关注、取消关注、关注关系和互关状态查询
- 帖子点赞、取消点赞、收藏、取消收藏、分享、收藏列表
- 信息流聚合，合并内容、作者、关注关系和互动状态
- 基于 Elasticsearch 的帖子、用户、话题搜索
- 基于 Kafka 的事件驱动计数更新和最终一致性处理
- 基于 Canal 的 MySQL Binlog 到 Kafka/Elasticsearch 同步链路
- Docker Compose 与 Docker Desktop Kubernetes 本地部署支持

## 技术栈

### 后端

- Go `1.25.0`
- go-zero `1.10.0`
- gRPC / Protocol Buffers
- GORM / MySQL Driver
- Redis
- Kafka / Sarama
- Elasticsearch Go Client v8
- MinIO SDK
- JWT

### 前端

- Next.js `15`
- React `19`
- TypeScript `5`
- Tailwind CSS `3`
- React Hook Form
- Zod
- lucide-react

### 基础设施

- MySQL 8.0
- Redis 7.0
- Etcd 3.5
- Kafka 3.9.1 KRaft
- Canal Server 1.1.7
- Elasticsearch 8.6.1
- Kibana 8.6.1
- MinIO
- Docker / Docker Compose
- Docker Desktop Kubernetes / Kustomize

## 系统架构

```text
                         +----------------+
                         |  Next.js Web   |
                         +-------+--------+
                                 |
                                 | HTTP /api/*
                                 v
                         +-------+--------+
                         |    Gateway     |
                         |    :8086       |
                         +---+---+---+----+
                             |   |   |
             +---------------+   |   +----------------+
             |                   |                    |
             v                   v                    v
       +-----+------+      +-----+------+       +-----+------+
       | User RPC   |      | Content RPC|       | Comment RPC|
       | :8080      |      | :8082      |       | :8083      |
       +-----+------+      +-----+------+       +-----+------+
             |                   |                    |
             v                   v                    v
          MySQL               MySQL                 MySQL

       +------------+      +-------------+      +-------------+
       | Relation   |      | Interaction |      | Feed RPC    |
       | RPC :8081  |      | RPC :8084   |      | :8087       |
       +-----+------+      +------+------+      +------+------+
             |                    |                    |
             v                    v                    |
          MySQL              MySQL / Redis             |
                                  |                    |
                                  v                    |
                                Kafka <----------------+

                         +----------------+
                         | Search API     |
                         | :8895          |
                         +-------+--------+
                                 |
                                 v
                         Elasticsearch

       MySQL Binlog -> Canal Server -> Kafka -> Canal Sync -> Elasticsearch
```

Etcd 用于 go-zero RPC 服务发现。Gateway 作为统一 HTTP 入口，直接调用各 RPC 服务，并将搜索请求代理到 Search API。

## 服务说明

| 服务 | 路径 | 默认端口 | 类型 | 职责 |
| --- | --- | ---: | --- | --- |
| gateway | `service/gateway` | `8086` | HTTP API | 统一网关、鉴权上下文、调用 RPC、搜索代理 |
| user-rpc | `service/user/rpc` | `8080` | gRPC | 注册登录、令牌、用户资料、用户统计 |
| relation-rpc | `service/relation/rpc` | `8081` | gRPC | 关注、取关、粉丝/关注列表、互关状态 |
| content-rpc | `service/content/rpc` | `8082` | gRPC | 帖子、帖子内容、话题、标签、置顶/精华、浏览量 |
| comment-rpc | `service/comment/rpc` | `8083` | gRPC | 评论、回复、删除、点赞/点踩 |
| interaction-rpc | `service/interaction/rpc` | `8084` | gRPC | 帖子点赞、收藏、分享和互动状态 |
| search-api | `service/search/api` | `8895` | HTTP API | 帖子、用户、话题搜索 |
| feed-rpc | `service/feed/rpc` | `8087` | gRPC | 信息流聚合，组装内容、作者和浏览者状态 |
| canal-sync | `service/canal` | `8088` | 后台服务 | 消费 Canal/Kafka 消息并同步 Elasticsearch |

## 目录结构

```text
.
├── common/                         # 后端公共库
│   ├── authctx/                    # 鉴权上下文和中间件
│   ├── cacheprotect/               # 缓存保护工具
│   ├── cacheutil/                  # 缓存辅助工具
│   ├── cryptx/                     # 密码/加密相关工具
│   ├── db/                         # MySQL 连接和迁移辅助
│   ├── error/                      # 统一错误码和错误响应
│   ├── es/                         # Elasticsearch 客户端封装
│   ├── jwt/                        # JWT 生成和解析
│   ├── kafka/                      # Kafka 生产/消费工具
│   ├── minio/                      # MinIO 客户端封装
│   ├── mq/                         # 业务消息结构和生产者
│   ├── observability/              # HTTP 可观测性辅助
│   ├── redis/                      # Redis 客户端、Key、Lua 脚本
│   └── uuid/                       # ID/UUID 工具
├── service/                        # 后端业务服务
│   ├── gateway/                    # HTTP 网关
│   ├── user/rpc/                   # 用户 RPC 服务
│   ├── relation/rpc/               # 关注关系 RPC 服务
│   ├── content/rpc/                # 内容 RPC 服务
│   ├── comment/rpc/                # 评论 RPC 服务
│   ├── interaction/rpc/            # 互动 RPC 服务
│   ├── feed/rpc/                   # 信息流 RPC 服务
│   ├── search/api/                 # 搜索 API 服务
│   └── canal/                      # Canal 到 ES 同步服务
├── web/                            # Next.js 前端应用
├── deploy/                         # Docker Compose 和 Kubernetes 部署
├── docs/                           # Swagger、原型、面试/部署文档
├── es_index/                       # Elasticsearch 索引 Mapping
├── openspec/                       # OpenSpec 需求、设计和任务文档
├── scripts/                        # 辅助脚本
├── db.sql                          # MySQL 建表脚本
├── go.mod                          # Go Module
└── go.sum
```

## 数据模型概览

数据库脚本位于 `db.sql`，当前定义的核心表包括：

| 表 | 说明 |
| --- | --- |
| `user` | 用户账号、密码哈希、状态、管理员标记、登录信息 |
| `user_profile` | 用户扩展资料，如性别、生日、签名、所在地、网站、公司等 |
| `user_stat` | 用户统计，如发帖数、评论数、粉丝数、关注数、获赞数、收藏数 |
| `topic` | 话题及引用次数 |
| `post` | 帖子主表，包含标题、封面、统计、可见性、置顶、精华、锁定等 |
| `post_content` | 帖子正文内容，多段内容按类型和排序组织 |
| `post_topic` | 帖子和话题关联 |
| `post_star` | 帖子点赞记录 |
| `post_collection` | 帖子收藏记录 |
| `comment` | 评论主表，支持根评论和回复 |
| `comment_content` | 评论正文 |
| `comment_vote` | 评论点赞/点踩记录 |
| `comment_stat` | 评论统计 |
| `following` | 用户关注关系 |

Elasticsearch Mapping 位于 `es_index/`：

- `post_index.json`
- `user_index.json`
- `topic_index.json`

Docker Desktop Kubernetes 部署会通过 `job-es-init.yaml` 初始化这些索引。

## 主要业务链路

### 用户认证

1. 前端调用 Gateway 的注册/登录接口。
2. Gateway 调用 `user-rpc`。
3. `user-rpc` 校验账号密码，生成 Access Token 和 Refresh Token。
4. 前端将 Access Token 写入本地存储，并在后续请求中携带 `Authorization: Bearer <token>`。
5. Gateway 解析令牌并把当前用户 ID 注入业务请求。

### 内容发布

1. 前端通过 Gateway 创建帖子。
2. Gateway 调用 `content-rpc`。
3. `content-rpc` 写入 `post`、`post_content`、`topic`、`post_topic` 等表。
4. 创建/删除帖子时向 Kafka 发送事件。
5. `user-rpc` 消费帖子事件更新用户发帖统计。
6. Canal 链路将 MySQL 变化同步到 Elasticsearch，用于搜索。

### 评论与计数

1. 前端通过 Gateway 创建或删除评论。
2. Gateway 调用 `comment-rpc`。
3. `comment-rpc` 写入评论相关表，并通过 Kafka 发送评论事件。
4. `content-rpc` 消费评论事件更新帖子评论数和最后回复时间。

### 互动

1. 用户点赞、收藏、分享帖子。
2. Gateway 调用 `interaction-rpc`。
3. `interaction-rpc` 写入互动表或 Redis 状态，并发布 Kafka 事件。
4. `content-rpc` 消费互动事件更新帖子统计。
5. `feed-rpc` 查询信息流时会聚合当前用户的点赞、收藏、关注状态。

### 搜索同步

```text
MySQL -> Binlog -> Canal Server -> Kafka canal_topic -> canal-sync -> Elasticsearch
```

Search API 直接查询 Elasticsearch。Gateway 中 `/api/search/*` 路由会代理到 Search API。

## HTTP API 概览

Gateway 默认监听 `http://localhost:8086`。

### 健康检查

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 网关健康检查 |

### 认证与用户

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/refresh` | 刷新令牌 |
| GET | `/api/users/:id` | 获取用户完整信息 |
| PUT | `/api/users/me/profile` | 更新当前用户资料 |
| PUT | `/api/users/me/avatar` | 更新当前用户头像 |

### 内容与话题

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/feed` | 获取信息流 |
| GET | `/api/posts` | 获取帖子列表 |
| POST | `/api/posts` | 创建帖子 |
| GET | `/api/posts/:id` | 获取帖子详情 |
| PUT | `/api/posts/:id` | 更新帖子 |
| DELETE | `/api/posts/:id` | 删除帖子 |
| GET | `/api/topics` | 获取话题列表 |

### 评论

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/posts/:id/comments` | 获取帖子评论列表 |
| POST | `/api/comments` | 创建评论 |
| DELETE | `/api/comments/:id` | 删除评论 |
| POST | `/api/comments/:id/vote` | 评论点赞/点踩 |

### 关系与互动

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/users/:id/follow` | 关注用户 |
| DELETE | `/api/users/:id/follow` | 取消关注 |
| POST | `/api/posts/:id/star` | 点赞帖子 |
| DELETE | `/api/posts/:id/star` | 取消点赞 |
| POST | `/api/posts/:id/collect` | 收藏帖子 |
| DELETE | `/api/posts/:id/collect` | 取消收藏 |
| POST | `/api/posts/:id/share` | 分享帖子 |
| GET | `/api/users/me/collections` | 当前用户收藏列表 |

### 搜索

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/search/post` | 搜索帖子 |
| GET | `/api/search/user` | 搜索用户 |
| GET | `/api/search/topic` | 搜索话题 |

Search API 自身默认监听 `http://localhost:8895`，内部路由为 `/search/post`、`/search/user`、`/search/topic`。

## 环境准备

建议本地准备：

- Go 1.25 或与 `go.mod` 兼容的 Go 版本
- Node.js 22 或兼容版本
- npm
- Docker Desktop
- Docker Compose
- kubectl，可选
- goctl，可选，用于重新生成 go-zero 代码
- protoc，可选，用于重新生成 protobuf 代码

## 配置说明

后端服务配置文件位于各服务 `etc/` 目录，例如：

- `service/gateway/etc/gateway.yaml`
- `service/user/rpc/etc/user-rpc.yaml`
- `service/content/rpc/etc/content-rpc.yaml`
- `service/comment/rpc/etc/comment-rpc.yaml`
- `service/relation/rpc/etc/relation-rpc.yaml`
- `service/interaction/rpc/etc/interaction-rpc.yaml`
- `service/feed/rpc/etc/feed-rpc.yaml`
- `service/search/api/etc/search-api.yaml`
- `service/canal/etc/canal.yaml`

本地开发配置默认使用：

- MySQL: `127.0.0.1:3306`
- Redis: `127.0.0.1:6379`
- Etcd: `127.0.0.1:2379`
- Kafka: `127.0.0.1:9092`
- Elasticsearch: `http://127.0.0.1:9200`
- Gateway: `http://127.0.0.1:8086`
- Search API: `http://127.0.0.1:8895`

生产或容器网络中应使用 `deploy/k8s/docker-desktop/config/` 下的配置模板，里面的依赖地址会指向容器服务名，例如 `etcd:2379`、`mysql:3306`、`redis:6379`、`kafka:9092`、`elasticsearch:9200`。

前端 API 地址通过环境变量控制：

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8086
```

未设置时，前端默认请求 `http://localhost:8086`。

## 本地启动

### 1. 启动基础设施

如果只需要基础设施，可以使用：

```powershell
docker compose -f deploy/docker-compose.yml up -d
```

如果希望同时通过 Compose 构建并启动后端业务服务，可以使用：

```powershell
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

常用访问地址：

- Gateway: `http://localhost:8086`
- Search API: `http://localhost:8895`
- Elasticsearch: `http://localhost:9200`
- Kibana: `http://localhost:5601`
- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`

### 2. 初始化数据库

`deploy/docker-compose.prod.yml` 会将 `db.sql` 挂载到 MySQL 初始化目录。对于已经存在数据卷的 MySQL，初始化脚本不会自动重复执行，可以手动导入：

```powershell
Get-Content db.sql | docker exec -i mysql mysql -uroot -p123456 schill
```

如果你的 MySQL 密码或数据库名不同，请同步调整命令和服务配置。

### 3. 初始化 Elasticsearch 索引

可以手动执行：

```powershell
curl -X PUT "http://localhost:9200/user" -H "Content-Type: application/json" --data-binary "@es_index/user_index.json"
curl -X PUT "http://localhost:9200/post" -H "Content-Type: application/json" --data-binary "@es_index/post_index.json"
curl -X PUT "http://localhost:9200/topic" -H "Content-Type: application/json" --data-binary "@es_index/topic_index.json"
```

Kubernetes 部署方式中，`job-es-init.yaml` 会自动创建索引。

### 4. 以源码方式启动后端服务

基础设施启动后，可以在不同终端中启动各服务：

```powershell
go run service/user/rpc/user.go -f service/user/rpc/etc/user-rpc.yaml
go run service/relation/rpc/relation.go -f service/relation/rpc/etc/relation-rpc.yaml
go run service/content/rpc/content.go -f service/content/rpc/etc/content-rpc.yaml
go run service/comment/rpc/comment.go -f service/comment/rpc/etc/comment-rpc.yaml
go run service/interaction/rpc/interaction.go -f service/interaction/rpc/etc/interaction-rpc.yaml
go run service/feed/rpc/feed.go -f service/feed/rpc/etc/feed-rpc.yaml
go run service/search/api/search.go -f service/search/api/etc/search-api.yaml
go run service/canal/canal.go -f service/canal/etc/canal.yaml
go run service/gateway/gateway.go -f service/gateway/etc/gateway.yaml
```

建议启动顺序：

1. MySQL、Redis、Etcd、Kafka、Elasticsearch、Canal、MinIO
2. `user-rpc`
3. `relation-rpc`
4. `content-rpc`
5. `comment-rpc`
6. `interaction-rpc`
7. `feed-rpc`
8. `search-api`
9. `canal-sync`
10. `gateway`

### 5. 启动前端

```powershell
cd web
npm install
npm run dev
```

Next.js 默认监听 `http://localhost:3000`。

前端常用脚本：

```powershell
npm run dev
npm run build
npm run start
npm run lint
npm run typecheck
```

## Docker Desktop Kubernetes 部署

Kubernetes 清单位于 `deploy/k8s/docker-desktop/`，面向 Docker Desktop 自带的单节点 Kubernetes。

### 1. 切换上下文

```powershell
kubectl config use-context docker-desktop
kubectl config current-context
```

### 2. 构建本地镜像

```powershell
pwsh .\deploy\k8s\docker-desktop\build-images.ps1
```

也可以指定镜像标签：

```powershell
pwsh .\deploy\k8s\docker-desktop\build-images.ps1 -Tag <tag>
```

### 3. 部署

```powershell
kubectl apply -k .\deploy\k8s\docker-desktop
kubectl get pods -n schill -w
```

### 4. 验证

```powershell
kubectl get all -n schill
kubectl logs -n schill deploy/gateway
kubectl logs -n schill deploy/user-rpc
kubectl logs -n schill deploy/search-api
kubectl logs -n schill job/schill-es-init
```

Kubernetes 默认对外入口：

- Gateway NodePort: `http://localhost:30086`
- MinIO API: `http://localhost:30090`
- MinIO Console: `http://localhost:30091`

### 5. 清理

```powershell
kubectl delete -k .\deploy\k8s\docker-desktop
```

本地 `hostPath` 数据不会随资源删除自动清空。

## 代码生成

项目使用 go-zero 的 API/RPC 代码生成模式。通常只在修改 `.proto` 或 `.api` 文件后重新生成。

### RPC 服务

示例：

```powershell
cd service/content/rpc
goctl rpc protoc content.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

其他 RPC 服务同理：

- `service/user/rpc/user.proto`
- `service/relation/rpc/relation.proto`
- `service/content/rpc/content.proto`
- `service/comment/rpc/comment.proto`
- `service/interaction/rpc/interaction.proto`
- `service/feed/rpc/feed.proto`

### API 服务

Search API 使用 `.api` 文件：

```powershell
cd service/search/api
goctl api go -api search.api -dir .
```

代码生成后应检查：

- 生成代码是否覆盖了自定义改动
- `internal/logic/` 中业务实现是否仍然正确
- `go test ./...` 是否通过
- `go mod tidy` 是否产生预期变更

## 测试与质量检查

后端：

```powershell
go test ./...
```

前端：

```powershell
cd web
npm run lint
npm run typecheck
npm run build
```

当前仓库中包含若干 benchmark 文件，例如：

- `common/cryptx/crypt_bench_test.go`
- `common/error/error_bench_test.go`
- `common/jwt/jwt_bench_test.go`
- `service/gateway/internal/handler/bench_test.go`
- `service/user/rpc/internal/logic/bench_test.go`

可以使用：

```powershell
go test -bench=. ./...
```

## 前端页面

前端位于 `web/`，主要页面包括：

| 路径 | 说明 |
| --- | --- |
| `/` | 首页/内容入口 |
| `/login` | 登录 |
| `/register` | 注册 |
| `/feed` | 信息流 |
| `/topics` | 话题 |
| `/search` | 搜索 |
| `/posts/new` | 发布帖子 |
| `/posts/[postId]` | 帖子详情 |
| `/posts/[postId]/edit` | 编辑帖子 |
| `/users/[userId]` | 用户主页 |
| `/collections` | 收藏相关页面 |

通用组件位于 `web/components/`，API 封装位于 `web/lib/api.ts`，类型定义位于 `web/lib/types.ts`。

## 事件主题

当前配置中出现的 Kafka 主题包括：

| 主题 | 生产/消费场景 |
| --- | --- |
| `post-created` | 内容服务发布，用户服务消费以更新发帖数 |
| `post-deleted` | 内容服务发布，用户服务消费以更新发帖数 |
| `user-followed` | 关系服务发布，用户服务消费以更新关注/粉丝统计 |
| `user-unfollowed` | 关系服务发布，用户服务消费以更新关注/粉丝统计 |
| `comment-create` | 评论创建异步处理 |
| `comment-created` | 评论创建完成，内容服务消费以更新帖子评论数 |
| `comment-deleted` | 评论删除，内容服务消费以更新帖子评论数 |
| `comment-vote` | 评论投票异步处理 |
| `comment-dlq` | 评论消费失败死信队列 |
| `post-star` | 帖子点赞事件 |
| `post-unstar` | 帖子取消点赞事件 |
| `post-collect` | 帖子收藏事件 |
| `post-uncollect` | 帖子取消收藏事件 |
| `interaction-dlq` | 互动消费失败死信队列 |
| `canal_topic` | Canal Server 输出 MySQL Binlog 变更 |

## 缓存与一致性

项目使用 Redis 和本地缓存承担高频读取与互动状态查询：

- `common/redis/keys.go` 维护 Redis Key 约定。
- `common/redis/lua/` 包含点赞、收藏等原子操作 Lua 脚本。
- `content-rpc` 配置了 `LocalCache`，用于降低热点内容读压力。
- Kafka 事件用于跨服务计数同步，整体采用最终一致性模型。
- Canal 同步链路用于搜索索引更新，搜索结果和 MySQL 主库之间可能存在短暂延迟。

## OpenSpec 文档

`openspec/` 保存需求和变更设计：

- `openspec/specs/user-service/spec.md`
- `openspec/specs/feed-system/spec.md`
- `openspec/changes/add-feed-system/`
- `openspec/changes/align-user-service-spec/`
- `openspec/changes/build-enterprise-frontend-pages/`
- `openspec/changes/convert-user-to-rpc-only-service/`
- `openspec/changes/patch-user-profile-optional-fields/`

这些文档适合用于理解历史设计决策、需求边界和待办任务。

## Swagger 与其他文档

Swagger 相关文件位于 `docs/swagger/`：

- `docs/swagger/index.html`
- `docs/swagger/content.json`
- `docs/swagger/comment.json`
- `docs/swagger/interaction.json`
- `docs/swagger/relation.json`

其他文档包括：

- `docs/profile-redesign/`：个人主页改版原型和设计计划
- `docs/interview/`：系统升级、Kubernetes 等补充资料
- `deploy/k8s/docker-desktop/README.md`：Docker Desktop Kubernetes 部署说明
- `service/canal/README.md`：Canal 同步服务说明

## 常见问题

### 1. RPC 服务启动后 Gateway 调不到服务

检查 Etcd 是否启动，以及各服务配置中的 `Etcd.Hosts` 和 `Key` 是否一致：

```powershell
docker ps
docker logs etcd
```

本地源码运行使用 `127.0.0.1:2379`，容器/Kubernetes 中通常使用 `etcd:2379`。

### 2. 数据库表不存在

确认 `db.sql` 是否已导入到当前 MySQL 数据库。对于已存在的数据卷，MySQL 初始化脚本不会重复执行，需要手动导入。

### 3. 搜索无结果

依次检查：

1. Elasticsearch 是否可访问。
2. `user`、`post`、`topic` 索引是否已创建。
3. Canal Server 是否正确监听 MySQL Binlog。
4. Kafka `canal_topic` 是否有消息。
5. `canal-sync` 是否正常消费并写入 Elasticsearch。

### 4. 前端请求失败

检查：

1. Gateway 是否启动并监听 `8086`。
2. 前端 `NEXT_PUBLIC_API_BASE_URL` 是否指向正确地址。
3. 浏览器控制台是否存在 CORS、401、500 等错误。
4. Gateway 日志和对应 RPC 服务日志。

### 5. Kafka 消费没有更新计数

检查：

1. Kafka 是否启动。
2. 生产者和消费者配置的 topic 名称是否一致。
3. Consumer group 是否正常消费。
4. 对应业务服务是否启动，例如用户统计依赖 `user-rpc` 消费帖子/关注事件。

## 开发约定

- 后端业务遵循 go-zero 常见分层：handler 负责 HTTP 接入，logic 负责业务，model 负责数据访问。
- RPC 接口优先通过 `.proto` 定义，再使用 goctl/protoc 生成代码。
- Search API 通过 `.api` 定义 HTTP 接口，再生成 handler/types。
- 公共逻辑放入 `common/`，业务服务内的私有逻辑放在各自 `internal/` 下。
- 配置不要硬编码在业务逻辑中，应放入服务 `etc/*.yaml` 或部署配置。
- 涉及跨服务统计更新时，优先通过 Kafka 事件和消费者维护最终一致性。
- 修改数据库结构后，同步更新 `db.sql`、模型代码和相关文档。
- 修改搜索字段后，同步更新 `es_index/` Mapping、Canal 同步逻辑和 Search API。

## 安全注意事项

仓库中的默认配置偏向本地开发，例如：

- JWT Secret 使用示例值
- MySQL 默认账号密码示例为 `root:123456`
- Elasticsearch 关闭了安全认证
- Etcd 允许无认证访问
- MinIO 使用环境变量传入默认账号密码

部署到真实环境前必须：

- 更换所有密钥和密码
- 开启必要的网络访问控制
- 为 Elasticsearch、Etcd、MySQL、Redis、MinIO 配置认证和最小权限
- 使用独立配置管理或 Secret 管理敏感信息
- 配置日志、监控、告警、备份和恢复策略

## 快速验证

后端 Gateway 健康检查：

```powershell
curl http://localhost:8086/health
```

注册用户示例：

```powershell
curl -X POST http://localhost:8086/api/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"demo","password":"demo123456"}'
```

登录示例：

```powershell
curl -X POST http://localhost:8086/api/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"demo","password":"demo123456"}'
```

搜索帖子示例：

```powershell
curl "http://localhost:8086/api/search/post?keyword=demo&page=1&pageSize=10"
```

## 当前状态说明

本 README 根据当前仓库源码、配置、数据库脚本、前端 API 封装和部署清单整理。部分既有中文注释或文档存在编码显示异常，但不影响从代码结构和配置识别项目功能。若后续服务接口、数据库结构或部署方式发生变化，请同步更新本文档。
