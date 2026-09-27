# agent.md — LAF 项目 AI 协作指南

> 本文件面向后续在本仓库工作的 AI / 协作者，目标是让任何接手者**不改错约定、不踩已知坑**。
> 内容以当前真实源码（`E:\study\LAF`，即沙箱内 `/mnt/local/LAF`）为准。
> **最后更新：2026-09-27**（上一版为 2026-09-19，已补齐申诉 / 对话 / 定位 / 超级管理员等新模块，并同步图片相对路径、作者姓名回填等改动）。

---

## 0. 先读这几条（最高优先级）

1. **全仓 CRLF 换行**：新增/修改文件请保持 CRLF，否则整文件 diff。
2. **沙箱可编译**：优先在**副本目录**装 Go 工具链并 `go build ./...`（见第 10 节）；装不上再用 `gofmt -e` 做语法校验。**`gofmt -l` 全仓报错属既有现象**，别据此判断代码损坏。
3. **改既有函数签名后，必须全仓核对所有调用点**（历史上改过 `router.New` 的入参）。
4. **每个 `AbortWithException` / `AbortWithError` 之后必须紧跟 `return`**（本项目曾因此出现权限绕过）。
5. 涉及多文件/多接口的改动，**先出方案，用户确认后再写入**；琐碎小改可直接动手。

---

## 1. 项目概览

- **项目名**：LAF（校园失物招领后端）
- **模块名**：`LAF`（见 `go.mod`）
- **技术栈**：Go 1.26.5 + Gin v1.12.0 + GORM v1.31.2（`gorm.io/driver/mysql` v1.6.0）+ Viper v1.21.0 + `golang-jwt/jwt/v5` v5.3.1 + bcrypt（`golang.org/x/crypto` v0.57.0）。**不要擅自引入新的第三方依赖**。
- **入口**：`main.go` → `config.Load()` → 若 `cfg.Database.Enabled` 则 `database.ConnectSQL(cfg.Database)` → `router.New(db, cfg.JWTConfig)` → `engine.Run(":port")`
- **默认端口**：8080（`server.port`）
- **统一响应**：`{ "code": <int>, "msg": <string>, "data": <any> }`，成功为 `code=0, msg="success"`

---

## 2. 分层架构（务必遵循）

```text
main.go                         # 入口：config.Load -> database.ConnectSQL -> router.New -> Run
config/config.go + config.yaml  # Viper 读取（固定路径 config/config.yaml）+ 结构体 mapstructure 标签
internal/
  router/router.go              # 唯一装配点：repository -> service -> handler 构造注入 + 路由注册
  handler/<领域>/<动作>.go       # 一个接口一个文件，工厂函数返回 gin.HandlerFunc
  service/<领域>.go              # 业务校验/权限，XxxInput / XxxResult DTO
  repository/<领域>.go           # GORM 数据访问，哨兵错误 + 软删除 + 分页
  middleware/                   # auth.go（JWT: Auth/OptionalAuth/RequireRole）、error.go（ErrorHandler）
  model/model.go                # GORM 模型，中文列注释，软删除字段，状态枚举常量
  database/database.go          # 全局 db + RWMutex + 每 5 秒 Ping 断线自动重连
pkg/
  apperror/                     # Error{Code,Msg} + 预定义错误码 + Abort/Handle 助手
  response/                     # {code,msg,data} 信封
  pagination/                   # page/page_size 解析（默认 20，上限 100）
  geo/                          # 校园预设地点与坐标匹配（无外部依赖）
migrations/tables.sql           # 手写 DDL（DROP IF EXISTS + CREATE，CHECK/外键/索引）
frontend-demo/                  # 前端示例（Vue3 + Vite + Pinia + axios）；详见第 12 节
```

**职责红线**
- **handler 只做 HTTP**：参数绑定、取当前用户、调 service、写响应；**不写业务判断、不碰 GORM**。
- **service 只做业务**：规则校验、状态流转、权限判断（如 `CheckpostPermission`、`CheckCommentPermission`），把仓库领域错误翻译成 `apperror`。
- **repository 只做数据访问**：把 GORM/MySQL 错误翻译成领域哨兵错误（`ErrXxxNotFound` 等）后向上抛。
- 参数绑定失败 → 立刻返回 400，不要继续。

---

## 3. 目录速查（当前真实文件）

