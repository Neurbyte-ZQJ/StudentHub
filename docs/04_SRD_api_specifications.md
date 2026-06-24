# 学生"一站式"自主管理过程管理系统 · 系统需求规范（SRD）/ RESTful API 接口文档

| 文档版本 | 修订日期   | 编写者                 | 文档状态 |
| -------- | ---------- | ---------------------- | -------- |
| V1.0     | 2026-06-14 | 前后端全栈接口专家     | 评审稿   |

> **配套文档**：[01_PRD.md](./01_PRD.md) · [02_ADR.md](./02_ADR.md) · [03_database_design_spec.md](./03_database_design_spec.md)
>
> **协议**：HTTPS + JSON（RFC 8259）
> **风格**：RESTful（资源 + HTTP 动词），辅以"动作 endpoint"用于状态机推进
> **基础地址**：`https://{domain}/api/v1`
> **后端**：Go 1.22 / Gin / GORM v2 · **前端**：Vue 3 + Pinia + Element-Plus / uni-app
> **认证**：JWT Bearer + 5 级角色 RBAC（详见 §2.2）

---

## 0. 阅读指引

- 章节 1-4：通读，理解通用约定（路径、鉴权、分页、错误、状态机）。
- 章节 5：基础层（认证、用户、组织、字典、文件、通知、审计）。
- 章节 6-9：业务模块（TY 团员发展 / ST 社团活动 / SQ 学生社区 / QG 勤工助学）。
- 章节 10：CMP 综合素质量化、IDX 学生画像。
- 章节 11：状态机/事件/Webhook/SSE 实时推送。
- 章节 12：OpenAPI 3.0 骨架与 Mock 指南。
- 章节 13：附录（错误码表、字段加密策略、字典清单）。

---

## 1. 设计目标与原则

### 1.1 目标

| 目标       | 描述                                                                       |
| ---------- | -------------------------------------------------------------------------- |
| 全栈对齐   | 一份契约同时驱动前端（TS 类型）、后端（DTO/Validator）、文档（Swagger UI） |
| 可演进     | URL 含版本号 `/api/v1`，新增字段不破坏旧客户端                             |
| 可观测     | 每次请求强制 `X-Request-ID`，错误结构统一带追踪码                          |
| 强一致     | 状态机推进通过专用动作端点（`POST /xxx/{id}/submit` 等），杜绝侧通道改 status |
| 安全合规   | 字段级加密（身份证/手机/银行卡）、列表脱敏、审批流签字哈希落库             |
| 移动端友好 | 单接口聚合学生 5 视图、推送走 SSE/WebSocket                                |

### 1.2 设计原则

1. **资源优先**：URI 用名词复数，CRUD 用 HTTP 动词；状态推进用 `:action` 子资源。
2. **幂等友好**：所有 `PUT/DELETE` 幂等；`POST` 创建型携带 `Idempotency-Key` 头可去重。
3. **统一封包**：业务响应一律 `{code, message, data, request_id}`。
4. **驼峰传输**：JSON 字段使用 `snake_case`（与 DB/后端 Go tag 一致），减少映射成本。
5. **批量与流式**：列表用游标/页码分页，导出走 `Accept: text/csv` + `Content-Disposition`。
6. **悲观时区**：所有时间 RFC3339 + `+08:00`，前端不做转换。

---

## 2. 通用约定

### 2.1 URI 与 HTTP 动词

| 动词    | 用途                                    | 示例                                         |
| ------- | --------------------------------------- | -------------------------------------------- |
| GET     | 查询资源（列表/详情）                   | `GET /ty/applications`                       |
| POST    | 创建 / 触发动作（`:submit/:approve`）   | `POST /ty/applications/{id}/submit`          |
| PUT     | 全量更新（资源所有可写字段）            | `PUT /ty/applications/{id}`                  |
| PATCH   | 部分更新                                | `PATCH /ty/applications/{id}`                |
| DELETE  | 软删除（设置 is_deleted=1）             | `DELETE /ty/applications/{id}`               |

URI 命名约束：

- 模块前缀：`/ty`、`/st`、`/sq`、`/qg`、`/cmp`、`/idx`、`/sys`、`/auth`、`/files`、`/notifications`。
- 资源蛇形复数：`/ty/recommendation-meetings`（保留 `-` 风格）。
- 资源 id 为 `INTEGER`，业务编号查询用 `?biz_no=TY-2026-0001` 或专用端点 `/ty/applications/by-biz-no/{biz_no}`。

### 2.2 鉴权与角色

#### 2.2.1 认证流程

1. 学生 / 教师调用 `POST /auth/login` 获取 `access_token`（30 min）+ `refresh_token`（7 d）。
2. 后续请求携带 `Authorization: Bearer {access_token}`。
3. `access_token` 失效返回 `401 AUTH_TOKEN_EXPIRED`，前端透明刷新。

#### 2.2.2 角色清单（5 级 RBAC）

| 角色编码          | 名称           | 范围         | 主要权限                                 |
| ----------------- | -------------- | ------------ | ---------------------------------------- |
| `R-SY-ADMIN`      | 系统管理员     | 校级         | 所有模块只读 + 用户/角色管理             |
| `R-SY-LEAGUE`     | 校团委         | 校级         | TY 终审、ST 社团/活动审批、CMP 规则      |
| `R-SY-AFFAIRS`    | 学生处         | 校级         | QG 终审、SQ 复核、综合素质权重           |
| `R-COL-LEAGUE`    | 院系团委       | 院系         | TY 院系审批、推优大会监督                |
| `R-COL-COUN`      | 辅导员         | 院系         | TY 推荐、QG 签字、SQ 巡查复核            |
| `R-COL-TUTOR`     | 社团指导教师   | 社团         | ST 活动指导教师审批                      |
| `R-DORM-ADMIN`    | 楼栋管理员     | 楼栋         | SQ 巡查、违规电器、寝室调整              |
| `R-STU-NORM`      | 普通学生       | 自身         | 提交申请、查询本人数据                   |
| `R-STU-LEAGUE`    | 团支书         | 团支部       | TY 推优大会发起、培养记录录入            |
| `R-STU-ASSOC`    | 社团社长/干部   | 单社团       | ST 活动立项、招新、经费报销              |
| `R-STU-COMMUNITY` | 楼层长/寝室长   | 楼层/寝室    | SQ 巡查上报、晚归记录                    |

#### 2.2.3 权限标记法

每个端点用以下标签标注：

```
@auth: required
@roles: R-COL-COUN | R-COL-LEAGUE | R-SY-LEAGUE
@scope: branch={path.branch_id}     # 院系级数据隔离
```

### 2.3 请求/响应封包

#### 2.3.1 响应统一封包

```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "request_id": "01J0X1V8P9KQYWZS2H3FYZRN1A"
}
```

| 字段       | 类型    | 说明                                            |
| ---------- | ------- | ----------------------------------------------- |
| code       | integer | `0`=成功；非零=业务错误（详见 §3）              |
| message    | string  | 人类可读消息（已根据 `Accept-Language` 本地化） |
| data       | object  | 业务负载，列表/分页放在此对象内                 |
| request_id | string  | ULID，与 `X-Request-ID` 响应头一致              |

