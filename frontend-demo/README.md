# 定位功能示例与后端接口对接说明（frontend-geo-demo）

> 本文件夹**仅供前端同学参考**，是后端能力的配套示例，不属于后端编译产物。
> 技术栈按本项目前端约定：**Vue 3 全家桶（Vue 3 + Vue Router（可选）+ Pinia（推荐）+ axios）**。

本 README 分两部分：
- **第一部分**：定位功能（`pkg/geo`）的对接思路（核心）；
- **第二部分**：后端**目前已存在的全部接口**的对接思路，便于前端整体规划。

> **📌 本次更新（对话 / 完成寻找功能）**>> 后端新增了“申领 / 召领 → 对话 → 完成寻找”一条链路，前端需新增对接（帖子列表筛选也在本次调整）：>> | 类型 | 变化 | 对接要点 |> |---|---|---|> | 新增接口 | `POST /api/v1/posts/:post_id/conversations` | **申领 / 召领合一**：对 `found` 帖发起即“申领”、对 `lost` 帖发起即“召领”，后端按帖子类型自动判定，前端无需区分 |> | 新增接口 | `GET /api/v1/conversations` | “我的会话”列表（发起方 / 楼主均可见），分页 |> | 新增接口 | `GET /api/v1/conversations/:conversation_id/messages` | 会话消息列表（分页，仅参与方） |> | 新增接口 | `POST /api/v1/conversations/:conversation_id/messages` | 发送消息：body `{ content }` |> | 新增接口 | `POST /api/v1/conversations/:conversation_id/finish-requests` | 发起“完成寻找”申请（对话任一方） |> | 新增接口 | `PATCH /api/v1/conversations/:conversation_id/finish-requests/:request_id` | 另一方同意 / 拒绝：body `{ status: "agreed"｜"rejected" }`；**同意后帖子变为已完成** |> | 帖子列表 | `GET /api/v1/posts` 新增 `finished` 查询参数 | `finished=true` 只看已完成、`false` 只看未完成、不传为全部（**所有用户可用**，建议与 `type` 筛选项并列） |> | 排序变化 | 帖子列表 | **未完成优先，已完成沉到列表末尾**（后端已处理，前端一般无需改动） |> | 新增错误码 | `403 无权参与该对话`、`404 对话不存在`、`409 该帖子已完成` 等 | 见下文 F 节 |

> **📌 本次更新（帖子位置参数）**
>
> 发布帖子（`POST /posts`，multipart）**新增可选位置字段**，用于就近推荐：
>
> | 字段 | 类型 | 说明 |
> |---|---|---|
> | `location_id` | string | 手动选择的校园预设地点 ID（与坐标二选一） |
> | `latitude` | float | 自动定位纬度（WGS-84，与 `longitude` 成对） |
> | `longitude` | float | 自动定位经度（WGS-84） |
> | `supplement` | string | 地点补充说明，≤200 字符 |
>
> 后端会把地点解析为可读地名并**冗余存入帖子**，列表/详情直接返回 `location_id` / `location_name` / `supplement`，**前端无需再查一次地点接口**。可直接复用第一部分的 `LocationPicker.vue` 拿到 `location_id`，或拿 `latitude`/`longitude` 交给后端自动匹配。


---

## 〇、通用约定（所有接口都适用）

- **统一响应信封**：`{ "code": <业务码>, "msg": "<说明>", "data": <数据> }`。成功 `code=0`；`code != 0` 即业务失败，`msg` 可直接展示。
- **鉴权头**：需要登录的接口带 `Authorization: Bearer <access_token>`（注意 `Bearer` 后有且仅有一个空格）。
- **令牌来源**：登录接口返回的 `data.access_token`，建议存 Pinia + localStorage，并在 axios 请求拦截器里统一注入。
- **分页约定**：列表接口统一 `page`（从 1 开始，默认 1）+ `page_size`（默认 20，上限 100），返回 `{ list, total, page, page_size }`。
- **baseURL**：示例用 `/api/v1`（走前端反向代理）；如直连后端，改成 `http://<host>:8080/api/v1`。
- **错误处理**：axios 响应拦截器里判断 `code != 0` 抛错并弹提示；HTTP 401 统一跳登录页。

参考封装（Pinia + axios，可放在 `src/api/http.js`）：

```js
import axios from "axios";
const http = axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1" });
http.interceptors.request.use((cfg) => {
  const token = localStorage.getItem("access_token");
  if (token) cfg.headers.Authorization = `Bearer ${token}`;
  return cfg;
});
http.interceptors.response.use(
  (res) => {
    const body = res.data;
    if (body && body.code !== 0) return Promise.reject(body); // 业务错误
    return body.data; // 直接给到业务数据
  },
  (err) => Promise.reject(err.response?.data || err)
);
export default http;
```