- 用户：`handler/user/{register,login,profile,update_profile,update_password,deactivate}.go` + `service/user.go` + `repository/user.go`
- 帖子：`handler/post/{create,list,detail,delete,recover_post}.go` + `service/post.go` + `repository/post.go`
- 帖子管理：`handler/postadmin/{review_post,update_status,list_deleted}.go` + `service/postadmin.go` + `repository/postadmin.go`
- 评论：`handler/comment/{create,list,delete}.go` + `service/comment.go` + `repository/comment.go`
- 申诉：`handler/appeal/create.go` + `service/appeal.go` + `repository/appeal.go`
- 对话/完成寻找：`handler/conversation/{start,list,messages,send,finish,get_pending_finish,review_finish,withdraw_finish}.go` + `service/conversation.go` + `repository/{conversation,message,finish_request}.go`
- 定位：`handler/geo/{locations,locate}.go` + `service/geo.go` + `pkg/geo/`
- 超级管理员：`handler/mainadmin/{delete_user,recover_user,review_appeal,list_reviews}.go` + `service/mainadmin.go` + `repository/mainadmin.go`
- 作者姓名回填：`repository/authorname.go`（跨 Post/Comment 复用）
- 分页结果结构：`service/pageresult.go`（泛型 `PageResult[T]`）
- 路由：`internal/router/router.go`

---

## 4. 接口与路由约定（以 `internal/router/router.go` 为准）

**前缀**：所有业务接口挂在 `/api/v1` 之下。路径参数一律用 `:post_id` / `:comment_id` / `:user_id` / `:appeal_id` / `:conversation_id` / `:request_id`，handler 内 `strconv.ParseUint(..., 10, 64)` 并校验 `> 0`，非法返回 400。

### 认证/用户 `/api/v1/auth`
| 方法 | 路径 | 中间件 | handler |
|------|------|--------|---------|
| POST | `/register` | 公开 | `user.Register` |
| POST | `/login` | 公开 | `user.Login` |
| GET | `/profile` | Auth | `user.GetProfile` |
| PATCH | `/profile` | Auth | `user.UpdateProfile` |
| PATCH | `/password` | Auth | `user.UpdatePassword` |
| DELETE | `/account` | Auth | `user.DeactivateAccount`（注销本人账号，软删本人及本人内容） |

### 帖子 `/api/v1/posts`
| 方法 | 路径 | 中间件 | handler |
|------|------|--------|---------|
| GET | `` | OptionalAuth | `post.ListPosts` |
| POST | `` | Auth | `post.Create`（multipart/form-data） |
| GET | `/:post_id` | OptionalAuth | `post.GetPost` |
| DELETE | `/:post_id` | Auth | `post.DeletePost` |
| PATCH | `/:post_id/recover` | Auth | `post.RecoverPost` |
| PATCH | `/:post_id/review` | Auth + RequireRole[postadmin,mainadmin] | `postadmin.ReviewPost` |
| POST | `/:post_id/conversations` | Auth | `conversation.Start`（申领/召领：按 type 自动判定） |
| GET | `/:post_id/comments` | 公开 | `comment.List` |

### 评论 `/api/v1/comments`
| 方法 | 路径 | 中间件 | handler |
|------|------|--------|---------|
| POST | `` | Auth | `comment.Create` |
| DELETE | `/:comment_id` | Auth | `comment.Delete` |

### 申诉 `/api/v1/appeals`
| 方法 | 路径 | 中间件 | handler |
|------|------|--------|---------|
| POST | `` | 公开 | `appeal.Create`（账号被注销后无法登录，故公开提交） |

### 定位 `/api/v1/geo`（公开）
| 方法 | 路径 | handler |
|------|------|---------|
| GET | `/locations` | `geo.ListLocations` |
| POST | `/locate` | `geo.Locate` |

### 对话 `/api/v1/conversations`（全部需 Auth）
| 方法 | 路径 | handler |
|------|------|---------|
| GET | `` | `conversation.List` |
| GET | `/:conversation_id/messages` | `conversation.ListMessages` |
| POST | `/:conversation_id/messages` | `conversation.SendMessage` |
| POST | `/:conversation_id/finish-requests` | `conversation.Finish` |
| PATCH | `/:conversation_id/finish-requests/:request_id` | `conversation.ReviewFinish` |
| GET | `/:conversation_id/finish-requests` | `conversation.GetPendingFinish` |
| DELETE | `/:conversation_id/finish-requests/:request_id` | `conversation.WithdrawFinish` |

