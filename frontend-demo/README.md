# LAF 校园失物招领 · 前端示例（frontend-demo）

> 本文件夹是**后端能力的完整前端示例**，仅供前端同学参考/复用，不属于后端编译产物。
> 技术栈：**Vue 3 全家桶 —— Vue 3 + Vue Router 4 + Pinia + axios + Vite**。

这是一个可直接 `npm install && npm run dev` 跑起来的完整工程，**对接后端全部现有接口**，并做了统一的视觉风格（卡片化、清新绿主题、状态徽章、轻提示、分页等）。

---

## 一、快速开始

```bash
cd frontend-demo
npm install
npm run dev        # 打开 http://localhost:5173
```

- 开发服务器已配置 **proxy**：`/api` 与 `/uploads` 均转发到 `http://localhost:8080`，本地无需处理跨域。
- 后端需先启动（`go run main.go`，默认 8080），并按 `migrations/tables.sql` 建表。

### 环境变量

复制 `.env.example` 为 `.env`（或 `.env.local`）：

```bash
VITE_API_BASE_URL=/api/v1      # 走 dev proxy（推荐本地开发）
# VITE_API_BASE_URL=http://10.0.0.5:8080/api/v1   # 前后端分离且已配置 CORS 时直连
```

---

## 二、目录结构

```text
frontend-demo/
├── index.html
├── package.json
├── vite.config.js            # Vue 插件 + @ 别名 + /api、/uploads 代理
├── .env.example
└── src/
    ├── main.js               # 入口：装配 Pinia + Router
    ├── App.vue               # 根组件：导航 + 路由出口 + 轻提示
    ├── router/index.js       # 路由表 + 登录/角色守卫
    ├── styles/main.css       # 设计系统（CSS 变量 + 基础组件类）
    ├── api/                  # 按模块拆分的接口封装
    │   ├── http.js           # axios 实例 + 拦截器 + 数组参数序列化 + token 存取
    │   ├── auth.js  post.js  comment.js
    │   ├── appeal.js  geo.js  conversation.js  admin.js
    ├── stores/               # Pinia
    │   ├── auth.js           # 登录态、当前用户、角色判断
    │   ├── toast.js          # 全局轻提示
    │   └── location.js       # 定位结果缓存
    ├── components/
    │   ├── AppNavbar.vue     # 顶部导航（按角色显隐入口）
    │   ├── PostCard.vue      # 帖子卡片
    │   ├── StatusBadge.vue   # 类型/状态/完成 徽章
    │   ├── Pagination.vue    # 通用分页
    │   ├── EmptyState.vue    # 空状态
    │   ├── ToastHost.vue     # 轻提示容器
    │   └── LocationPicker.vue# 校园地点选择（自动定位 + 手动选择 + 补充说明）
    └── views/
        ├── LoginView.vue  RegisterView.vue  AppealView.vue
        ├── PostListView.vue  PostCreateView.vue  PostDetailView.vue
        ├── ProfileView.vue
        ├── ConversationListView.vue  ChatView.vue
        ├── NotFoundView.vue
        └── admin/
            ├── AdminLayout.vue        # 管理后台布局（子导航）
            ├── ReviewsView.vue        # 待审批（帖子 + 申诉）
            ├── DeletedPostsView.vue   # 已删帖子 + 恢复
            ├── AppealsView.vue        # 申诉审核（mainadmin）
            └── UsersView.vue          # 用户注销/恢复（mainadmin）
```

---

## 三、通用约定（所有接口都适用）

- **统一响应信封**：`{ "code": <业务码>, "msg": "<说明>", "data": <数据> }`。成功 `code=0`。
  - 已在 `api/http.js` 的响应拦截器统一处理：`code===0` 直接返回 `data`；否则抛 `{code,msg}`；HTTP 401 自动清 token 并派发 `laf:unauthorized` 事件。