---

# 第一部分：定位功能（校园地点 / 就近推荐）

## 1. 这个能力是什么

在校园失物招领场景里，用户发布/编辑失物帖子时需要标注地点，便于“就近推荐”。后端提供**校园预设地点 + 坐标匹配**能力（选型方案 E+D）：

- 前端拿坐标（浏览器 GPS）或让用户手动选地点；
- 后端用校园预设地点表把坐标翻译成可读地名（如“家和东苑3号楼”），并算距离；
- 用户选择后还可以补一句说明（如“图书馆东门台阶旁”）。

**为什么不用 IP 定位**：校园网出口 IP 通常只有一两个，全校用户会被解析到同一区域，无法区分具体楼栋，对“就近”没有意义。所以采用“客户端坐标 + 校园预设地点”。

> 地点已**细化到每一栋楼**：屏峰·家和东苑1-18号楼、家和西苑1-15号楼；朝晖·尚德园1-9号楼、梦溪村1-7号楼、综合楼；莫干山·德馨苑德一楼~德十楼。

## 2. 文件

```
frontend-geo-demo/
├── README.md                     # 本文件
└── src/
    ├── api/
    │   └── geo.js                # 定位两个接口的封装
    └── components/
        └── LocationPicker.vue    # 地点选择组件（自动定位 + 手动选择 + 补充说明）
```

## 3. 接口契约（均公开，无需登录）

### 3.1 获取校园预设地点列表
- `GET /api/v1/geo/locations`
- 响应 `data`：按校区分组的数组

```json
[
  { "campus": "屏峰校区", "locations": [
      { "id": "pf-jiahe-east-3", "campus": "屏峰校区", "name": "家和东苑3号楼",
        "category": "宿舍", "address": "屏峰·家和东苑", "latitude": 30.23254, "longitude": 120.0936 }
  ]}
]
```

### 3.2 定位 / 匹配最近地点
- `POST /api/v1/geo/locate`
- 请求体（`location_id` 与 `latitude`+`longitude` **二选一**）：

```json
{ "location_id": "pf-jiahe-east-3", "supplement": "3号楼门口快递柜旁" }
```

```json
{ "latitude": 30.2338, "longitude": 120.0930, "supplement": "图书馆东门台阶旁" }
```

- 响应 `data`：`{ location, distance_meters, match_type("auto"|"manual"), supplement }`
- 错误：`400 参数校验失败`；`400 无效的坐标`；`404 定位地点不存在`

## 4. 推荐流程

```
点击“一键自动定位”
   ├─ 授权成功 ─► 经纬度 ─► POST /geo/locate ─► 显示“家和东苑3号楼（约 12 米）”
   └─ 拒绝/不支持/失败 ─► 降级：GET /geo/locations 手动选择
                             └─ 可选补充说明 ─► 再次 POST /geo/locate 确认
```

最终把结果（`location` + `supplement`）连同帖子一起提交给后端保存。

## 5. 前端接入要点

- **Pinia**：建一个 `useLocationStore` 存当前定位结果，发布页/编辑页共用，避免重复定位。
- **组件**：`LocationPicker.vue` 通过 `emit("confirm", payload)` 把结果交给父组件，可直接放进发布表单。
- **axios**：复用项目统一的实例（拦截器/鉴权头），不要另起一个。
- **无 axios 时**用 `fetch`：
  ```js
  export async function fetchCampusLocations() {
    const res = await fetch("/api/v1/geo/locations");
    return (await res.json()).data;
  }
  ```

## 6. 注意事项

1. **定位需 HTTPS 或 localhost**：`navigator.geolocation` 在非安全上下文被禁用，http 域名下只能走手动选择。
2. **坐标系**：浏览器返回 WGS-84，后端预设也是 WGS-84；接入高德等(GCJ-02)需先纠偏。
3. **一定要有手动兜底**：授权可能被拒或室内不准，手动列表是必需项。
4. **预设地点需校准**：坐标为近似值，上线前按官方校园地图校准；新增地点只改后端 `pkg/geo/campus.go`。
5. **补充说明**是自由文本，前端应限长（≤200 字符）并做必要校验/转义。

---

# 第二部分：后端现有接口对接思路汇总

> 以下按模块列出**当前后端已存在的全部接口**及前端对接建议。鉴权列：`公开`=无需登录；`可选`=带 token 识别身份，不带也能访问；`需登录`=必须 token；角色列标注管理员限制。