#### 2.3.2 列表分页（双模式）

**默认页码模式**：

```
GET /ty/applications?page=1&page_size=20&sort=-created_at&keyword=张三
```

```json
{
  "code": 0,
  "data": {
    "items": [ ... ],
    "page": 1,
    "page_size": 20,
    "total": 3421,
    "total_pages": 172
  }
}
```

**游标模式**（事件流、巡查、打卡等大数据量）：

```
GET /event-logs?cursor=eyJpZCI6MTAwfQ==&limit=100
```

```json
{
  "code": 0,
  "data": {
    "items": [ ... ],
    "next_cursor": "eyJpZCI6MjAwfQ==",
    "has_more": true
  }
}
```

#### 2.3.3 排序

`sort` 参数用 `,` 分隔多字段，`-` 前缀表示降序：`sort=-created_at,name`。

#### 2.3.4 字段筛选

`fields=id,name,status` 仅返回指定字段；`expand=student,branch` 触发关联展开。

### 2.4 通用请求头

| 头                   | 必填 | 说明                                         |
| -------------------- | ---- | -------------------------------------------- |
| `Authorization`      | ✓    | `Bearer {jwt}`                               |
| `X-Request-ID`       | 可选 | 客户端可指定，缺省服务端生成                 |
| `Accept-Language`    | 可选 | `zh-CN`(默认) / `en-US`                      |
| `Idempotency-Key`    | 可选 | POST 创建型必填以保证幂等                    |
| `X-Client-Type`      | 可选 | `pc-vue` / `mp-uni` / `mobile-web`           |
| `If-Match`           | 可选 | 乐观锁，值为资源 `etag`（updated_at 哈希）   |

### 2.5 通用响应头

| 头                       | 说明                                |
| ------------------------ | ----------------------------------- |
| `X-Request-ID`           | 同请求                              |
| `X-RateLimit-Remaining`  | 当前窗口剩余次数                    |
| `X-RateLimit-Reset`      | 窗口重置 epoch 秒                   |
| `ETag`                   | 资源版本，配合 `If-Match` 用于更新  |
| `Content-Disposition`    | 导出 / 下载场景                     |

---

## 3. 错误码与异常

### 3.1 错误响应结构

```json
{
  "code": 40301,
  "message": "辅导员未在该院系范围内，无权操作",
  "data": null,
  "request_id": "01J0X1V8P9KQYWZS2H3FYZRN1A",
  "errors": [
    { "field": "branch_id", "rule": "scope", "detail": "branch outside of college" }
  ]
}
```

### 3.2 HTTP 状态与业务码（节选，详见附录 §13.1）

| HTTP | 业务码区间    | 含义               |
| ---- | ------------- | ------------------ |
| 200  | 0             | 成功               |
| 400  | 40000–40999   | 参数/校验          |
| 401  | 40100–40199   | 鉴权未通过/失效    |
| 403  | 40300–40399   | 无权限/范围越权    |
| 404  | 40400–40499   | 资源不存在         |
| 409  | 40900–40999   | 业务冲突/状态机非法 |
| 422  | 42200–42299   | 业务规则未满足     |
| 429  | 42900         | 限流               |
| 500  | 50000–50099   | 服务端错误         |
| 503  | 50300         | 维护中             |

常见业务码示例：

| code  | 含义                                       |
| ----- | ------------------------------------------ |
| 40001 | 必填字段缺失                               |
| 40002 | 字段值非法（CHECK 约束）                   |
| 40101 | Token 失效（请刷新）                       |
| 40102 | 账号被锁定                                 |
| 40301 | 范围越权（学生只能查自己 / 院系隔离）      |
| 40901 | 状态机非法跃迁（如 S0→S3 跳级）            |
| 40902 | 业务唯一约束冲突（同学生 1 份 S1/S2 申请） |
| 42201 | 思想汇报字数 < 1000                        |
| 42202 | 推优大会到会率 < 2/3                       |
| 42203 | 月工时 > 40 小时                           |
| 42204 | 未认定困难生不可申请岗位                   |

---

## 4. 状态机与动作端点

### 4.1 5 态通用状态机

| 取值 | 含义       | 动作                                |
| ---- | ---------- | ----------------------------------- |
| S0   | 草稿       | `:submit` → S1                       |
| S1   | 待审批     | `:approve`/`:reject` → S2/S0         |
| S2   | 院系通过   | `:approve`/`:reject` → S3/S1         |
| S3   | 校级通过   | `:close`/`:revoke`                   |
| S4   | 已归档/失败 | -                                   |

### 4.2 动作端点命名

```
POST /ty/applications/{id}:submit
POST /ty/applications/{id}:approve   # body: {opinion, level}
POST /ty/applications/{id}:reject    # body: {opinion, level}
POST /ty/applications/{id}:revoke
POST /ty/applications/{id}:archive
```

> 冒号前缀的子资源动作端点不会与资源 id 冲突（保留 RFC3986 `:` 字符）。
> 兼容性方案：客户端可使用 `POST /ty/applications/{id}/actions/submit`（等价路由）。

### 4.3 状态查询

```
GET /ty/applications/{id}/timeline    → 返回事件流（event_log 投影）
```

---

## 5. 基础层 API

### 5.1 认证（/auth）

#### 5.1.1 登录

```
POST /auth/login
```

请求：

```json
{
  "username": "20231001",
  "password": "Pwd@2025",
  "captcha_token": "cap_xxx",
  "client_type": "pc-vue"
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "rfk_xxx",
    "token_type": "Bearer",
    "expires_in": 1800,
    "user": {
      "id": 12,
      "username": "20231001",
      "display_name": "张三",
      "avatar_url": null,
      "roles": [
        { "code": "R-STU-NORM", "scope": "student" }
      ],
      "student_id": 88,
      "college_id": 3,
      "must_change_password": false
    }
  }
}
```

#### 5.1.2 刷新

```
POST /auth/refresh
```

> Cookie 中 `refresh_token` 自动携带；如缺失可 body 兜底 `{ "refresh_token": "rfk_xxx" }`。

**语义**：
- 服务端校验 jti 黑名单 + `claims.token_version == sys_user.token_version`；
- 校验通过则**轮换**签发新 token 对（access + refresh），旧 RT jti 立即进黑名单直到其原 exp；
- 任一校验失败返回 `40103 RT_REVOKED`（视为盗用风险，**不**触发自动再试）。

