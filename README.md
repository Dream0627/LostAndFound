# LAF 失物招领后端 API 文档

校园失物招领平台的 Go 后端，技术栈 **Go + Gin + GORM(MySQL) + Viper + JWT + bcrypt**。

- Base URL：`http://localhost:8080`
- 统一响应信封：`{ "code": <业务码>, "msg": "<说明>", "data": <数据> }`
- 成功：HTTP 200，`code = 0`，`msg = "success"`；失败：按错误类型返回对应 HTTP 状态码与业务码。
- 鉴权头：`Authorization: Bearer <access_token>`（`Bearer` 后须有**一个空格**）

---

## 快速上手

1. **配置**：复制 `config/config.example.yaml` 为 `config/config.yaml`，填写 MySQL 连接与 JWT secret（该文件已加入 `.gitignore`）。
2. **建表**：执行 `migrations/tables.sql`（按顺序 DROP → CREATE）。
3. **启动**：`go run main.go`（默认端口 8080）。
4. **注册**：调用「注册」接口创建账号。
5. **登录**：调用「登录」接口拿到 `access_token`，后续鉴权接口在请求头携带该 token。

---

## 角色与用户限制

| 角色 | 说明 | 权限范围 |
|------|------|----------|
| `student` | 普通用户 | 只能操作本人的帖子/评论；发帖后状态为 `pending`（待审核）；列表/详情仅能看到 `approved` 帖子 |
| `postadmin` | 帖子管理员 | 可管理任意帖子/评论；发帖直接 `approved`；可使用全部帖子管理接口 |
| `mainadmin` | 超级管理员 | 拥有 `postadmin` 的全部权限（角色白名单中二者并列；删除用户等接口目前为预留、未接线） |

> 说明：权限采用角色白名单机制，`RequireRole([]string{"postadmin","mainadmin"})` 即表示两类管理员均可访问。

---

## 错误码速查

| HTTP | code | 含义 | 典型场景 |
|------|------|------|----------|
| 200 | 0 | success | 成功 |
| 400 | 400 | 参数校验失败 | 必填缺失、长度/格式非法 |
| 400 | 400 | 原密码错误 | 改密时原密码不匹配 |
| 400 | 400 | 无效的帖子状态 | 传入非 `pending/approved/rejected` |
| 400 | 400 | 该帖子不可审核 | 对非 `pending` 帖子调用审核接口 |
| 401 | 401 | 未登录或令牌无效 | 缺少/过期/伪造 token |
| 401 | 401 | 账号或密码错误 | 登录失败 |
| 403 | 403 | 无权操作他人资源 | student 操作（删除/恢复）他人帖子/评论 |
| 403 | 403 | 仅管理员可操作 | 非管理员调用管理员接口 |
| 404 | 404 | 帖子不存在 | 帖子 ID 错误、已软删除，或对普通用户隐藏 |
| 404 | 404 | 用户不存在 | 用户 ID 错误 |
| 404 | 404 | 评论不存在 | 评论 ID 错误或已软删除 |
| 409 | 409 | 该学号或工号已存在 | 重复注册（学号/工号已被占用） |
| 1000 | 1000 | 系统异常 | 未知服务端错误（HTTP 落 500） |
| 1001 | 1001 | 数据库操作失败 | 数据库连接/查询/写入异常（HTTP 落 500） |

---

## 一、认证与用户模块（/api/v1/auth）

### 1. 注册
- **接口**：`POST /api/v1/auth/register`
- **鉴权**：无需
- **用途**：创建新账号。用户名须为纯数字（学号）。
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `username` | string | 是 | 学号，纯数字，1–32 字符 |
  | `name` | string | 是 | 姓名，1–32 字符 |
  | `password` | string | 是 | 密码，8–16 位 |
  | `role` | string | 是 | 角色，如 `student`（受信任注册暂未限制） |
- **响应**：返回创建的用户对象（不含密码哈希）
- **常见错误**：`400` 参数非法；`409` 学号/工号已存在

### 2. 登录
- **接口**：`POST /api/v1/auth/login`
- **鉴权**：无需
- **用途**：校验账号密码并签发 JWT。
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `username` | string | 是 | 学号 |
  | `password` | string | 是 | 密码 |