## A. 认证与用户（`/api/v1/auth`）

| 接口 | 方法 | 鉴权 | 对接思路 |
|---|---|---|---|
| `/auth/register` | POST | 公开 | 注册表单：`username`(纯数字学号)/`name`/`password`(8-16)/`role`。成功后可引导去登录 |
| `/auth/login` | POST | 公开 | 登录表单：`username`/`password`；成功后存 `data.access_token`、`data.user` |
| `/auth/profile` | GET | 需登录 | “我的”页面：返回本人信息 + 本人帖子列表 `{user, posts}` |
| `/auth/profile` | PATCH | 需登录 | 编辑资料：`name`/`username` 非空才更新 |
| `/auth/password` | PATCH | 需登录 | 改密：`old_password`/`new_password`/`confirm_password` |
| `/auth/account` | DELETE | 需登录 | 注销本人账号（软删除本人及本人帖子/评论）；建议二次确认弹窗 |

## B. 帖子（`/api/v1/posts`）

| 接口 | 方法 | 鉴权 | 对接思路 |
|---|---|---|---|
| `/posts` | GET | 可选 | 首页/列表：`type`(`lost`/`found` 可多选)、`status`(仅管理员)、`finished`(`true`已完成/`false`未完成/不传全部，所有用户可用)、`page`/`page_size`；返回分页信封。**未完成帖子排在前面** |
| `/posts` | POST | 需登录 | 发布：**multipart/form-data**（`type`/`title`/`content`/可选 `image`/可选位置 `location_id` 或 `latitude`+`longitude` + `supplement`）。学生发帖为 `pending`，管理员直接 `approved` |
| `/posts/:post_id` | GET | 可选 | 详情页；普通用户仅能看 `approved`，否则 404 |
| `/posts/:post_id` | DELETE | 需登录 | 删除本人帖子（管理员可删任意）；会级联软删该帖评论 |
| `/posts/:post_id/recover` | PATCH | 需登录 | 恢复被删帖子 |
| `/posts/:post_id/review` | PATCH | `postadmin`/`mainadmin` | 审核：`{status: "approved"\|"rejected"}`，每帖一次 |
| `/posts/:post_id/comments` | GET | 公开 | 帖子详情页的评论列表（读操作挂在帖子下） |

> 发布帖子是 **multipart**，axios 用 `FormData` 提交，**不要手动设 `Content-Type`**（浏览器会自动带 boundary）。

```js
const fd = new FormData();
fd.append("type", "lost");
fd.append("title", "遗失一张校园卡");
fd.append("content", "……");
if (imageFile) fd.append("image", imageFile);
// 可选位置：手动选的 location_id，或自动定位的 latitude/longitude（二选一）
if (locationId) fd.append("location_id", locationId);
if (coords) { fd.append("latitude", coords.latitude); fd.append("longitude", coords.longitude); }
if (supplement) fd.append("supplement", supplement);
await http.post("/posts", fd);
```

## C. 评论（`/api/v1/comments`）

| 接口 | 方法 | 鉴权 | 对接思路 |
|---|---|---|---|
| `/comments` | POST | 需登录 | 发表评论：**body 里带 `post_id`**（评论已迁到评论分组） |
| `/comments/:comment_id` | DELETE | 需登录 | 删除本人评论（管理员可删任意） |

## D. 申诉（`/api/v1/appeals`）

| 接口 | 方法 | 鉴权 | 对接思路 |
|---|---|---|---|
| `/appeals` | POST | 公开 | 注销/封禁后无法登录，故申诉公开：`username`(被注销学号) + `reason`(`self_regret`自行注销反悔/`wrongful_ban`被误封请求撤回/`other`其他) + `content`(reason=other 必填)。适合做“申诉页” |

## E. 地理位置（`/api/v1/geo`）

| 接口 | 方法 | 鉴权 | 对接思路 |
|---|---|---|---|
| `/geo/locations` | GET | 公开 | 拉取校园预设地点（按校区分组），渲染手动选择器 |
| `/geo/locate` | POST | 公开 | 定位/匹配最近地点，见第一部分 |

## F. 管理员（`/api/v1/admin`）