- **鉴权头**：`Authorization: Bearer <access_token>`（请求拦截器自动注入，token 存 `localStorage`）。
- **分页**：`page`（从 1 开始）+ `page_size`（默认 20，上限 100），返回 `{list,total,page,page_size}`。
- **数组参数**：后端用可重复参数读取（如 `type`），`http.js` 里自定义了 `paramsSerializer`，会序列化为 `type=lost&type=found`，**不要用 axios 默认的 `type[]=` 形式**。
- **multipart**：仅“发布帖子”用 `FormData`，**不要手动设 `Content-Type`**（浏览器会自动带 boundary；示例中显式设置 `multipart/form-data` 亦可，axios 会补 boundary）。

---

## 四、接口对接总览（全覆盖）

> 鉴权列：`公开`=无需登录；`可选`=带 token 识别身份；`需登录`=必须 token；角色列标注管理员限制。

### 认证与用户（`/api/v1/auth`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/auth/register` | POST | 公开 | `views/RegisterView.vue` → `api/auth.register` |
| `/auth/login` | POST | 公开 | `views/LoginView.vue` → `api/auth.login` |
| `/auth/profile` | GET | 需登录 | `views/ProfileView.vue` → `api/auth.getProfile` |
| `/auth/profile` | PATCH | 需登录 | `views/ProfileView.vue` → `api/auth.updateProfile` |
| `/auth/password` | PATCH | 需登录 | `views/ProfileView.vue` → `api/auth.changePassword` |
| `/auth/account` | DELETE | 需登录 | `views/ProfileView.vue` → `api/auth.deleteAccount` |

### 帖子（`/api/v1/posts`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/posts` | GET | 可选 | `views/PostListView.vue`（类型/完成/管理员状态筛选） → `api/post.listPosts` |
| `/posts` | POST | 需登录 | `views/PostCreateView.vue`（multipart + 位置） → `api/post.createPost` |
| `/posts/:post_id` | GET | 可选 | `views/PostDetailView.vue` → `api/post.getPost` |
| `/posts/:post_id` | DELETE | 需登录 | `views/PostDetailView.vue` → `api/post.deletePost` |
| `/posts/:post_id/recover` | PATCH | 需登录 | `views/admin/DeletedPostsView.vue` → `api/post.recoverPost` |
| `/posts/:post_id/review` | PATCH | 管理员 | `views/PostDetailView.vue`、`views/admin/ReviewsView.vue` → `api/post.reviewPost` |
| `/posts/:post_id/comments` | GET | 公开 | `views/PostDetailView.vue` → `api/post.listPostComments` |
| `/posts/:post_id/conversations` | POST | 需登录 | `views/PostDetailView.vue`（申领/召领） → `api/post.startConversation` |

### 评论（`/api/v1/comments`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/comments` | POST | 需登录 | `views/PostDetailView.vue` → `api/comment.createComment` |
| `/comments/:comment_id` | DELETE | 需登录 | `views/PostDetailView.vue` → `api/comment.deleteComment` |

### 申诉（`/api/v1/appeals`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/appeals` | POST | 公开 | `views/AppealView.vue` → `api/appeal.submitAppeal` |

### 地理位置（`/api/v1/geo`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/geo/locations` | GET | 公开 | `components/LocationPicker.vue` → `api/geo.fetchCampusLocations` |
| `/geo/locate` | POST | 公开 | `components/LocationPicker.vue` → `api/geo.locate` |

### 对话 / 完成寻找（`/api/v1/conversations`）

| 接口 | 方法 | 鉴权 | 前端位置 |
|---|---|---|---|
| `/conversations` | GET | 需登录 | `views/ConversationListView.vue` → `api/conversation.listMyConversations` |
| `/conversations/:id/messages` | GET | 需登录 | `views/ChatView.vue` → `api/conversation.listMessages` |
| `/conversations/:id/messages` | POST | 需登录 | `views/ChatView.vue` → `api/conversation.sendMessage` |
| `/conversations/:id/finish-requests` | POST | 需登录 | `views/ChatView.vue` → `api/conversation.createFinishRequest` |
| `/conversations/:id/finish-requests/:rid` | PATCH | 需登录 | `views/ChatView.vue` → `api/conversation.handleFinishRequest` |