**响应**：

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOi...",
    "token_type":   "Bearer",
    "expires_in":   900
  }
}
```

#### 5.1.3 登出

```
POST /auth/logout
```

**语义**：
- 读取当前 `refresh_token` Cookie，将对应 jti 加入黑名单直到其原 exp；
- 清除 `refresh_token` Cookie；
- 客户端收到 200 后必须清空 `localStorage` 中的 `access_token` 并跳登录页。

#### 5.1.4 当前用户信息

```
GET /auth/me
```

#### 5.1.5 修改密码

```
POST /auth/password
{ "old_password": "...", "new_password": "..." }
```

**语义**：
- 校验旧密码（bcrypt）；
- 更新 `password_hash`、**`sys_user.token_version = token_version + 1`**；
- 副作用：本用户**全部未过期 RT 立即失效**（token_version 失配），前端应在 200 后清登录态并跳登录页；
- 新密码强度按 PRD 规则校验（≥ 8 位、含字母 + 数字、不可与近 3 次相同）。

**错误码**：

| code | 含义 |
| ---- | ---- |
| 40001 | 必填字段缺失 |
| 40002 | 新密码强度不足 |
| 40104 | 旧密码错误 |
| 40103 | 旧 Token 已吊销（请重新登录） |

#### 5.1.6 验证码

```
GET /auth/captcha?type=image    → { token, image_base64 }
POST /auth/captcha/verify       → { token, code }
```

### 5.2 组织与字典（/sys）

| 方法   | 路径                            | 说明                  | 角色             |
| ------ | ------------------------------- | --------------------- | ---------------- |
| GET    | `/sys/colleges`                 | 院系列表              | any logged-in    |
| POST   | `/sys/colleges`                 | 新增院系              | R-SY-ADMIN       |
| PUT    | `/sys/colleges/{id}`            | 更新院系              | R-SY-ADMIN       |
| DELETE | `/sys/colleges/{id}`            | 软删院系              | R-SY-ADMIN       |
| GET    | `/sys/majors`                   | 专业列表              | any              |
| GET    | `/sys/classes`                  | 行政班列表            | any              |
| GET    | `/sys/dorm-buildings`           | 楼栋列表              | any              |
| GET    | `/sys/dicts/{category}`         | 字典查询              | any              |
| GET    | `/sys/dicts`                    | 字典总览（多 category）| any              |
| GET    | `/sys/users`                    | 用户列表              | R-SY-ADMIN       |
| POST   | `/sys/users`                    | 新建用户              | R-SY-ADMIN       |
| POST   | `/sys/users/{id}:reset-password`| 重置密码              | R-SY-ADMIN       |
| POST   | `/sys/users/{id}:lock`          | 锁定                  | R-SY-ADMIN       |
| POST   | `/sys/users/{id}:unlock`        | 解锁                  | R-SY-ADMIN       |
| GET    | `/sys/roles`                    | 角色列表              | R-SY-ADMIN       |
| POST   | `/sys/users/{id}/roles`         | 授予角色              | R-SY-ADMIN       |
| DELETE | `/sys/users/{id}/roles/{rid}`   | 撤销角色              | R-SY-ADMIN       |

#### 字典查询示例

```
GET /sys/dicts/difficulty_level
```

```json
{
  "code": 0,
  "data": {
    "category": "difficulty_level",
    "items": [
      { "code": "special", "name_zh": "特别困难", "sort": 1 },
      { "code": "hard",    "name_zh": "困难",     "sort": 2 },
      { "code": "normal",  "name_zh": "一般困难", "sort": 3 },
      { "code": "none",    "name_zh": "不困难",   "sort": 4 }
    ]
  }
}
```

### 5.3 文件 (/files)

#### 5.3.1 申请上传 token（前端直传 / 后端代传二选一）

```
POST /files/upload-tokens
{
  "module": "TY",
  "biz_type": "thought_report",
  "original_name": "2026Q1.docx",
  "size_bytes": 1048576,
  "mime_type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "sha256": "..."
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "file_id": 9211,
    "upload_url": "/files/uploads/9211",
    "method": "PUT",
    "headers": { "Content-Type": "..." },
    "expires_at": "2026-06-14T10:30:00+08:00"
  }
}
```

#### 5.3.2 上传二进制

```
PUT /files/uploads/{file_id}     (multipart 或 raw binary)
```

#### 5.3.3 元数据 / 下载

```
GET    /files/{id}                  → 返回元数据（鉴权后给签名 URL）
GET    /files/{id}/download         → 302 跳转私有签名 URL（10 min）
DELETE /files/{id}                  → 软删（仅上传者/管理员）
```

#### 文件可见性

`visibility ∈ private | org | public`，列表/详情接口对外仅返回当前角色可见文件。

### 5.4 通知 (/notifications)

| 方法   | 路径                                  | 说明                       |
| ------ | ------------------------------------- | -------------------------- |
| GET    | `/notifications?is_read=0`            | 我的通知列表               |
| GET    | `/notifications/unread-count`         | 未读数（badge 用）         |
| POST   | `/notifications/{id}:read`            | 标记已读                   |
| POST   | `/notifications:read-all`             | 全部已读                   |
| POST   | `/notifications/{id}:archive`         | 归档                       |
| GET    | `/notifications/stream`               | SSE 推送（Authorization 走 query token） |

### 5.5 审计与事件

| 方法 | 路径                          | 角色          | 说明                      |
| ---- | ----------------------------- | ------------- | ------------------------- |
| GET  | `/event-logs`                 | R-SY-ADMIN    | 业务事件检索（游标分页）  |
| GET  | `/audit-logs`                 | R-SY-ADMIN    | 访问审计                  |
| GET  | `/event-logs/by-aggregate`    | scope-aware   | `?aggregate=ty.application&aggregate_id=88` |

### 5.6 工作台（/dashboard）

首页工作台接口，按登录用户角色域（`student` / `college` / `school`）返回差异化内容。

| 方法 | 路径                  | 角色   | 说明                                       |
| ---- | --------------------- | ------ | ------------------------------------------ |
| GET  | `/dashboard/overview` | 已登录 | 工作台概览（用户、stats、待办、快捷入口） |

`Overview.data` 字段：

| 字段         | 类型        | 说明                                   |
| ------------ | ----------- | -------------------------------------- |
| user         | object      | 当前用户基本信息（用户名/显示名/角色） |
| role_scope   | string      | 角色域：student / college / school     |
| stats        | object      | 指标卡（按角色填充，详见下表）         |
| todo_items   | TodoItem[]  | 待办事项                               |
| quick_links  | QuickLink[] | 快捷入口                               |

`stats` 字段（按角色填充；未使用的字段对当前角色始终为 0/空）：

| 字段名                    | 类型    | 角色域  | 含义                                                                 |
| ------------------------- | ------- | ------- | -------------------------------------------------------------------- |
| my_ty_status              | string  | student | 我的入团申请状态短码（S1~S6 / S7_MEMBER / 长码）                     |
| my_cmp_score              | int     | student | 我的综合分（最近学期）                                               |
| my_activity_count         | int     | student | 我参加的活动数（基于 `st_activity_checkins`）                        |
| unread_noti_count         | int     | all     | 未读通知数                                                           |
| recruiting_assoc_count    | int     | student | **已废弃**（v1.0 起改用 `recruiting_plan_count`）                            |
| recruiting_plan_count     | int     | student | 招新中计划数（`st_recruit_plan.status='S3' AND is_finished=0` 的计划数）       |
| student_count             | int     | staff   | 管辖学生数（辅导员=所带班级；其他=全校）                              |
| ty_pending_count          | int     | staff   | 待审批入团申请数                                                     |
| incident_open_count       | int     | staff   | 未关闭社区事件数                                                     |
| qg_position_count         | int     | staff   | 在岗勤工岗位数                                                       |
| active_assoc_count        | int     | staff   | 活跃社团数                                                           |

---

## 6. 模块一 · TY 团员发展

### 6.1 资源映射

| 资源              | 路径前缀                       | 主表                     |
| ----------------- | ------------------------------ | ------------------------ |
| 团支部            | `/ty/branches`                 | `ty_branch`              |
| 团员花名册        | `/ty/members`                  | `ty_member_roster`       |
| 入团申请          | `/ty/applications`             | `ty_application`         |
| 推优大会          | `/ty/recommendation-meetings`  | `ty_recommendation_meeting` |
| 推优投票          | `…/{meeting_id}/votes`         | `ty_recommendation_vote` |
| 培养联系人        | `/ty/cultivation-links`        | `ty_cultivation_link`    |
| 培养考察记录      | `/ty/cultivation-records`      | `ty_cultivation_record`  |
| 团课记录          | `/ty/course-records`           | `ty_course_record`       |
| 思想汇报          | `/ty/thought-reports`          | `ty_thought_report`      |
| 发展对象          | `/ty/development-objects`      | `ty_development_object`  |
| 政审              | `/ty/political-reviews`        | `ty_political_review`    |
| 发展大会          | `/ty/development-meetings`     | `ty_development_meeting` |
| 预备期考察        | `/ty/probationary-records`     | `ty_probationary_record` |
| 转正大会          | `/ty/probationary-meetings`    | `ty_probationary_meeting` |

### 6.2 入团申请（核心状态机示范）

#### 6.2.1 列表

```
GET /ty/applications?status=S1,S2&branch_id=12&keyword=张三&page=1&page_size=20
```

`@roles`:

- 学生只能看本人申请；`branch_id` 必须等于本人所在团支部，否则 403。
- 团支书 / 辅导员限制为 `branch.college_id = 当前用户.scope_college_id`。

响应 `items[*]`：

```json
{
  "id": 88,
  "biz_no": "TY-2026-0001",
  "student": { "id": 12, "name": "张三", "student_no": "20231001", "class_name": "软工23-1" },
  "branch": { "id": 4, "name": "软件学院团支部" },
  "apply_date": "2026-03-15",
  "status": "S1",
  "status_text": "辅导员审批中",
  "counselor_user_id": null,
  "counselor_at": null,
  "created_at": "2026-03-15T10:23:01+08:00",
  "updated_at": "2026-03-15T10:23:01+08:00"
}
```

#### 6.2.2 详情

```
GET /ty/applications/{id}?expand=branch,student,votes,cultivation_records,thought_reports
```

#### 6.2.3 创建草稿（学生）

```
POST /ty/applications
{
  "branch_id": 4,
  "apply_date": "2026-03-15",
  "self_statement": "...（≥500 字）",
  "family_members": [
    { "relation": "father", "name": "张大山", "political_status": "群众", "occupation": "农民" }
  ],
  "rewards_punishments": "..."
}
```

业务规则：

- BR-TY 同学生同一时间只允许 1 份 `S1/S2` → 服务端 `409 / 40902`。
- 申请人年龄 14–28（解析身份证）→ 服务端 `422 / 42210`。
- `self_statement` 长度 < 500 → `400 / 40002`。

#### 6.2.4 提交 → 审批

```
POST /ty/applications/{id}:submit                   # 学生：S0→S1
POST /ty/applications/{id}:approve                  # 辅导员/院系/校级：S1→S2→S3
{
  "level": "counselor",          # counselor | college | school
  "opinion": "表现积极，同意推荐"
}
POST /ty/applications/{id}:reject
{
  "level": "counselor",
  "opinion": "近期违纪记录，暂缓"
}
POST /ty/applications/{id}:revoke                   # 学生在 S1 前撤回
```

返回新状态及 timeline 最新事件，前端据此更新 UI。

### 6.3 推优大会

#### 6.3.1 创建大会

```
POST /ty/recommendation-meetings
{
  "application_id": 88,
  "meeting_at": "2026-03-25T14:00:00+08:00",
  "location": "图书馆 301",
  "expected_count": 30,
  "actual_count": 26,
  "photo_overall_id": 9201,
  "photo_vote_id": 9202,
  "decision": "pass",
  "decision_reason": "全票通过",
  "votes": [
    { "approve_count": 25, "against_count": 0, "abstain_count": 1 }
  ]
}
```

校验：

- `actual_count >= ceil(expected_count * 2/3)`，否则 `422 / 42202`。
- 必须存在 `photo_overall_id` 与 `photo_vote_id` → 否则 `400 / 40001`。
- `approve_count > actual_count/2` 才允许 `decision='pass'`。

### 6.4 思想汇报

```
POST /ty/thought-reports
{
  "application_id": 88,
  "title": "2026 年第一季度思想汇报",
  "content": "...（≥1000 字）",
  "quarter": "2026Q1"
}
```

服务端：

- 调用 AI 查重服务，`ai_similarity > 0.30` → `422 / 42201`，`is_qualified=0`；
- 字数不足 → 同样拒收。

```
GET /ty/thought-reports?application_id=88&keyword=...   # FTS5 全文检索
```

### 6.5 培养考察 / 团课 / 政审 / 发展大会 / 转正

| 资源                  | 主要端点                                                           |
| --------------------- | ------------------------------------------------------------------ |
| 培养联系人            | `POST /ty/cultivation-links`，`PATCH /…/{id}`，`POST /…/{id}:end`  |
| 月度培养记录          | `POST /ty/cultivation-records`，列表 `GET ?application_id=`        |
| 团课记录              | `POST /ty/course-records`，`GET ?student_id=…&semester=…`          |
| 发展对象              | `POST /ty/development-objects`，`POST /…/{id}:publicize`（公示）   |
| 政审                  | `POST /ty/political-reviews`（per relation）                       |
| 发展大会              | `POST /ty/development-meetings`，自动产出 `ty_member_roster`        |
| 预备期考察（季度）    | `POST /ty/probationary-records`                                    |
| 转正大会              | `POST /ty/probationary-meetings:vote` + `:close`                   |

**说明**：预备期考察与转正大会的列表接口 `GET /ty/probationary-records` 与 `GET /ty/probationary-meetings` 均支持可选 query `application_id`（按申请ID过滤），缺省时返回全部分页数据，供"转正流程管理"列表页使用；分页参数 `page`、`page_size`（默认 20，最大 100）。返回结构同 §2.3 分页封包：`{ items, total, page, page_size }`。

### 6.6 团员花名册

```
GET /ty/members?branch_id=4&status=active&keyword=张
PATCH /ty/members/{id}                  # 修改团员证号、维护信息
POST /ty/members/{id}:transfer-out      # 转出
POST /ty/members/{id}:overtime          # 标记超龄（也可定时任务自动）
POST /ty/members/{id}:archive           # 归档（保留 5 年）
```

### 6.7 TY 业务规则到接口的落点

| 规则       | 接口落点                                                                   |
| ---------- | -------------------------------------------------------------------------- |
| BR-TY-01   | 状态机引擎；前端禁用越级按钮，后端兜底 409                                 |
| BR-TY-02   | 推优大会 photo 双照片必传                                                  |
| BR-TY-03   | 定时任务 `POST /ty/members/{id}:overtime` 系统调用                          |
| BR-TY-04   | `/ty/members/{id}:archive`，记录 `archive_keep_until`                      |
| BR-TY-05   | `ty_member_roster.biz_no UNIQUE`，重复返回 `409 / 40903`                    |
| BR-TY-06   | 思想汇报 422 校验                                                          |

---

## 7. 模块二 · ST 社团活动

### 7.1 资源映射

| 资源       | 路径前缀                          |
| ---------- | --------------------------------- |
| 社团       | `/st/associations`                |
| 章程       | `/st/charters`                    |
| 发起人     | `/st/founders`                    |
| 成员       | `/st/members`                     |
| 招新计划   | `/st/recruit-plans`               |
| 招新申请   | `/st/recruit-applies`             |
| 活动立项   | `/st/activities`                  |
| 活动审批   | `/st/activities/{id}/approvals`   |
| 签到       | `/st/activities/{id}/checkins`    |
| 总结       | `/st/activities/{id}/summary`     |
| 照片       | `/st/activities/{id}/photos`      |
| 报销       | `/st/expenses`                    |
| 换届       | `/st/elections`                   |
| 评优       | `/st/ratings`                     |
| 黑名单     | `/st/blacklists`                  |

### 7.2 社团（Association）

#### 7.2.1 创建（试运行）

```
POST /st/associations
{
  "name": "无人机科创社",
  "college_id": 3,
  "tutor_user_id": 502,
  "president_student_id": 88,
  "business_scope": "...",
  "founders": [88, 91, 102],
  "charter_file_id": 9301,
  "chapter_count": 12
}
```

校验：

- 章程章节数 < 10 → `422`；
- 同名社团 3 年内不可复用 → `409 / 40904`（基于解散历史快照）；
- 指导教师同期 ≤ 3 个社团 → `422 / 42220`；
- 指导教师在 `st_blacklist` 内未到期 → `422 / 42221`。

#### 7.2.2 状态推进

```
POST /st/associations/{id}:start-trial      → preparing → trial（自动 +6 个月）
POST /st/associations/{id}:register         → trial → registered（满足条件）
POST /st/associations/{id}:rectify          → registered → rectifying（管理员）
POST /st/associations/{id}:resume           → rectifying → registered
POST /st/associations/{id}:cancel           → 任何 → cancelled（理由必填）
```

### 7.3 招新

```
POST /st/recruit-plans                  # 社长发起
POST /st/recruit-plans/{id}:approve     # 院系/校社联
POST /st/recruit-plans/{id}:publish     # 发布
POST /st/recruit-plans/{id}:finish      # 提前结束招新（仅 S3 状态可用，不可逆）