### 管理员 `/api/v1/admin`
| 方法 | 路径 | 角色 | handler |
|------|------|------|---------|
| PATCH | `/posts/:post_id/status` | postadmin, mainadmin | `postadmin.UpdatePostStatus` |
| GET | `/posts/deleted` | postadmin, mainadmin | `postadmin.ListDeletedPosts` |
| DELETE | `/users/:user_id` | mainadmin | `mainadmin.DeleteUser` |
| PATCH | `/users/:user_id/recover` | mainadmin | `mainadmin.RecoverUser` |
| PATCH | `/appeals/:appeal_id/review` | mainadmin | `mainadmin.ReviewAppeal` |
| GET | `/reviews` | mainadmin | `mainadmin.ListReviews`（`?type=post|appeal` 过滤） |

**约定**
- 列表接口统一走 `pkg/pagination`：`page`（默认 1）+ `page_size`（默认 20，上限 100），**不对外暴露 limit/offset**；返回信封 `{ list, total, page, page_size }`（见 `service.PageResult[T]`）。
- 列表重复参数（如帖子 type）用 `c.QueryArray("type")`；前端 axios 已自定义 `paramsSerializer` 序列化成 `type=lost&type=found`。
- 列表/详情常用 `middleware.OptionalAuth`：带 token 时注入身份，不带则匿名。
- 帖子列表排序：`is_finished asc, id desc`（未完成优先，新帖在前）。

---

## 5. 鉴权与权限模型

**Token 传递**：请求头 `Authorization: Bearer <access_token>`（`Bearer` 后有**一个空格**）。

**JWT claims**：`{ user_id, role, RegisteredClaims }`；HS256 签名；密钥/有效期/签发者来自 `config.jwt`（`secret` / `expires` / `issuer`）。校验启用 `WithValidMethods([HS256])` + `WithExpirationRequired()`。校验还要求 `role != ""` 且 `user_id != 0`。

**角色**
| 角色 | 权限 |
|------|------|
| `student` | 只能操作本人帖子/评论；发帖为 `pending`；列表/详情仅见 `approved`（`postadmin`/`mainadmin` 可见全部） |
| `postadmin` | 管理任意帖子/评论；发帖直接 `approved`；可用帖子管理接口 |
| `mainadmin` | 拥有 `postadmin` 全部权限（白名单并列），另有注销/恢复用户、审核申诉、汇总待审批 |

**中间件（`internal/middleware/auth.go`）**
- `Auth(jwtConfig)`：强制登录，失败 401。
- `OptionalAuth(jwtConfig)`：可选登录，无效也放行。
- `RequireRole([]string{...})`：角色白名单，常与 `Auth` 组合；未登录或角色不符返回 403（`AdminForbiddenError`）。
- `ErrorHandler()`：全局兜底，把 `c.Errors.Last()` 交给 `apperror.Handle`。
- 读取上下文：`CurrentUserID(c)` / `CurrentRole(c)`，Context 键为 `auth_user_id` / `auth_role`。

> 权限判断集中在 service 层；handler 只负责编排。

---

## 6. 错误处理约定

- 统一结构体与预定义错误在 `pkg/apperror`（`Error{Code,Msg}` + `NewError` + `AbortWithException` / `AbortWithError` / `Handle`）。
- 关键码：`0` 成功、`400` 参数/状态非法/原密码错误、`401` 未登录或账号密码错误、`403` 越权、`404` 资源不存在、`409` 用户名已存在/状态冲突、`1000` 系统异常、`1001` 数据库操作失败。
- 抛错用 `apperror` 常量构造，**不要**在 handler 里手写 `c.JSON(500,...)`。
- `response.Error` 会把业务码直接当 HTTP 状态码；非 `100~599` 时回退 500。

---

## 7. 数据模型与软删除

**表 / 模型（`internal/model/model.go`，须与 `migrations/tables.sql` 同步）**：`users`、`posts`、`comments`、`appeals`、`conversations`、`messages`、`finish_requests`。均含 `created_at, updated_at, deleted_at`（GORM 软删除）。

**状态枚举常量**
- 帖子审核：`pending` / `approved` / `rejected`；帖子类型：`lost` / `found`
- 申诉原因：`self_regret` / `wrongful_ban` / `other`；申诉状态：`pending` / `approved` / `rejected`
- 完成申请状态：`pending` / `agreed` / `rejected`（同意后对应帖子 `is_finished=true`）

