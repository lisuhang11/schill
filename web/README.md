# SChill 前端项目

基于 Vue 3 + Element Plus 的社交平台前端项目。

## 技术栈

- **框架**: Vue 3 (Composition API) + TypeScript
- **UI 组件库**: Element Plus
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **HTTP 客户端**: Axios
- **构建工具**: Vite

## 项目结构

```
web/
├── src/
│   ├── api/              # API 接口封装
│   │   ├── user.ts       # 用户相关 API
│   │   ├── content.ts    # 内容相关 API
│   │   ├── comment.ts   # 评论相关 API
│   │   └── relation.ts  # 关系相关 API
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   │   ├── Layout.vue
│   │   ├── PostCard.vue
│   │   ├── CommentList.vue
│   │   ├── CommentItem.vue
│   │   ├── CommentInput.vue
│   │   └── FollowButton.vue
│   ├── router/           # 路由配置
│   │   └── index.ts
│   ├── stores/           # Pinia 状态管理
│   │   └── user.ts
│   ├── utils/            # 工具函数
│   │   ├── request.ts    # Axios 封装
│   │   └── date.ts     # 日期格式化
│   ├── views/            # 页面组件
│   │   ├── LoginPage.vue
│   │   ├── RegisterPage.vue
│   │   ├── HomePage.vue
│   │   ├── PostDetailPage.vue
│   │   ├── CreatePostPage.vue
│   │   ├── UserProfilePage.vue
│   │   ├── UserSettingsPage.vue
│   │   ├── FollowingListPage.vue
│   │   └── FollowerListPage.vue
│   ├── types/            # TypeScript 类型定义
│   │   └── index.ts
│   ├── App.vue          # 根组件
│   └── main.ts          # 入口文件
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md
```

## 功能模块

### 1. 用户认证模块
- 用户注册
- 用户登录
- Token 刷新
- 退出登录

### 2. 用户模块
- 查看用户基本信息
- 查看用户扩展信息
- 查看用户统计信息
- 更新用户头像
- 更新用户扩展信息

### 3. 内容模块
- 创建帖子
- 编辑帖子
- 删除帖子
- 获取帖子列表
- 获取帖子详情
- 增加浏览量

### 4. 评论模块
- 创建评论
- 删除评论
- 评论点赞/点踩
- 获取评论列表（抖音风格，两层展开）

### 5. 关系模块
- 关注用户
- 取消关注
- 获取关注列表
- 获取粉丝列表
- 检查关注状态

## 快速开始

### 安装依赖

```bash
cd web
npm install
```

### 开发模式

```bash
npm run dev
```

项目将在 `http://localhost:3000` 启动。

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## API 代理配置

开发环境中，API 请求会通过 Vite 代理到后端服务：

- 前端请求: `/api/*`
- 代理目标: `http://localhost:8888/*`

可在 `vite.config.ts` 中修改配置。

## 后端服务

确保后端服务在 `http://localhost:8888` 运行。
