# 个人主页前端重新设计方案

## 目标

本次 redesign 聚焦个人主页的信息组织和入口收敛：

- 首页不再承担“我的收藏”入口，避免全局导航过载。
- 收藏入口放入个人主页，并只在访问自己的主页时展示。
- 个人主页从“资料卡 + 文章列表”升级为“身份、数据、内容、收藏、关系”的聚合中心。
- UI 延续当前项目的海盐蓝白风格，保持产品化、清爽、可长期使用。

## 当前现状

前端已有能力：

- `web/app/users/[userId]/page.tsx`：个人主页，展示用户资料、统计、发布文章。
- `web/app/collections/page.tsx`：我的收藏独立页面。
- `web/components/AppHeader.tsx`：登录后 Header 内有收藏入口。
- `web/lib/api.ts`：已有 `getUserInfo`、`getPostList`、`getMyCollections` 等接口封装。

后端已有接口：

- `GET /api/users/:id`：用户资料、profile、stat。
- `GET /api/posts?userId=&page=&pageSize=`：指定用户文章列表。
- `GET /api/users/me/collections?page=&pageSize=`：当前用户收藏列表。
- `GET /api/relation/following`、`GET /api/relation/follower`：关注/粉丝列表。
- `POST /api/users/:id/follow`、`DELETE /api/users/:id/follow`：关注操作。

## 信息架构

### 全局 Header 调整

移除登录态 Header 的“收藏”按钮。

保留：

- 首页
- 动态
- 搜索
- 话题
- 发布
- 用户入口
- 退出

收藏入口迁移到：

- 我的个人主页主导航 Tab：收藏
- 我的个人主页侧栏快捷入口：我的收藏
- 空状态引导：去首页发现内容

### 个人主页一级结构

首屏建议使用以下结构：

1. 顶部身份区
   - 头像
   - 用户名
   - 状态
   - 签名
   - 关注/编辑资料按钮
   - 注册时间、地区、公司、职位、网站

2. 数据概览
   - 文章
   - 评论
   - 粉丝
   - 关注
   - 获赞
   - 收藏

3. 内容导航
   - 文章
   - 收藏，仅本人可见
   - 关注
   - 粉丝
   - 资料

4. 主内容区
   - 当前 Tab 内容列表

5. 右侧辅助区，桌面端展示
   - 个人资料完整度
   - 最近活跃
   - 个人标签
   - 本人视角快捷入口

## 三版布局方向

### A 版：内容优先

适合当前社区产品。上方强化用户身份，中间用 Tab 切换文章/收藏，右侧保留轻量资料。

优点：

- 和当前代码改造成本最低。
- 收藏入口自然进入个人主页。
- 文章列表仍是主体验。

建议作为第一期落地版本。

### B 版：成长仪表盘

强调个人成长、徽章、资料完整度、互动数据。更有二次元陪伴感，但仍保持产品化。

优点：

- 个人中心感更强。
- 适合后续扩展等级、徽章、任务。

风险：

- 当前后端暂未提供等级、徽章、成长值，需要先用静态或本地派生数据，不能过早做重。

### C 版：收藏管理优先

把收藏做成个人主页内的二级工作台，适合用户经常回看内容。

优点：

- 明确满足“收藏页面按钮别放主页，放个人主页里面”。
- 收藏列表的查找、筛选、批量管理可以顺滑扩展。

风险：

- 如果收藏不是高频行为，首屏信息会偏工具化。

## 推荐方案

推荐采用 A 版作为正式改版基础，吸收 B 版的资料完整度模块和 C 版的收藏筛选能力。

最终布局：

- Header 移除收藏入口。
- `/users/[userId]` 增加 `tab` 查询参数：
  - `/users/123?tab=posts`
  - `/users/123?tab=collections`
  - `/users/123?tab=following`
  - `/users/123?tab=followers`
  - `/users/123?tab=profile`