- **响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": { "id": 1, "username": "20240001", "name": "张三", "role": "student" }
  }
}
```
- **常见错误**：`401 账号或密码错误`

### 3. 获取个人资料
- **接口**：`GET /api/v1/auth/profile`
- **鉴权**：需要
- **用途**：查看本人信息与本人发布的帖子列表。
- **响应**：`{ "user": {...}, "posts": [...] }`

### 4. 更新个人资料
- **接口**：`PATCH /api/v1/auth/profile`
- **鉴权**：需要
- **用途**：修改本人的姓名或用户名。字段为空则不改。
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `name` | string | 否 | 非空则更新姓名 |
  | `username` | string | 否 | 非空则更新用户名 |
- **响应**：返回 `{ "user": {...}, "posts": [...] }`

### 5. 修改密码
- **接口**：`PATCH /api/v1/auth/password`
- **鉴权**：需要
- **用途**：修改本人密码，需校验原密码，新密码以 bcrypt 哈希存储。
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `old_password` | string | 是 | 原密码 |
  | `new_password` | string | 是 | 新密码，8–16 位 |
  | `confirm_password` | string | 是 | 须与 `new_password` 一致 |
- **响应**：`{ "code":0, "msg":"success", "data":null }`
- **常见错误**：`400 原密码错误`；`400` 新密码长度非法或两次不一致

---

## 二、帖子模块（/api/v1/posts）

### 1. 发布帖子
- **接口**：`POST /api/v1/posts`
- **鉴权**：需要
- **用途**：发布失物（`lost`）或寻物（`found`）帖子。**学生发帖为 `pending`（待审核）；管理员发帖直接 `approved`。**
- **请求方式**：`multipart/form-data`（含图片上传）
- **表单字段**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `type` | string | 是 | `lost` 或 `found` |
  | `title` | string | 是 | 标题 |
  | `content` | string | 是 | 内容，1–2000 字符 |
  | `image` | file | 否 | 图片，保存至 `./uploads/posts/`，返回公网 URL |
- **响应**：返回帖子对象（含 `id, type, title, content, image_url, status, is_finished, user_id, created_at`）

### 2. 帖子列表（筛选 + 分页）
- **接口**：`GET /api/v1/posts`
- **鉴权**：可选（带 token 会识别身份，不带则匿名）
- **用途**：查询帖子列表。**普通用户仅见 `approved`；管理员可见全部并可按状态筛选。**
- **查询参数**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `type` | string[] | 否 | 可多选：`?type=lost&type=found` |
  | `status` | string[] | 否 | 可多选：`pending/approved/rejected`；仅管理员有效，普通用户强制 `approved` |
  | `page` | int | 否 | 页码，从 1 开始，默认 1 |
  | `page_size` | int | 否 | 每页数量，默认 20，上限 100 |
- **响应**：`{ list:[...], total, page, page_size }`

### 3. 帖子详情
- **接口**：`GET /api/v1/posts/:post_id`
- **鉴权**：可选
- **用途**：查看单个帖子详情。
- **路径参数**：`post_id`（uint64，必传，正整数）
- **可见性**：普通用户仅当帖子为 `approved` 时返回，否则 `404`；管理员返回全部
- **响应**：单个帖子对象

### 4. 删除帖子
- **接口**：`DELETE /api/v1/posts/:post_id`
- **鉴权**：需要
- **用途**：软删除帖子，**并级联软删除该帖的全部评论**（事务内先删评论，再删帖子）。
- **路径参数**：`post_id`（uint64，必传）
- **权限**：`student` 仅能删本人帖子；`postadmin`/`mainadmin` 可删任意
- **响应**：`{ "code":0, "msg":"success", "data":{ "post_id": 10 } }`

### 5. 恢复帖子
- **接口**：`PATCH /api/v1/posts/:post_id/recover`
- **鉴权**：需要
- **用途**：恢复被软删除的帖子（`deleted_at` 置回 NULL）。**不会自动恢复其评论。**
- **路径参数**：`post_id`（uint64，必传）
- **权限**：同删除
- **响应**：`{ "code":0, "msg":"success", "data":{ "post_id": 10 } }`

### 6. 审核帖子（仅管理员）
- **接口**：`PATCH /api/v1/posts/:post_id/review`
- **鉴权**：需要，角色为 `postadmin` 或 `mainadmin`
- **用途**：审核 `pending` 帖子，通过或驳回；**每帖仅可审核一次**。
- **路径参数**：`post_id`（uint64，必传）
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `status` | string | 是 | `approved` 或 `rejected` |
- **响应**：`{ "code":0, "msg":"success", "data":{ "post_id":10, "status":"approved" } }`
- **常见错误**：`400 无效的帖子状态`；`400 该帖子不可审核`（非 pending）

---

## 三、管理员专用（/api/v1/admin）

### 1. 修改帖子状态（仅管理员）
- **接口**：`PATCH /api/v1/admin/posts/:post_id/status`
- **鉴权**：需要，角色为 `postadmin` 或 `mainadmin`
- **用途**：直接修改帖子状态（可任意设置为 `pending/approved/rejected`）。
- **路径参数**：`post_id`（uint64，必传）
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `status` | string | 是 | `pending` / `approved` / `rejected` |
- **响应**：`{ "code":0, "msg":"success", "data":{ "post_id":10, "status":"pending" } }`

### 2. 已删除帖子列表（仅管理员）
- **接口**：`GET /api/v1/admin/posts/deleted`
- **鉴权**：需要，角色为 `postadmin` 或 `mainadmin`
- **用途**：查询已被软删除的帖子列表。
- **查询参数**：`page`（默认 1）、`page_size`（默认 20，上限 100）
- **响应**：`{ list:[...], total, page, page_size }`

---

## 四、评论模块

### 1. 发表评论
- **接口**：`POST /api/v1/comments`
- **鉴权**：需要
- **用途**：对指定帖子发表评论。**作者取自 token，请求体传 `post_id` 与 `content`。**
- **请求体（JSON）**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `post_id` | uint64 | 是 | 目标帖子 ID，须为正整数且帖子存在 |
  | `content` | string | 是 | 评论内容，1–1000 字符 |
- **响应**：返回创建的评论对象（`id, post_id, user_id, content, created_at` 等）

### 2. 评论列表（分页）
- **接口**：`GET /api/v1/posts/:post_id/comments`
- **鉴权**：无需（公开）
- **用途**：查询某帖评论，按 `id` 倒序（新评论在前）。
- **路径参数**：`post_id`（uint64，必传）
- **查询参数**：
  | 参数 | 类型 | 必传 | 说明 |
  |------|------|------|------|
  | `page` | int | 否 | 默认 1 |
  | `page_size` | int | 否 | 默认 20，上限 100 |
- **响应**：`{ list:[...], total, page, page_size }`

### 3. 删除评论
- **接口**：`DELETE /api/v1/comments/:comment_id`
- **鉴权**：需要
- **用途**：软删除评论。
- **路径参数**：`comment_id`（uint64，必传）
- **权限**：`student` 仅能删本人评论；`postadmin`/`mainadmin` 可删任意
- **响应**：`{ "code":0, "msg":"success", "data":null }`

---

## 认证与鉴权细节

- **令牌格式**：`Authorization: Bearer <access_token>`（`Bearer` 后必须有一个空格）
- **令牌来源**：登录接口返回的 `access_token`
- **有效期**：由 `config.yaml` 的 `jwt.expire_seconds` 决定
- **可选鉴权（OptionalAuth）**：帖子列表/详情即使不带 token 也可访问；普通用户仅见 `approved`，带管理员 token 可查看/筛选全部
- **权限判断位置**：集中在 service 层（`CheckpostPermission`、`CheckCommentPermission` 等），handler 只负责绑定参数与响应

---

## 文件上传说明

- 仅「发布帖子」支持图片上传（字段名 `image`）
- 保存路径：`./uploads/posts/<UnixNano>_<Unix><ext>`；公网 URL 形如 `http://localhost:8080/uploads/posts/...`
- 当前未限制文件大小与 MIME 类型，对外部署前建议增加白名单与大小限制