POST /st/recruit-applies                # 学生投递（同一学年最多 3 社团）
POST /st/recruit-applies/{id}:result    # 录入面试结果（accepted/rejected）
```

#### 7.3.1 提前结束招新

```
POST /api/v1/st/recruit-plans/{id}/finish
Content-Type: application/json
{ "reason": "招新人数已满足，提前结束" }
```

规则：
- 仅 `status='S3'`（已通过/可投递）的计划可被结束；其他状态 → `409 / 40901`
- 仅 `is_finished=0`（招新中）的计划可被结束；重复结束 → `409 / 40901`
- 权限：与招新计划审批同集（`R-SY-ADMIN` / `R-SY-LEAGUE` / `R-COL-LEAGUE` / `R-COL-COUN` / `R-COL-TUTOR`）
- 操作后写入：`is_finished=1`、`finished_at=now`、`finished_by=actor_user_id`、`finished_reason=request.reason`
- 写业务事件 `StRecruitPlanFinished`（aggregate=`st.recruit_plan`）
- 结束操作**不可逆**（不提供 cancel）

响应：`200`，body 为更新后的 `RecruitPlanView`（含 `is_finished / finished_at / finished_by / finished_reason`）。

#### 7.3.2 投递校验

`POST /st/recruit-applies` 在原有校验（plan 存在、status=S3、唯一性、3 社团上限）之上，新增：
- 计划 `is_finished=1` → 拒绝 `409 / 40901`（message: `该招新计划已结束，不可投递`）

> 同一学年学生加入社团数量校验在 `POST /st/recruit-applies` 时聚合；返回 `422 / 42230`。

### 7.4 活动立项 + 审批 + 实施

#### 7.4.1 创建

```
POST /st/activities
{
  "association_id": 12,
  "title": "无人机校园飞行体验",
  "activity_level": "B",
  "expected_participants": 200,
  "budget_cents": 80000,
  "plan_file_id": 9401,
  "emergency_plan_file_id": 9402,
  "safety_commit_file_id": null,
  "location": "操场",
  "started_at": "2026-04-12T13:30:00+08:00",
  "ended_at": "2026-04-12T17:00:00+08:00",
  "expected_count": 50
}
```

校验（按 `activity_level`）：

- A/B：`emergency_plan_file_id` 必传；
- ≥500 人或户外：`safety_commit_file_id` 必传；
- `plan_file` 字数后端校验 ≥ 1000；
- `reject_count >= 3` 30 天内禁止再创建（`422 / 42233`）。

#### 7.4.2 审批

```
POST /st/activities/{id}:submit
POST /st/activities/{id}:approve
{ "step_no": 2, "opinion": "同意…(≥30 字)" }
POST /st/activities/{id}:reject
{ "step_no": 2, "opinion": "..." }
```

#### 7.4.3 签到

```
POST /st/activities/{id}/checkins:qr-scan       # 学生扫码
{
  "qr_token": "...",
  "geo": { "lat": 30.5, "lng": 114.3 }
}
GET  /st/activities/{id}/checkins?student_id=
POST /st/activities/{id}/checkins/manual         # 干部补登（写入 method=manual + 审批留痕）
```

#### 7.4.4 总结 + 照片

```
POST /st/activities/{id}/summary
{
  "actual_participants": 178,
  "achievement_score": 88,
  "suggestions": "..."
}
POST /st/activities/{id}/photos
{
  "files": [{ "file_id": 9501, "caption": "全景" }]
}
```

校验：照片 ≥ 3 张，否则 `422 / 42234`；提交时间 > 结束时间 + 3 工作日 ⇒ `is_overdue=1`。

### 7.5 经费报销

```
POST /st/expenses
{
  "activity_id": 23,
  "amount_cents": 9800,
  "invoice_count": 3,
  "invoice_files": [9601, 9602, 9603]
}
POST /st/expenses/{id}:review            # 单签
POST /st/expenses/{id}:co-sign           # > 1 万元第二签
POST /st/expenses/{id}:pay
```

### 7.6 评优 / 换届 / 黑名单

```
POST /st/ratings:compute?academic_year=2025-2026   # 校社联触发批量打分
GET  /st/ratings?academic_year=2025-2026
POST /st/ratings/{id}:public-vote                  # 5 星全校公投开启
POST /st/elections                                  # 换届创建
POST /st/elections/{id}:publicize                   # 公示
POST /st/blacklists                                 # 加入黑名单（用户+理由+时长）
DELETE /st/blacklists/{id}                          # 解除
```

---

## 8. 模块三 · SQ 学生社区与自治

### 8.1 资源映射

| 资源       | 路径前缀                            |
| ---------- | ----------------------------------- |
| 自治职务   | `/sq/positions`                     |
| 巡查       | `/sq/inspections`                   |
| 巡查扣分项 | `/sq/inspections/{id}/deductions`   |
| 异常事件   | `/sq/incidents`                     |
| 事件附件   | `/sq/incidents/{id}/attachments`    |
| 处置记录   | `/sq/incidents/{id}/actions`        |
| 自治活动   | `/sq/activities`                    |
| 考核       | `/sq/assessments`                   |
| 晚归       | `/sq/late-returns`                  |
| 违规电器   | `/sq/violations`                    |
| 寒暑假留校 | `/sq/vacation-stays`                |
| 寝室调整   | `/sq/room-changes`                  |

### 8.2 自治职务

```
POST /sq/positions
{
  "student_id": 88,
  "scope_type": "floor",
  "scope_id": 12,
  "position": "floor_leader",
  "start_at": "2026-03-01"
}
POST /sq/positions/{id}:publicize        # 进入公示（≥3 天）
POST /sq/positions/{id}:appoint          # 任命转 formal
POST /sq/positions/{id}:dismiss          # 解聘（理由）
POST /sq/positions/{id}:renew            # 续任
```

跨模块校验：`student_id` 不可同时为团支书 + 寝室长（BR-SQ-01）→ `422 / 42240`。

### 8.3 巡查

```
POST /sq/inspections
{
  "inspection_type": "hygiene",
  "building_id": 1,
  "floor_id": 12,
  "room_id": 305,
  "inspected_at": "2026-04-01T19:00:00+08:00",
  "score": 85,
  "summary": "床铺凌乱",
  "deductions": [
    { "item": "床铺整理", "deduction": 5, "photo_file_id": 9701 },
    { "item": "桌面物品", "deduction": 10, "photo_file_id": 9702 }
  ]
}
```

> 提交后自动派生 `score = 100 - sum(deductions)`，前端不用计算；服务端兜底校验。

```
GET /sq/inspections?building_id=1&inspected_at_from=2026-04-01&inspected_at_to=2026-04-30
GET /sq/buildings/{id}/score-trend?period=monthly        # 楼栋评分趋势（聚合接口）
```

### 8.4 异常事件 4 等级响应

```
POST /sq/incidents
{
  "incident_level": "L3",
  "incident_type": "violation_appliance",
  "occurred_at": "...",
  "building_id": 1,
  "room_id": 305,
  "reporter_user_id": 502,
  "involved_student_ids": [88, 91],
  "initial_action": "现场查处...",
  "attachments": [9801, 9802]
}
```

服务端按 `incident_level` 触发通知 SLA：L1 30 分钟内 → 楼栋管理员；L4 立即 → 学生处。

```
POST /sq/incidents/{id}:assign         # 指派处置人
POST /sq/incidents/{id}:add-action     # 处置过程
POST /sq/incidents/{id}:close
{
  "final_action": "...",
  "closed_by": 502
}
```

校验：`incident_level='L4'` 时 `closed_by` 必须为教师（`R-COL-COUN/R-DORM-ADMIN/R-SY-AFFAIRS`）→ 否则 `403 / 40310`。

### 8.5 自治活动 / 考核 / 晚归 / 违规电器 / 寒暑假 / 调整

| 接口                                          | 用途                                              |
| --------------------------------------------- | ------------------------------------------------- |
| `POST /sq/activities`                         | 自治活动立项（同 ST 简化版状态机）                |
| `POST /sq/assessments:compute?cycle_key=...`  | 月度/学期考核批量计算                             |
| `GET  /sq/assessments?target_user_id=`        | 查询某楼层长考核                                  |
| `POST /sq/late-returns`                       | 晚归登记（学期累计 3 次自动通知 BR-SQ-02）         |
| `POST /sq/violations`                         | 违规电器（含照片+学生签字）                       |
| `POST /sq/violations/{id}:report-college`     | 二次违规上报学院                                  |
| `POST /sq/vacation-stays`                     | 寒暑假留校申请（必须 ≥ 14 天前）                   |
| `POST /sq/room-changes`                       | 寝室调整（双签：辅导员+楼层会）                   |

---

## 9. 模块四 · QG 勤工助学

### 9.1 资源映射

| 资源       | 路径前缀                       |
| ---------- | ------------------------------ |
| 困难认定   | `/qg/difficulty-certs`         |
| 岗位       | `/qg/positions`                |
| 岗位申请   | `/qg/applies`                  |
| 工时打卡   | `/qg/attendances`              |
| 补卡       | `/qg/makeup-attends`           |
| 请假       | `/qg/leaves`                   |
| 月度考核   | `/qg/monthly-assessments`      |
| 薪酬       | `/qg/payrolls`                 |
| 薪酬明细   | `/qg/payrolls/{id}/details`    |
| 续聘/解聘  | `/qg/renewal-terms`            |
| 申诉       | `/qg/complaints`               |

### 9.2 困难认定

```
POST /qg/difficulty-certs
{
  "academic_year": "2025-2026",
  "level": "hard",
  "cert_files": [9901, 9902, 9903]
}
POST /qg/difficulty-certs/{id}:submit
POST /qg/difficulty-certs/{id}:approve   { "level": "college" | "school" }
POST /qg/difficulty-certs/{id}:reject    { "opinion": "..." }
POST /qg/difficulty-certs/{id}:publicize # 公示 ≥5 工作日，自动结束转 S3
```

唯一约束：`(student_id, academic_year)` → `409 / 40905`。

### 9.3 岗位发布与申请

```
POST /qg/positions
{
  "dept_type": "admin",
  "dept_name": "图书馆",
  "title": "图书整理员",
  "description": "...",
  "headcount": 4,
  "weekly_hours_limit": 12,
  "hourly_rate_cents": 1500,
  "start_at": "2026-04-01",
  "end_at": "2026-07-31",
  "supervisor_user_id": 700,
  "risk_notes": null,
  "kpi": { ... }
}
POST /qg/positions/{id}:publish

