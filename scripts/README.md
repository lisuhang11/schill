# SoChill 服务启动脚本说明

## 📁 脚本列表

| 脚本 | 说明 | 推荐使用 |
|------|------|---------|
| `start-all-separate.ps1` | 每个服务独立窗口 | ⭐ 推荐 |
| `start-all.ps1` | 所有服务在后台运行 | |

---

## 🚀 使用方法

### 方法一：独立窗口版（推荐）

```powershell
# 在 PowerShell 中执行
cd scripts
.\start-all-separate.ps1
```

**特点**：
- 每个服务打开独立的 PowerShell 窗口
- 可以方便查看每个服务的日志输出
- 关闭窗口即可停止对应服务
- 先启动 RPC 服务，再启动 API 服务

**启动顺序**：
1. User RPC
2. Relation RPC
3. Content RPC
4. Comment RPC
5. User API
6. Relation API
7. Content API
8. Comment API

---

### 方法二：后台运行版

```powershell
# 在 PowerShell 中执行
cd scripts
.\start-all.ps1
```

**特点**：
- 所有服务在后台运行
- 按 `Ctrl+C` 停止所有服务
- 无法看到实时日志输出

---

## 📋 服务端口说明

| 服务 | 类型 | 端口 | 配置文件 |
|------|------|------|---------|
| User API | REST API | 8888 | service/user/api/etc/user-api.yaml |
| User RPC | gRPC | 8081 | service/user/rpc/etc/user.yaml |
| Relation API | REST API | 8889 | service/relation/api/etc/relation-api.yaml |
| Relation RPC | gRPC | 8082 | service/relation/rpc/etc/relation.yaml |
| Content API | REST API | 8890 | service/content/api/etc/content-api.yaml |
| Content RPC | gRPC | 8083 | service/content/rpc/etc/content.yaml |
| Comment API | REST API | 8892 | service/comment/api/etc/comment-api.yaml |
| Comment RPC | gRPC | 8085 | service/comment/rpc/etc/comment.yaml |

---

## 🔧 前置要求

1. **Go 环境**
   - 已安装 Go 1.19+
   - `go version` 检查版本

2. **依赖服务**
   - MySQL 数据库（已配置并运行）
   - Etcd 服务发现（已配置并运行）
   - Kafka 消息队列（已配置并运行）

3. **数据库**
   - 已执行 `db.sql` 初始化数据库
   - 数据库连接配置正确

---

## ⚠️ 常见问题

### 1. PowerShell 执行策略限制

如果遇到以下错误：
```
无法加载文件，因为在此系统上禁止运行脚本
```

**解决方法**：
```powershell
# 临时允许（当前会话有效）
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# 然后再运行脚本
.\start-all-separate.ps1
```

### 2. 端口被占用

如果某个端口被占用，可以：
1. 关闭占用端口的程序
2. 或修改对应服务的配置文件中的端口号

### 3. 依赖服务未启动

确保以下服务已启动：
- MySQL
- Etcd
- Kafka

---

## 🛑 停止服务

### 独立窗口版
直接关闭各个服务的 PowerShell 窗口即可。

### 后台运行版
在运行脚本的窗口按 `Ctrl+C`，会自动停止所有服务。

---

## 📝 注意事项

1. **启动顺序**：RPC 服务必须在 API 服务之前启动
2. **日志查看**：推荐使用独立窗口版，方便查看日志
3. **开发调试**：建议先启动 RPC 服务，确认正常后再启动 API 服务

---

## 🔗 相关文档

- 产品需求文档：`web/产品需求文档.md`
- 前端开发计划：`web/前端开发计划.md`
- 数据库脚本：`db.sql`