---

## 常见问题（FAQ）

1. **为什么我发的帖子别人看不到？** 学生发帖默认为 `pending`，需管理员审核通过（`approved`）后才会展示。
2. **如何审核帖子？** 用 `PATCH /api/v1/posts/:post_id/review` 传 `status: "approved"|"rejected"`，每帖仅可审核一次。
3. **误删了帖子怎么办？** 用 `PATCH /api/v1/posts/:post_id/recover` 恢复（不自动恢复评论）。
4. **如何删除评论？** 用 `DELETE /api/v1/comments/:comment_id`，普通用户只能删自己的评论。
5. **列表为何返回空？** 普通用户强制只看 `approved`，无匹配则为空数组。
6. **token 失效怎么办？** 重新登录获取新 `access_token` 并替换请求头。
7. **如何成为管理员？** 代码中注册角色校验 `role != "student"` 已被注释（临时放开）；生产环境请由后台创建管理员账号。

---

## 附录：数据库表结构（摘要）

- **users**：`id, username, name, password_hash, role, created_at, updated_at, deleted_at`
- **posts**：`id, type, title, content, image_url, is_finished, status(pending/approved/rejected), user_id, created_at, updated_at, deleted_at`；索引 `type`、`status`；外键 `user_id → users(id)`
- **comments**：`id, post_id, user_id, content, created_at, updated_at, deleted_at`；索引 `post_id`、`user_id`、`created_at DESC`；外键 `post_id → posts(id)`、`user_id → users(id)`（均 `ON DELETE CASCADE`）

---

**最后更新**：2026-09-18  
**本地路径**：`E:\study\LAF`
