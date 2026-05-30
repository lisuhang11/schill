# Swagger API 文档

本目录存放通过 goctl 从 API 文件生成的 Swagger 文档。

## 📁 文件说明

- `comment.json` - 评论服务 API 文档
- `content.json` - 内容服务 API 文档
- `interaction.json` - 互动服务 API 文档
- `relation.json` - 关系服务 API 文档
- `index.html` - Swagger UI 集成页面（推荐使用）

## 🚀 快速开始

### 方式一：使用集成的 Swagger UI（推荐）

#### Windows 用户

双击运行 `serve.ps1`，或在 PowerShell 中执行：

```powershell
cd docs/swagger
.\serve.ps1
```

#### Linux/Mac 用户

```bash
cd docs/swagger
chmod +x serve.sh
./serve.sh
```

然后在浏览器中访问：**http://localhost:8080**

### 方式二：使用 VS Code Live Server

1. 安装 [Live Server 扩展](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer)
2. 在 VS Code 中右键点击 `index.html`
3. 选择 "Open with Live Server"

### 方式三：使用在线 Swagger UI

访问 [https://petstore.swagger.io/](https://petstore.swagger.io/)，然后：
- 点击 "Explore" 按钮
- 上传你的 JSON 文件或输入文件的 URL

## 🔧 生成命令

如需重新生成 Swagger 文档，使用以下命令：

```bash
# 生成所有服务的 swagger 文档
goctl api swagger --api service/comment/api/comment.api --dir docs/swagger --filename comment
goctl api swagger --api service/content/api/content.api --dir docs/swagger --filename content
goctl api swagger --api service/interaction/api/interaction.api --dir docs/swagger --filename interaction
goctl api swagger --api service/relation/api/relation.api --dir docs/swagger --filename relation

# 如果需要 yaml 格式，添加 --yaml 参数
goctl api swagger --api service/comment/api/comment.api --dir docs/swagger --filename comment --yaml
```

## 📚 使用说明

### Swagger UI 功能

- 📖 **查看 API 文档** - 清晰展示所有接口、参数、响应格式
- 🧪 **在线测试** - 直接在页面中调用 API 进行测试
- 📋 **复制请求** - 一键复制 curl 命令或请求代码
- 🏷️ **分组查看** - 按服务和功能分组查看接口

### Try it out 测试

1. 找到你要测试的接口
2. 点击 "Try it out" 按钮
3. 填写必要的参数
4. 点击 "Execute" 发送请求
5. 查看响应结果

## 🔐 认证说明

部分接口需要 JWT 认证，测试时需要：

1. 先调用登录接口获取 token
2. 点击页面右上角的 "Authorize" 按钮
3. 输入 `Bearer <你的token>`
4. 点击 "Authorize" 完成认证
5. 现在可以测试需要认证的接口了

## 📝 更新文档

当 API 文件更新后，重新运行生成命令即可更新 Swagger 文档。