POST /qg/applies
{
  "position_id": 31,
  "resume_file_id": 9920
}
```

服务端校验链：

1. 学生 `qg_difficulty_cert` 必须存在 `level != none` 且当年度 `status='S3'`，否则 `422 / 42204`。
2. 同岗位禁止重复投递。
3. 危险工种（`risk_notes` 非空）⇒ 学生确认 token 必传。
4. 寒暑假岗位 `end_at - start_at <= 60 天`。

#### 录用流转

```
POST /qg/applies/{id}:interview            # 排面试
POST /qg/applies/{id}:accept               # 录用，触发 confirm_deadline +3 工作日
POST /qg/applies/{id}:confirm              # 学生确认
POST /qg/applies/{id}:onboard
POST /qg/applies/{id}:offboard
```

### 9.4 工时打卡

```
POST /qg/attendances:clock-in
{
  "apply_id": 91,
  "method": "gps_face",
  "geo": { "lat": ..., "lng": ... },
  "face_token": "..."
}
POST /qg/attendances:clock-out             # 同 apply_id 当日匹配 in
GET  /qg/attendances?apply_id=&work_date_from=&work_date_to=
GET  /qg/attendances/monthly-summary?apply_id=&year=2026&month=4
```

服务端：

- 每日 1 条；不允许跨日。
- 月累计 > 40h 或周累计 > 20h，`clock-in` 直接返回 `422 / 42203`。
- `late_minutes / early_minutes` 自动计算并提示。

### 9.5 补卡 / 请假

```
POST /qg/makeup-attends
POST /qg/makeup-attends/{id}:approve     # 双签
POST /qg/leaves
POST /qg/leaves/{id}:approve
```

### 9.6 月度考核 + 薪酬

```
POST /qg/monthly-assessments:compute?year=2026&month=4
GET  /qg/monthly-assessments?year=2026&month=4&apply_id=