- 本人访问自己的页面时显示收藏 Tab。
- 他人访问时隐藏收藏 Tab，避免暴露私有收藏。
- `/collections` 页面可保留为兼容路由，但建议重定向到 `/users/{me}?tab=collections`。

## UI 对应

### 顶部身份区

组件建议：

- Avatar：64-88px，使用头像；无头像时使用用户名首字。
- Action：
  - 本人：编辑资料、进入收藏
  - 他人：关注/已关注、私信预留
- 状态 Badge：正常、禁言、冻结。

### 数据区

六个指标横向排列，移动端两行三列。

字段映射：

- `stat.postCount`
- `stat.commentCount`
- `stat.followerCount`
- `stat.followingCount`
- `stat.likeCount`
- `stat.collectionCount`

### Tab 区

Tab 映射：

- 文章：`GET /api/posts?userId=:id`
- 收藏：`GET /api/users/me/collections`
- 关注：`GET /api/relation/following?userId=:id`
- 粉丝：`GET /api/relation/follower?userId=:id`
- 资料：`GET /api/users/:id`

### 收藏列表

收藏列表放在个人主页中，建议支持：

- 按收藏时间倒序，沿用当前后端默认。
- 搜索收藏，二期新增。
- 标签筛选，二期新增。
- 取消收藏，复用 `DELETE /api/posts/:id/collect`。

## 后端接口对应

### 一期复用现有接口

| 页面模块 | 方法 | 接口 | 说明 |
| --- | --- | --- | --- |
| 用户资料 | GET | `/api/users/:id` | 返回 userInfo/profile/stat |
| 用户文章 | GET | `/api/posts?userId=:id&page=&pageSize=` | 返回文章列表 |
| 我的收藏 | GET | `/api/users/me/collections?page=&pageSize=` | 仅登录本人 |
| 关注列表 | GET | `/api/relation/following?userId=:id&page=&pageSize=` | 已有前端封装 |
| 粉丝列表 | GET | `/api/relation/follower?userId=:id&page=&pageSize=` | 已有前端封装 |
| 关注用户 | POST | `/api/users/:id/follow` | 他人主页按钮 |
| 取消关注 | DELETE | `/api/users/:id/follow` | 他人主页按钮 |
| 取消收藏 | DELETE | `/api/posts/:id/collect` | 收藏 Tab 内操作 |

### 二期建议新增接口

#### `GET /api/users/me/profile-home`

用于减少个人主页首屏多接口并发。

响应建议：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "userInfo": {},
    "profile": {},
    "stat": {},
    "viewer": {
      "isSelf": true,
      "isFollowing": false
    },
    "recentPosts": [],
    "recentCollections": []
  }
}
```

#### `GET /api/users/:id/viewer-state`

用于他人主页关注态。

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "isSelf": false,
    "isFollowing": true
  }
}
```

#### `GET /api/users/me/collections/search`

用于收藏页内搜索和筛选。

查询参数：

- `keyword`
- `tag`
- `page`
- `pageSize`

## 前端实现建议

### 路由

保留：

- `/users/[userId]`

新增或扩展：

- `/users/[userId]?tab=posts`
- `/users/[userId]?tab=collections`

兼容：

- `/collections` 登录后跳转到 `/users/{currentUserId}?tab=collections`

### 组件拆分

建议拆为：

- `ProfileHeader`
- `ProfileStats`
- `ProfileTabs`
- `ProfilePostList`
- `ProfileCollectionList`
- `ProfileRelationList`
- `ProfileSidePanel`

### 权限规则

- 未登录可以访问他人主页和文章 Tab。
- 未登录访问收藏 Tab，跳转登录或显示登录引导。
- 登录用户访问自己的主页，显示收藏 Tab。
- 登录用户访问他人主页，不显示收藏 Tab。

## 原型文件

本方案包含三版 HTML 原型：

- `prototype-a-content-first.html`
- `prototype-b-growth-dashboard.html`
- `prototype-c-collection-workbench.html`

直接用浏览器打开即可查看。