**关键规则**
- **软删除不触发外键级联**：删父表（帖子）时必须在**同一事务内先软删子表（评论）再软删帖子**（见 `PostRepository.DeletePost`）。DDL 的 `ON DELETE CASCADE` 只对物理删除生效。
- **恢复不级联**：恢复帖子只把帖子 `deleted_at` 置 NULL，不自动恢复评论；恢复用户时只恢复“与其同批删除”的帖子与评论。
- 查含已删记录用带 `Unscoped()` 的方法；恢复用 `Unscoped()` 把 `deleted_at` 置 NULL。
- **`model.go` 与 `migrations/tables.sql` 必须同步修改**。
- **作者姓名回填**：`Post`/`Comment` 的 `AuthorName` 标注 `gorm:"-" json:"author_name"`（不入库），由 `repository/authorname.go` 按 `user_id` 批量关联 `users` 后回填；查不到作者（如已软删）兜底为 `"未知用户"`。查询帖子的仓库方法（`GetPostByID` / `GetPosts` / `GetDeletedPosts` 等）末尾都会调用回填。

---

## 8. 本次会话确认的坑（重要）

1. **路由前缀是 `/api/v1/auth`，不是 `/api/v1/users`**：写文档/前端联调勿凭直觉写错。
2. **注册（`service/user.go`）当前放开了角色限制**：`role != "student"` 的校验被注释（临时），`role` 为必填字段。生产环境请改为后台创建管理员或恢复校验。
3. **登录响应字段**：`{ access_token, token_type:"Bearer", expires_in, user }`；`expires_in` 来自 `config.jwt.expires`。
4. **更新资料响应**：返回 `{ user, posts }`（含该用户帖子），不是裸 user。
5. **发帖是 `multipart/form-data`**（因含可选图片字段名 `image`），其余多为 JSON。
6. **图片只存相对路径**：`internal/handler/post/create.go` 存成 `/uploads/posts/xxx.png`，**不再写死主机**（历史遗留的 `server.public_base_url` / `buildPublicImageURL` 已移除）。同源访问由后端 `engine.Static("/uploads", "./uploads")` 与前端 Nginx `/uploads/` 反代提供。前端 `src/utils/image.js` 的 `resolveImageUrl` 会把历史绝对地址归一成 `/uploads/...`。
7. **图片上传仍无大小/MIME 限制**：保存于 `./uploads/posts/`；上生产前建议补白名单与大小限制。
8. **`gofmt -l` 全仓报错属既有现象**（CRLF / 既有格式），不要据此判断代码损坏。
9. **热点函数签名（改前必查调用点）**：`router.New(db, jwtConfig)`、`repository.FillXxx`、`service.PageResult[T]` 的使用。

---

## 9. 配置说明（config/）

`config/config.go` 用 Viper 读取**固定路径 `config/config.yaml`**，支持 `WatchConfig` 热更新，但：

> ⚠️ `jwt` / `database` 在启动时已注入各服务，**热更新不生效，需重启**（代码里有对应提示打印）。

结构体字段：
- `server`: `port`(int)
- `database`: `enabled`(bool) / `host` / `port` / `username` / `password` / `name`
- `jwt`: `secret` / `expires`(int64, 秒) / `issuer`

文件约定：
- `config/config.yaml` **已被 `.gitignore` 忽略**（含数据库密码/JWT secret），新环境从 `config/config.example.yaml` 复制。
- `config/config.docker.yaml` **同样被忽略**（容器用，`database.host=mysql`），由 compose 挂载为容器内 `/app/config/config.yaml`。

`internal/database`：连接池 `MaxIdle=10 / MaxOpen=100 / ConnMaxLifetime=1h`，并有每 5 秒的断线自动重连 goroutine。

---

## 10. 构建与校验

- **优先编译验证**：在**副本目录**（勿在用户挂载目录直接构建）装与 `go.mod` 一致的 Go 版本（可用 `mirrors.aliyun.com/golang/`），用 `GOPROXY=https://goproxy.cn,direct` 拉依赖后 `go build ./...`，再 `go vet ./...`。
- 装不上工具链时，用 **`gofmt -e`** 做语法解析校验 + 跨层引用/调用点审计，并在交付说明中注明“请在本地运行 `go build ./...` 做最终编译确认”。
- **源码统一 CRLF**：新建/修改文件保持 CRLF。

---

## 11. 新增一个接口的标准流程（checklist）