### 管理员（`/api/v1/admin`）

| 接口 | 方法 | 角色 | 前端位置 |
|---|---|---|---|
| `/admin/posts/:post_id/status` | PATCH | postadmin/mainadmin | `api/admin.updatePostStatus`（可直接在后台调用） |
| `/admin/posts/deleted` | GET | postadmin/mainadmin | `views/admin/DeletedPostsView.vue` → `api/admin.listDeletedPosts` |
| `/admin/users/:user_id` | DELETE | mainadmin | `views/admin/UsersView.vue` → `api/admin.deleteUser` |
| `/admin/users/:user_id/recover` | PATCH | mainadmin | `views/admin/UsersView.vue` → `api/admin.recoverUser` |
| `/admin/appeals/:appeal_id/review` | PATCH | mainadmin | `views/admin/AppealsView.vue`、`ReviewsView.vue` → `api/admin.reviewAppeal` |
| `/admin/reviews` | GET | mainadmin | `views/admin/ReviewsView.vue`、`AppealsView.vue` → `api/admin.listReviews` |

---

## 五、页面与角色可见性

| 页面 | 路由 | 需要登录 | 角色限制 |
|---|---|---|---|
| 帖子广场 | `/posts` | 否 | — |
| 发布帖子 | `/posts/new` | 是 | — |
| 帖子详情 | `/posts/:id` | 否 | 管理员可见所有状态 |
| 我的 | `/me` | 是 | — |
| 我的消息 | `/conversations` | 是 | — |
| 会话聊天 | `/conversations/:id` | 是 | 参与方 |
| 登录/注册/申诉 | `/login` `/register` `/appeal` | 否 | — |
| 管理后台 | `/admin/*` | 是 | postadmin / mainadmin |
| 申诉审核、用户管理 | `/admin/appeals` `/admin/users` | 是 | mainadmin |

> 导航栏与后台子菜单都会按当前角色显隐；路由守卫在 `router/index.js` 统一拦截 `requiresAuth` 与 `meta.roles`。

---

## 六、关键实现说明

1. **响应信封与错误处理**：所有 `api/*.js` 直接返回 `data`；拦截器把 `code!==0` 转成 `{code,msg}` 抛出，视图层 `try/catch` 后用 `toast.error(e.msg)` 提示，401 自动登出。
2. **数组筛选**：帖子类型/状态多选，`paramsSerializer` 保证重复参数形式（`type=lost&type=found`）。
3. **帖子位置**：`PostCreateView` 复用 `LocationPicker`，拿到地点后提交 `location_id` 或坐标 + `supplement`；列表/详情直接展示后端返回的 `location_name`（冗余快照），无需再查地点接口。
4. **完成寻找**：`ChatView` 展示待处理申请横幅，发起方显示“撤回”，另一方显示“同意/拒绝”；同意后帖子置为已完成。
5. **管理员审核**：`PostDetailView`（单帖快捷审核）与 `admin/ReviewsView`（待办工作台）两处入口。

---

## 七、注意事项

1. **定位需 HTTPS 或 localhost**：`navigator.geolocation` 在非安全上下文被禁用，此时只能手动选择地点。
2. **坐标系**：浏览器与后端预设均为 WGS-84；接入高德等 GCJ-02 地图时需先纠偏。
3. **用户列表**：后端未提供“用户列表/搜索”接口，`UsersView` 以“输入用户 ID”为入口执行注销/恢复。
4. **图片**：帖子图片 URL 形如 `http://localhost:8080/uploads/posts/...`，开发环境经 `/uploads` 代理直接可显示。
5. **管理员账号**：代码中注册角色校验已临时放开，可直接注册 `postadmin` / `mainadmin` 用于联调；生产环境请由后台创建。