POST /qg/payrolls:compute?year=2026&month=4
{ "scope": "all" | "college" | "apply_ids" }
GET  /qg/payrolls?year=2026&month=4&status=draft
POST /qg/payrolls/{id}:review
POST /qg/payrolls/{id}:pay
{
  "bank_batch_no": "...",
  "paid_at": "2026-05-10T14:00:00+08:00"
}
POST /qg/payrolls/{id}:mark-failed
{ "failure_reason": "..." }
```

返回示例（脱敏）：

```json
{
  "id": 1234,
  "biz_no": "QG-PAY-202604-0001",
  "student": { "id": 88, "name": "张***" },
  "total_hours": 32.0,
  "gross_cents": 48000,
  "tax_cents": 0,
  "net_cents": 48000,
  "coefficient": 1.0,
  "bank_account_last4": "8821",
  "status": "reviewed"
}
```

### 9.7 续聘/解聘 与 申诉

```
POST /qg/renewal-terms          # type = renewal | termination
POST /qg/renewal-terms/{id}:counselor-sign
POST /qg/renewal-terms/{id}:affairs-sign
POST /qg/complaints                          # 学生申诉（attendance/assess/payroll）
POST /qg/complaints/{id}:reply
{ "result": "...", "decision": "support" | "reject" }
```

> 自动化：连续 2 月考核 < 60 → 系统自动 `POST /qg/renewal-terms` type=termination；连续 3 次无故缺勤同步触发。

---

## 10. CMP 综合素质量化 / IDX 学生画像

### 10.1 综合素质量化（/cmp）

| 方法 | 路径                                      | 用途                                  |
| ---- | ----------------------------------------- | ------------------------------------- |
| GET  | `/cmp/scores`                             | 列表 + 排序（年度/班级/学院维度）     |
| GET  | `/cmp/scores/{id}`                        | 详情，含 `details[]` 分维度           |
| GET  | `/cmp/scores/me`                          | 学生本人当年度分数                    |
| GET  | `/cmp/scores/me/history`                  | 跨学年历史                            |
| POST | `/cmp/scores:compute`                     | 触发批量计算（指定年度/范围）         |
| GET  | `/cmp/rule-versions`                      | 规则版本列表                          |
| POST | `/cmp/rule-versions`                      | 新增规则版本（草稿）                  |
| POST | `/cmp/rule-versions/{id}:activate`        | 设为生效                              |

`GET /cmp/scores/{id}` 返回：

```json
{
  "id": 1001,
  "student": { "id": 88, "name": "张三" },
  "academic_year": "2025-2026",
  "total_score": 86.5,
  "rank_in_class": 3,
  "rank_in_college": 27,
  "rule_version": "v2026.1",
  "details": [
    { "dimension": "league",    "sub_item": "团内任职",   "score": 8,  "weight": 0.20 },
    { "dimension": "assoc",     "sub_item": "活动组织",   "score": 12, "weight": 0.20 },
    { "dimension": "community", "sub_item": "文明寝室",   "score": 9,  "weight": 0.15 },
    { "dimension": "workstudy", "sub_item": "履职完成度", "score": 10, "weight": 0.15 },
    { "dimension": "academic",  "sub_item": "GPA",        "score": 47.5, "weight": 0.30 }
  ],
  "computed_at": "2026-06-01T02:00:00+08:00"
}
```

### 10.2 学生画像（/idx）

| 方法 | 路径                                | 用途                                              |
| ---- | ----------------------------------- | ------------------------------------------------- |
| GET  | `/idx/students`                     | 学生检索（学号、姓名、院系、班级、政治面貌）      |
| GET  | `/idx/students/{id}`                | 学生详情（脱敏）                                  |
| GET  | `/idx/students/{id}/profile`        | 一站式画像（聚合 TY/ST/SQ/QG/CMP 关键指标）       |
| GET  | `/idx/students/{id}/activities`    | 跨模块时间轴                                      |
| PATCH| `/idx/students/{id}`                | 更新基础信息（学生处）                            |
| POST | `/idx/students/{id}:set-difficulty`| 直接标记困难等级（同步认定记录）                  |

`GET /idx/students/{id}/profile` 响应（聚合）：

```json
{
  "student": { "id": 88, "name": "张三", "student_no": "20231001", "college": "软件学院", "class": "软工23-1" },
  "ty": { "status": "probationary", "biz_no_apply": "TY-2026-0001", "thought_reports_qualified": 3 },
  "st": { "associations": 2, "core_officer": false, "activities_attended": 14 },
  "sq": { "is_floor_leader": true, "late_returns_semester": 0, "violations": 0 },
  "qg": { "active_position_title": "图书整理员", "this_month_hours": 32.0, "total_income_cents": 192000 },
  "cmp": { "total_score": 86.5, "rank_in_class": 3 }
}
```

---

## 11. 实时与流式

### 11.1 SSE 通知流

```
GET /notifications/stream?token={jwt}
Accept: text/event-stream
```

事件帧：

```
event: notification
id: 01J0X1V8P9KQYWZS2H3FYZRN1A
data: {"id": 555, "title": "您的入团申请已被辅导员批准", "link_url": "/ty/applications/88"}
```

### 11.2 状态变更流（管理端大屏）

```
GET /sq/incidents/stream?level=L4
GET /qg/payrolls/stream?status=paid
```

### 11.3 Webhook（外部系统接入，可选）

```
POST /sys/webhooks
{
  "name": "外部党建系统",
  "url": "https://partybuild.example.com/hook",
  "events": ["TyMemberJoined","TyMemberOvertime"],
  "secret": "..."
}
```

签名：`X-StudentHub-Signature: sha256={hmac(payload, secret)}`。

---

## 12. OpenAPI 3.0 骨架

> 完整 YAML 与本 SRD 同步生成于 `docs/openapi/v1.yaml`（待实施阶段产出）。下面给出骨架与可复用组件。

### 12.1 Servers / Security

```yaml
openapi: "3.0.3"
info:
  title: StudentHub API
  version: "1.0.0"