1. `internal/model`：如需新字段，改 model **并同步** `migrations/tables.sql`。
2. `internal/repository`：加数据访问方法（哨兵错误翻译 + 软删除 + 分页；需要作者姓名则调用回填）。
3. `internal/service`：写业务逻辑 + 权限判断（错误用 `apperror` 常量，列表返回 `PageResult[T]`）。
4. `internal/handler/<领域>/<动作>.go`：绑定参数（校验必填/类型）→ 取当前用户 → 调 service → `response.Success`。
5. `internal/router/router.go`：在 `New` 里装配依赖并注册路由，按需挂 `Auth`/`OptionalAuth`/`RequireRole`。
6. 校验：`go build ./...`（副本）或 `gofmt -e`；保持 CRLF。
7. 同步更新 `README.md` 的接口清单与参数表、以及本文件的目录/路由速查。

---

## 12. 前端示例（frontend-demo/）

- 技术栈：Vue3 + Vue Router + Pinia + axios + Vite；`axios` baseURL 取 `VITE_API_BASE_URL`（默认 `/api/v1`）。
- 统一响应拦截在 `src/api/http.js`：`code===0` 取 `data`，否则抛错；请求拦截自动注入 `Authorization: Bearer <token>`。
- 图片地址归一：`src/utils/image.js` 的 `resolveImageUrl`。
- 视图：`views/` 下含帖子列表/详情/发布、登录/注册、个人资料、申诉、对话/聊天、404，以及 `views/admin/`（Reviews/Appeals/Users/DeletedPosts）。
- 构建：`npm install && npm run build`；容器内由 Nginx 托管（`frontend-demo/nginx.conf` 做 SPA 回退 + `/api`、`/uploads` 同源反代到 `backend:8080`）。

---

## 13. 部署（在线更新）

**形态**：前后端分端口部署。后端 `8080` 直接对外提供 API 与 `/uploads` 图片；临时前端 `9090`（容器内 Nginx 同源反代后端）；正式前端由前端同学部署到 `5173`，跨域调用后端（来源须加入 `config/config.docker.yaml` 的 `server.cors_allow_origins`）。MySQL 仅容器内网可达。三服务编排见 `docker-compose.yml`（mysql + backend + frontend）。

**相关文件**：`Dockerfile`（后端多阶段构建）、`frontend-demo/Dockerfile` + `nginx.conf`、`config/config.docker.yaml`、`docker-compose.yml`、`deploy.sh`（服务器一键部署）、`migrations/tables.sql`（首次启动自动建表）。

**服务器更新流程**（详见 `DEPLOY.md`）：
1. 服务器拉取/覆盖最新代码。
2. **一次性修正历史图片数据**（把写死主机的绝对地址改回相对路径）：
   ```bash
   docker exec -i laf-mysql mysql -uroot -proot123456 laf_db -e \
   "UPDATE posts SET image_url=CONCAT('/uploads/',SUBSTRING_INDEX(image_url,'/uploads/',-1)) WHERE image_url LIKE '%/uploads/%';"
   ```
3. 重建启动：`docker compose up -d --build`（数据卷与 uploads 卷保留，不丢数据）。
4. 验证：`http://<公网IP>:8080/api/v1/geo/locations` 返回 200；`docker compose logs -f backend`。

**注意**：定位功能（`navigator.geolocation`）需 HTTPS；经 `http://<IP>` 访问时浏览器禁用定位，仅能手动选点。首次启动数据卷为空时自动执行 `migrations/tables.sql`。

---

## 14. 常用约定速记

- 响应：`{code,msg,data}`；成功 `code=0, msg="success"`。
- 分页：`page`(≥1) / `page_size`(1–100)，返回 `{list,total,page,page_size}`。
- 鉴权：`Authorization: Bearer <token>`（HS256，必带过期）。
- ID 解析：`ParseUint(...,10,64)` 且校验 `>0`。
- 权限：student 仅本人；postadmin 管理帖子/评论；mainadmin 另有账号/申诉/审批。
- 软删级联：事务内先子后父；恢复不级联。
- 图片：只存 `/uploads/...` 相对路径。
- 作者姓名：`AuthorName`（`gorm:"-"`）由 `repository/authorname.go` 回填，缺失兜底 `"未知用户"`。
- **Abort 后必须 return**。

---

**生成时间**：2026-09-27
**依据**：本会话对 `/mnt/local/LAF`（即 `E:\study\LAF`）源码的通读，重点核对 `main.go`、`config/config.go`、`internal/router/router.go`、`internal/model/model.go`、`internal/middleware/auth.go`、各 `repository`/`service`/`handler`、`pkg/*`、`migrations/tables.sql`、`docker-compose.yml` 及部署文档。
