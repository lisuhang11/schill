# SChill 服务启动脚本

本目录包含用于启动 SChill 所有服务的脚本。

## 服务列表

项目包含以下服务（按启动顺序）：

### RPC 服务
1. **user-rpc** - 用户 RPC 服务
2. **content-rpc** - 内容 RPC 服务
3. **relation-rpc** - 关系 RPC 服务

### API 服务
4. **user-api** - 用户 API 服务
5. **content-api** - 内容 API 服务
6. **relation-api** - 关系 API 服务
7. **comment-api** - 评论 API 服务

## 使用方法

### Windows 系统

使用 PowerShell 脚本：

```powershell
# 进入 scripts 目录
cd scripts

# 运行启动脚本
.\start-all.ps1
```

**注意**：如果遇到执行策略限制，需要先允许脚本执行：

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Linux/Mac 系统

使用 Shell 脚本：

```bash
# 进入 scripts 目录
cd scripts

# 添加执行权限
chmod +x start-all.sh

# 运行启动脚本
./start-all.sh
```

## 停止服务

在运行脚本的终端窗口中按 `Ctrl+C` 即可停止所有服务。

## 前置条件

1. 确保已安装 Go 1.19+
2. 确保数据库（MySQL）、Redis、ETCD 等基础设施已启动
3. 确保配置文件（`etc/*.yaml`）中的配置正确

## 注意事项

- RPC 服务需要先于 API 服务启动
- 每个服务会在单独的窗口/进程中运行
- 脚本会自动处理服务的启动和清理