servers:
  - url: https://{domain}/api/v1
    variables:
      domain:
        default: studenthub.local
security:
  - bearerAuth: []
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

### 12.2 通用 Schemas

```yaml
components:
  schemas:
    ApiResponse:
      type: object
      required: [code, message, data, request_id]
      properties:
        code: { type: integer, example: 0 }
        message: { type: string, example: ok }
        data: { nullable: true }
        request_id: { type: string }
    Pagination:
      type: object
      properties:
        page: { type: integer }
        page_size: { type: integer }
        total: { type: integer }
        total_pages: { type: integer }
    Error:
      allOf:
        - $ref: "#/components/schemas/ApiResponse"
        - type: object
          properties:
            errors:
              type: array
              items:
                type: object
                properties:
                  field: { type: string }
                  rule:  { type: string }
                  detail:{ type: string }
    StatusEnum:
      type: string
      enum: [S0, S1, S2, S3, S4]
    Money:
      type: integer
      description: 金额，单位：分
      minimum: 0
```

### 12.3 资源 Schema 节选

```yaml
TyApplication:
  type: object
  required: [id, biz_no, student_id, branch_id, apply_date, status]
  properties:
    id: { type: integer, format: int64 }
    biz_no: { type: string, pattern: '^TY-\d{4}-\d{4}$' }
    student_id: { type: integer }
    branch_id: { type: integer }
    apply_date: { type: string, format: date }
    self_statement: { type: string, minLength: 500 }
    family_members:
      type: array
      items:
        type: object
        properties:
          relation: { type: string, enum: [father, mother, spouse, sibling, other] }
          name: { type: string }
          political_status: { type: string }
          occupation: { type: string }
    status: { $ref: "#/components/schemas/StatusEnum" }
    counselor_opinion: { type: string }
    college_opinion: { type: string }
    school_opinion: { type: string }
    created_at: { type: string, format: date-time }
    updated_at: { type: string, format: date-time }
```