| 接口 | 方法 | 角色 | 对接思路 |
|---|---|---|---|
| `/admin/posts/:post_id/status` | PATCH | `postadmin`/`mainadmin` | 修改帖子状态 |
| `/admin/posts/deleted` | GET | `postadmin`/`mainadmin` | 已删帖子列表（可配合恢复） |
| `/admin/users/:user_id` | DELETE | `mainadmin` | 注销指定用户（同批软删其帖子/评论） |
| `/admin/users/:user_id/recover` | PATCH | `mainadmin` | 恢复用户（仅恢复同批删除内容） |
| `/admin/appeals/:appeal_id/review` | PATCH | `mainadmin` | 审核申诉：`{status:"approved"\|"rejected"}`；通过则自动恢复账号 |
| `/admin/reviews` | GET | `mainadmin` | 待审批列表：`type`(`post`/`appeal`，不传返回全部)+`page`/`page_size` |

> **管理端建议**：单独的路由区（如 `/admin/*`），进页面前校验 `user.role`；401 跳登录、403 提示“仅管理员可操作”。

---

## G. 对话 / 完成寻找（`/api/v1/conversations`）> 失主与拾得者通过“申领 / 召领”建立一对一会话协商归还；任一方可发起“完成寻找申请”，另一方同意后帖子置为已完成。> **申领与召领合一**：对 `found` 帖即申领、对 `lost` 帖即召领，同一个接口，前端无需区分。| 接口 | 方法 | 鉴权 | 对接思路 ||---|---|---|---|| `/posts/:post_id/conversations` | POST | 需登录 | **申领/召领**：在帖子详情页放“我要申领 / 我捡到了，联系失主”按钮 → 发起对话 → 跳转到会话页。不能对自己帖子发起；已完成帖子会报 `409` || `/conversations` | GET | 需登录 | “我的消息/会话”列表页（发起方与楼主均可见），`page`/`page_size` || `/conversations/:conversation_id/messages` | GET | 需登录 | 会话详情页拉取消息（分页，倒序=新在前；展示时一般需反转为旧在前） || `/conversations/:conversation_id/messages` | POST | 需登录 | 发送消息：`{ content }`（1–1000 字） || `/conversations/:conversation_id/finish-requests` | POST | 需登录 | 发起“完成寻找”申请；同会话已有待处理申请时会报 `400` || `/conversations/:conversation_id/finish-requests/:request_id` | PATCH | 需登录 | 另一方处理：`{ status: "agreed"｜"rejected" }`；**agreed 后对应帖子 `is_finished=true`** |**对接建议**：- 会话可用轮询（如进入会话页后每 5–10s 拉一次消息）即可，无需 Websocket；- “完成寻找”做成会话内的系统提示条 +“同意/拒绝”按钮，仅在**对方发起**时显示操作按钮；- 帖子详情“已完成”状态（`is_finished=true`）时，隐藏申领/召领入口并置灰提示；- 常见错误：`403 无权参与该对话`、`404 对话不存在`、`409 该帖子已完成`、`400 该完成申请不可处理`。## 七、建议的前端整体结构（Vue3）

```
src/
├── api/            # 每模块一个文件：auth.js / post.js / comment.js / appeal.js / geo.js / admin.js / conversation.js
├── stores/         # Pinia：useAuthStore(令牌/用户) / useLocationStore(定位) / useConversationStore(可选，未读/会话缓存)
├── router/         # 路由 + 登录/角色守卫（meta.requiresAuth / meta.roles）
├── components/     # LocationPicker.vue 等通用组件
└── views/          # 页面：登录/注册、帖子列表/详情/发布、我的、申诉、管理端
```

- **鉴权守卫**：`router.beforeEach` 读 Pinia 的 token，未登录跳登录；`meta.roles` 不含当前角色则跳 403。
- **令牌失效**：axios 响应拦截器捕获 401 → 清 token → 跳登录。
- **就近推荐**：列表页可选把用户 `location` 传后端（后续可加坐标筛选），或前端对已取到的帖子按距离排序。

## 八、注意事项

1. **定位需 HTTPS 或 localhost**（同第一部分第 6 节）。
2. **多选参数**：帖子列表的 `type`/`status` 是数组，query 用重复键（`?type=lost&type=found`）；`finished` 是单值（`true`/`false`）。
   - 帖子列表**已完成帖子会沉到末尾**，若做“未完成优先”展示，前端一般无需额外排序。
   - **申领/召领**共用 `POST /posts/:post_id/conversations`，按帖子类型自动判定语义。
   - 会话消息与完成申请仅**对话参与方**可读写，非参与方会收到 `403`。
3. **multipart 上传**：不要手动设 `Content-Type`。
4. **角色差异**：学生只能看 `approved`、只能改本人资源；管理员接口 403 需前端提前拦截。
5. **错误信封**：所有接口都判 `code`，不要只看 HTTP 状态。