### 12.4 Mock Server

提供两套 Mock：

1. **Prism**：`npx @stoplight/prism mock docs/openapi/v1.yaml --port 4010`；
2. **本地 Python Mock**：见 ADR / api-dev skill 推荐脚本。

---

## 13. 附录

### 13.1 错误码表（核心）

| code  | HTTP | 含义                       |
| ----- | ---- | -------------------------- |
| 0     | 200  | 成功                       |
| 40001 | 400  | 必填字段缺失               |
| 40002 | 400  | 字段值非法                 |
| 40003 | 400  | 文件类型不支持             |
| 40004 | 400  | 文件大小超限               |
| 40101 | 401  | Token 失效                 |
| 40102 | 401  | 账号锁定                   |
| 40103 | 401  | 验证码错误                 |
| 40301 | 403  | 范围越权                   |
| 40302 | 403  | 角色未授权                 |
| 40310 | 403  | 必须由教师执行该操作       |
| 40404 | 404  | 资源不存在                 |
| 40901 | 409  | 状态机非法跃迁             |
| 40902 | 409  | 唯一约束冲突（业务规则）   |
| 40903 | 409  | 团员证号重复               |
| 40904 | 409  | 同名社团 3 年内不可复用    |
| 40905 | 409  | 困难认定本年度已存在       |
| 42201 | 422  | 思想汇报不达标（字数/查重）|
| 42202 | 422  | 推优大会到会率不足         |
| 42203 | 422  | 工时上限超出               |
| 42204 | 422  | 未认定困难生               |
| 42210 | 422  | 申请人年龄不符             |
| 42220 | 422  | 指导教师社团数超限         |
| 42221 | 422  | 指导教师在黑名单           |
| 42230 | 422  | 同学年加入社团数超限       |
| 42233 | 422  | 累计驳回 ≥3，30 天内禁止    |
| 42234 | 422  | 活动照片不足 3 张          |
| 42240 | 422  | 跨模块身份冲突             |
| 42900 | 429  | 限流                        |
| 50000 | 500  | 服务端异常                 |
| 50300 | 503  | 维护中                     |

### 13.2 字段加密与脱敏

| 字段                              | 加密         | API 默认脱敏          | 解密授权              |
| --------------------------------- | ------------ | --------------------- | --------------------- |
| `idx_student.id_card`             | AES-256-GCM  | `110***********0023`  | R-SY-AFFAIRS / 本人   |
| `idx_student.phone`               | AES-256-GCM  | `138****5678`         | 本人 / 辅导员         |
| `qg_payroll.bank_account`         | AES-256-GCM  | 仅末四位              | 财务 / 学生处         |
| `ty_political_review.target_id_card` | AES-256-GCM | 全部 `***`           | 校团委审核流程内      |

> 解密获取需通过显式 `?reveal=phone` 等参数 + 二次身份校验，全部入审计日志。

### 13.3 业务编号格式

| 模块         | 前缀         | 示例                  |
| ------------ | ------------ | --------------------- |
| TY 申请      | `TY-`        | `TY-2026-0001`        |
| TY 团支部    | `TY-BR-`     | `TY-BR-2026-0001`     |
| ST 社团      | `ST-`        | `ST-2026-0001`        |
| ST 活动      | `ST-ACT-`    | `ST-ACT-2026-0001`    |
| SQ 事件      | `SQ-`        | `SQ-2026-0001`        |
| QG 困难认定  | `QG-DIF-`    | `QG-DIF-2026-0001`    |
| QG 岗位      | `QG-POS-`    | `QG-POS-2026-0001`    |
| QG 薪酬      | `QG-PAY-`    | `QG-PAY-202605-0001`  |

### 13.4 限流策略

| 端点类型               | 默认配额                |
| ---------------------- | ----------------------- |
| 登录 / 验证码          | 10 次 / 5 分钟 / IP     |
| 普通查询               | 600 次 / 分钟 / 用户    |
| 写操作（POST/PUT/DEL） | 120 次 / 分钟 / 用户    |
| 文件上传 token         | 30 次 / 分钟 / 用户     |
| SSE 流                 | 单用户单连接            |

超限返回 `429 / 42900`，响应头 `Retry-After` 给重试间隔。

### 13.5 接口清单速查

| 模块 | 端点数（约） | 关键状态机                 |
| ---- | ------------ | -------------------------- |
| auth | 8            | -                          |
| sys  | 22           | 用户/角色 CRUD             |
| files / notifications / event-logs | 14 | -        |
| ty   | 38           | S0–S4 + 转正决议           |
| st   | 42           | 社团 5 态、活动 5 态、评优 |
| sq   | 36           | 事件 4 等级、职务 6 态     |
| qg   | 40           | 申请 5 阶段、薪酬 4 态     |
| cmp  | 8            | 规则版本激活               |
| idx  | 6            | 学生画像聚合               |
| **合计** | **≈ 214**| -                          |

### 13.6 测试 / Mock 建议

1. 后端单元测试：每个端点至少覆盖 4 个用例（成功 / 鉴权失败 / 校验失败 / 业务规则失败）。
2. 集成测试：使用 `api-test.sh`（详见 api-dev skill）跑全模块 smoke。
3. 前端 Mock：基于本文档生成的 OpenAPI 用 Prism 启动；CI 中以 schema diff 防回退。
4. 契约测试：Pact / Spring Cloud Contract 二选一，前后端各持一份。

### 13.7 待澄清事项（与 PRD §11 / DB §14.4 对齐）

- **校级学生会活动**：建议在 `POST /st/activities` body 增加 `scope: 'college' | 'school'` 字段，V1.1 兼容。
- **同岗不同等级薪酬**：`POST /qg/positions` body 增加 `hourly_rates: [{level: hard, hourly_rate_cents: 1800}, ...]`。
- **CMP 公式动态**：通过 `/cmp/rule-versions` 每学年一个新版本，旧版本保留只读。

---

**— 文档结束 —**
