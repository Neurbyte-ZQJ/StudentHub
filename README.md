# StudentHub · 学生"一站式"自主管理过程管理系统

> 一个面向高校"第二课堂"与学生事务管理的统一管理平台，围绕 **学生主体 + 过程档案 + 时间戳** 沉淀数据，覆盖 **团员发展、社团活动、学生社区与自治队伍、勤工助学** 四大核心模块，最终形成可量化的综合素质档案。
>
> 🤖 **AI 助手请优先阅读** [AGENTS.md](file:///d:/Teach/AI_Coding/StudentHub/AGENTS.md)，再按需打开 [docs/](file:///d:/Teach/AI_Coding/StudentHub/docs) 下的"宪法级"文档。

![status](https://img.shields.io/badge/status-active-blue)
![backend](https://img.shields.io/badge/backend-Go%201.25-00ADD8)
![frontend](https://img.shields.io/badge/frontend-Vue%203%20%2B%20Vite5-42b883)
![db](https://img.shields.io/badge/db-SQLite3-003B57)
![license](https://img.shields.io/badge/license-Internal-lightgrey)

---

## 目录

- [1. 项目简介](#1-项目简介)
  - [1.1 业务背景](#11-业务背景)
  - [1.2 业务目标](#12-业务目标)
  - [1.3 核心特色](#13-核心特色)
- [2. 技术栈](#2-技术栈)
- [3. 业务模块速览](#3-业务模块速览)
- [4. 系统架构](#4-系统架构)
- [5. 目录结构](#5-目录结构)
- [6. 快速开始](#6-快速开始)
  - [6.1 环境要求](#61-环境要求)
  - [6.2 克隆代码](#62-克隆代码)
  - [6.3 启动后端](#63-启动后端)
  - [6.4 启动前端](#64-启动前端)
  - [6.5 首次访问与默认账号](#65-首次访问与默认账号)
- [7. 常用命令](#7-常用命令)
- [8. 配置说明](#8-配置说明)
- [9. API 约定](#9-api-约定)
  - [9.1 统一响应封包](#91-统一响应封包)
  - [9.2 鉴权流程](#92-鉴权流程)
  - [9.3 健康检查](#93-健康检查)
  - [9.4 错误码](#94-错误码)
- [10. 数据库](#10-数据库)
- [11. 开发指南](#11-开发指南)
  - [11.1 垂直击穿](#111-垂直击穿)
  - [11.2 后端目录铁律](#112-后端目录铁律)
  - [11.3 前端目录铁律](#113-前端目录铁律)
  - [11.4 提交前自检](#114-提交前自检)
- [12. 生产部署](#12-生产部署)
  - [12.1 后端部署](#121-后端部署)
  - [12.2 前端部署](#122-前端部署)
  - [12.3 Nginx 反向代理示例](#123-nginx-反向代理示例)
- [13. 常见问题（FAQ）](#13-常见问题faq)
- [14. 文档导航](#14-文档导航)
- [15. 许可证](#15-许可证)

---

## 1. 项目简介

### 1.1 业务背景

高校"第二课堂"与学生事务管理长期存在三类痛点：

1. **过程数据散落**：团员发展材料、社团活动台账、自治值班记录、勤工工时统计分散在 Excel、微信群、纸质签字本中，过程留痕困难。
2. **规则执行不严**：推优大会到会率、表决通过率、岗位工时上限、考核合格线等业务硬性规则缺乏系统级卡控，存在合规风险。
3. **画像割裂**：学生在不同组织中的表现（团员先进性、社团贡献度、自治履职度、勤工履职度）相互孤立，无法形成统一的综合素质档案。

StudentHub 致力于构建 **"一个入口、一套身份、一条主线"** 的一站式管理平台。

### 1.2 业务目标

| O  | 目标       | KR 关键结果                                                                                |
|----|------------|--------------------------------------------------------------------------------------------|
| O1 | 过程管理合规化 | 团员发展 5 节点 100% 留痕；社团活动立项审批线上闭环率 ≥ 99%                              |
| O2 | 自治与安全提效 | 社区异常事件平均响应时长缩短 50%；自治队伍考核线上化率 100%                              |
| O3 | 勤工助学精准化 | 工时异常（>40h/月）系统卡控率 100%；困难生岗位覆盖率 ≥ 95%                               |
| O4 | 画像与决策数据化 | 学生过程档案字段完整率 ≥ 98%；综合素质量化分数为评奖评优提供唯一数据源                  |

### 1.3 核心特色

- **统一身份**：以"学生组织身份"为中心，绑定团籍、社团成员、社区楼层长、勤工岗位等从属关系。
- **过程留痕**：所有事件围绕"学生主体 + 过程档案 + 时间戳"沉淀，重要节点 100% 可追溯。
- **规则卡控**：将业务硬性规则（推优到会率、工时上限、考核合格线）落到 Service 层系统级卡控。
- **量化画像**：CMP 模块基于事件溯源对四大模块贡献做量化计算，形成综合素质分数。
- **权限矩阵**：RBAC + ABAC 混合模型，按角色编码（如 `R-SY-ADMIN`、`R-COL-COUN`、`R-STU-NORM`）+ 业务属性双重控制。

---

## 2. 技术栈

### 2.1 后端

| 类别     | 选型 |
|----------|------|
| 语言     | Go 1.25 |
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) v1.10 |
| ORM      | [GORM](https://gorm.io) v1.25 + `gorm.io/driver/sqlite` |
| 数据库   | SQLite 3.45（启用 WAL / foreign_keys / busy_timeout） |
| 鉴权     | JWT（access 15min + refresh 7d via HttpOnly Cookie） + `pkg/revokex` 黑名单 |
| 调度     | robfig/cron v3（`internal/scheduler`） |
| 日志     | Uber Zap 1.27 |
| 事件     | 进程内 EventBus（`internal/eventx`） |
| 状态机   | `internal/statem` 通用引擎（ty/st/qg 复用） |
| 静态检查 | [golangci-lint](https://golangci-lint.run)（[`.golangci.yml`](file:///d:/Teach/AI_Coding/StudentHub/backend/.golangci.yml)，启用 govet/staticcheck/errcheck/ineffassign/gofmt/revive/gocyclo≤20/gocognit≤25/misspell/bodyclose/sqlclosecheck/prealloc） |
| 容器化   | Dockerfile（多阶段：Node 20 → Go 1.25 → Alpine 3.20）+ docker-compose.yml |
| 业务编号 | `internal/idgen` 自研雪花 + Redis-less 序列 |

### 2.2 前端

| 类别     | 选型 |
|----------|------|
| 框架     | Vue 3.5（`<script setup>` + Composition API） |
| 构建     | Vite 5 |
| UI 库    | Element Plus 2.8 |
| 路由     | Vue Router 4 |
| 状态     | Pinia 3 |
| HTTP     | Axios 1.7（统一拦截器 + 401 自动刷新 + 排队重试） |
| 可视化   | ECharts 5.6 |
| 包管理   | pnpm（推荐）/ npm |

### 2.3 工程化

- **后端模块化**：`cmd/ + internal/modules/<name>/{api,service,repository,model,event,statemachine}` + `pkg/*`
- **统一 API 前缀**：`/api/v1/<module>`
- **统一响应封包**：`{code, message, data, request_id}`
- **自动化迁移**：GORM `AutoMigrate` 在启动时建表
- **启动种子**：`internal/boot/seed.go` 注入管理员、角色、字典、菜单、院系、学生、团支部、审批用户、演示业务数据

---

## 3. 业务模块速览

后端按业务域拆分为 11 个模块（[`backend/internal/modules`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/modules)），前端按相同前缀组织 [view 目录](file:///d:/Teach/AI_Coding/StudentHub/frontend/src/views)：

| 模块 | 编码 | 业务范围 | 主要子流程 |
| --- | --- | --- | --- |
| 鉴权 | `auth` | 登录、Token 刷新、当前用户 | 公开接口 + JWT 中间件 |
| 系统 | `sys` | 字典、菜单、组织、用户 | R-SY-ADMIN 受限 |
| 学生身份 | `idx` | 学生主数据、个人档案、组织树 | 4 大模块共同依赖 |
| 团员发展 | `ty` | 入团申请→推优大会→培养考察→发展对象→政审→发展大会→转正 | 7 步状态机 |
| 社团活动 | `st` | 社团、成员、活动立项、签到、汇总 | A/B/C/D 级动态审批链 |
| 学生社区 | `sq` | 楼栋/寝室/床位/巡查/事件 | 三级组织 + 5 级巡查 + L1–L4 事件 |
| 勤工助学 | `qg` | 困难认定/岗位/考勤/考核 | 月工时 40h 上限卡控 |
| 综合素质量化 | `cmp` | 量化分计算、规则版本、Dashboard | 订阅 TY/ST/SQ/QG 事件 |
| 通知 | `noti` | 站内信、铃铛中心 | 4 通道（钉钉/企微/短信/邮件） |
| 文件 | `file` | 异步归档、PDF/图片附件 | 本地存储 `storage/YYYY/MM/` |
| 工作台 | `dashboard` | 总览、Top 榜、待办聚合 | 所有登录用户可访问 |
| 定时任务 | `sys/jobs` | cron 任务监控 + 手动触发 + 执行日志 | 前端 `JobMonitor.vue` 配套 `api/job.js` |

---

## 4. 系统架构

```text
┌──────────────────────────────────────────────────────────────┐
│  Browser (Vue3 SPA + Element Plus + Pinia + ECharts)         │
│  http://127.0.0.1:5173  (dev) /  http://<host>/  (prod)     │
└───────────────────────────┬──────────────────────────────────┘
                            │  /api/v1/**  (Axios, JSON)
                            ▼
┌──────────────────────────────────────────────────────────────┐
│  Vite Dev Proxy        /         Nginx Reverse Proxy (prod)  │
│  127.0.0.1:5173        /         :80 / :443                  │
│  proxy /api -> :8080   /         location /api -> :8080      │
└───────────────────────────┬──────────────────────────────────┘
                            ▼
┌──────────────────────────────────────────────────────────────┐
│  Gin HTTP Server (Go 1.25)               :8080               │
│  ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌──────────────┐    │
│  │  Auth   │  │   Sys    │  │  IDX    │  │   Modules    │    │
│  │ /auth/* │  │  /sys/*  │  │  /idx/* │  │ ty/st/sq/qg  │    │
│  └────┬────┘  └────┬─────┘  └────┬────┘  └──────┬───────┘    │
│       └──────────┬─┴─────────────┴──────────────┘            │
│                  ▼                                            │
│         Service / Repository / StateMachine                  │
│                  │                                            │
│                  ▼                                            │
│  ┌────────┐  ┌────────┐  ┌──────────┐  ┌────────────────┐    │
│  │  CMP   │  │  Noti  │  │  File    │  │  Event Bus      │   │
│  │ 量化   │  │ 通知   │  │ 文件归档 │  │ (eventx)        │   │
│  └────────┘  └────────┘  └──────────┘  └────────────────┘    │
│                  │                                            │
│                  ▼                                            │
│       SQLite3 (WAL)  +  Local Storage  +  Scheduler          │
│       data/studenthub.db  storage/YYYY/MM  robfig/cron        │
└──────────────────────────────────────────────────────────────┘
```

---

## 5. 目录结构

```text
StudentHub/
├── backend/                                # Go 后端工程
│   ├── cmd/                                # 可执行入口
│   │   ├── server/main.go                  # 主服务（API + 调度 + 启动种子）
│   │   ├── seedall/                        # ★ 一站式灌入 TY/ST/SQ/QG 演示数据
│   │   │   ├── main.go
│   │   │   ├── seed_ty.go / seed_st.go / seed_sq.go / seed_qg.go
│   │   ├── seedstudent/main.go             # 旧版：为学生批量开通 student{NN} 测试账号（已被 SeedStudentUser 回填为学号）
│   │   ├── seedst/main.go                  # 社团活动种子（已被 seedall 涵盖）
│   │   ├── testseed/main.go                # 测试种子
│   │   ├── dbcheck/main.go                 # 数据库自检（表结构 / 索引 / 软删）
│   │   ├── dbcheck_encoding/main.go        # 中文/全角标点乱码自检
│   │   ├── dbcheckmenu/main.go             # sys_menu 孤儿 / 重名排查（--title / --cleanup / --code）
│   │   ├── renamedict/main.go              # 字典项重命名工具
│   │   ├── fixactivity/main.go             # 活动数据修复工具
│   │   └── fixinspbizno/main.go            # 巡查业务编号（sq_inspection.biz_no）修复
│   ├── configs/
│   │   └── config.yaml                     # 应用配置（端口 / DB / JWT / Log）
│   ├── data/                               # SQLite 数据库（自动生成）
│   │   └── studenthub.db (+ -wal, -shm)
│   ├── internal/
│   │   ├── boot/                           # 启动装配：配置/日志/DB/路由
│   │   │   ├── boot.go                     # 主装配入口
│   │   │   ├── middleware.go               # CORS / RequestID / UTF-8 守卫
│   │   │   └── seed.go                     # 启动种子（admin/角色/字典/演示数据）
│   │   ├── eventx/bus.go                   # 进程内事件总线（CMP 规则→得分解耦）
│   │   ├── idgen/biz_no.go                 # 业务编号生成器（sq_inspection.biz_no 等）
│   │   ├── middleware/                     # JWT / RBAC 中间件
│   │   │   ├── auth_middleware.go
│   │   │   └── rbac_middleware.go
│   │   ├── models/                         # GORM 模型（按模块拆文件）
│   │   │   ├── foundation_idx.go
│   │   │   ├── foundation_sys.go
│   │   │   ├── models.go
│   │   │   ├── module_cmp.go
│   │   │   ├── module_qg.go
│   │   │   ├── module_sq.go
│   │   │   ├── module_st.go
│   │   │   └── module_ty.go
│   │   ├── notifyx/channels/               # 多通道推送：dingtalk / email / sms / wecom
│   │   ├── scheduler/                      # cron 调度（sys_jobs）
│   │   ├── statem/state_machine.go         # 通用状态机引擎（ty/st/qg 复用）
│   │   ├── modules/                        # 业务模块（详见第 3 节）
│   │   │   ├── auth/{api,service,jwt,repository}/
│   │   │   ├── cmp/{api,service,repository,event}/
│   │   │   ├── dashboard/{api,service}/
│   │   │   ├── file/{api,service,repository}/
│   │   │   ├── idx/{api,service,repository}/
│   │   │   ├── noti/{api,service,repository}/
│   │   │   ├── qg/{api,service,repository,statemachine}/
│   │   │   │   └── statemachine/assess_sm.go            # 勤工考核状态机
│   │   │   ├── sq/{api,service,repository}/
│   │   │   ├── st/{api,service,repository,statemachine}/
│   │   │   │   ├── api/recruit_handler.go               # ★ 招新 API
│   │   │   │   ├── service/recruit_service.go
│   │   │   │   ├── repository/recruit_repository.go
│   │   │   │   └── statemachine/{activity_sm,recruit_plan_sm}.go
│   │   │   ├── sys/{api,service,repository}/
│   │   │   └── ty/{api,service,repository,statemachine}/
│   ├── pkg/                                # 通用工具
│   │   ├── cachex/lru.go                   # LRU 缓存（5min TTL, 512）
│   │   ├── cryptox/aes.go                  # AES 加解密
│   │   ├── logger/logger.go                # Zap 工厂
│   │   ├── response/response.go            # 统一响应封包
│   │   └── revokex/revokex.go              # Refresh Token jti 黑名单（ADR-005）
│   ├── storage/                            # 本地文件存储（按 年/月 组织）
│   ├── logs/                               # 运行日志（server-8088.log 等）
│   ├── go.mod / go.sum
│   ├── .golangci.yml                       # 静态检查门禁配置（13 类 linter）
│   ├── .gitignore
│   ├── tmp_probe_paths.ps1                 # 临时调试脚本（开发期路径探测）
│   ├── tmp_verify_st_recruit.ps1           # 临时调试脚本（ST 招新数据校验）
│   └── student-system.exe                  # 已构建的产物（Windows）
│
├── frontend/                               # Vue3 前端工程
│   ├── src/
│   │   ├── api/                            # 按模块拆分的 HTTP 客户端
│   │   │   ├── auth.js  cmp.js  dashboard.js  file.js  http.js
│   │   │   ├── idx.js  job.js  notification.js  qg.js  sq.js
│   │   │   ├── st.js  sys.js  sys-org.js  ty.js
│   │   ├── components/                     # 通用业务组件
│   │   │   ├── ApprovalDialog.vue
│   │   │   ├── ApprovalTimeline.vue
│   │   │   ├── DevelopmentTrack.vue
│   │   │   ├── DictSelect.vue
│   │   │   ├── LevelBadge.vue
│   │   │   └── UploadFile.vue
│   │   ├── layouts/
│   │   │   ├── DefaultLayout.vue
│   │   │   └── components/NotificationBell.vue
│   │   ├── router/index.js                 # 静态路由 + 动态菜单
│   │   ├── stores/                         # Pinia stores
│   │   │   ├── auth.js  dict.js  menu.js
│   │   ├── utils/                          # 工具（日期、ECharts）
│   │   │   ├── datetime.js  echarts.js
│   │   │   └── __tests__/datetime.spec.js  # Vitest 单元测试样例
│   │   ├── views/                          # 业务页面（与后端模块一一对应）
│   │   │   ├── Dashboard.vue  Login.vue  Forbidden.vue  NotFound.vue
│   │   │   ├── cmp/      Dashboard.vue / MyScore.vue / ScoreRanking.vue
│   │   │   ├── idx/      MyProfile.vue / StudentList.vue / StudentImport.vue
│   │   │   ├── qg/       PositionList.vue / DifficultyList.vue / AttendanceRecord.vue
│   │   │   ├── sq/       BuildingTree.vue / InspectionList.vue
│   │   │   │            InspectionForm.vue / InspectionDetail.vue
│   │   │   │            IncidentList.vue / IncidentDetail.vue / IncidentReport.vue
│   │   │   ├── st/       AssociationList/Detail/Form.vue
│   │   │   │            ActivityList/Detail/Form/Approval/Checkin/Summary.vue
│   │   │   │            RecruitPlanList/Detail/Form.vue
│   │   │   │            RecruitApplyList.vue                # ★ 招新申请（学生端）
│   │   │   ├── sys/      DictManage.vue / OrgManage.vue / UserManage.vue / JobMonitor.vue
│   │   │   ├── ty/       ApplicationList/Detail/Form.vue / ApprovalCenter.vue
│   │   │   │            RecommendationMeetingList/Form.vue
│   │   │   │            CultivationView.vue / DevelopmentObjectView.vue
│   │   │   │            PoliticalReviewView.vue / DevelopmentMeetingView.vue
│   │   │   │            DevelopmentTrackView.vue / ProbationaryView.vue
│   │   │   │            MemberRoster.vue / MyThoughtReport.vue
│   │   │   │            MyDevelopment.vue    # ★ 学生自助「我的团员发展」
│   │   │   └── notifications/ NotificationCenter.vue
│   │   ├── App.vue
│   │   └── main.js
│   ├── public/images/                      # 静态资源（logo / img01-02）
│   ├── dist/                               # 构建产物（pnpm build 产出）
│   ├── .eslintrc.cjs                       # ESLint 规则（vue / @vue/eslint-config-prettier）
│   ├── .prettierrc.json / .prettierignore  # Prettier 格式化
│   ├── .gitignore
│   ├── index.html                          # Vite 入口 HTML 模板
│   ├── package.json                        # 依赖与脚本（dev/build/lint/test/format）
│   ├── pnpm-lock.yaml / package-lock.json  # 锁定文件
│   ├── pnpm-workspace.yaml                 # pnpm 协议区（wkspc 协议声明）
│   ├── vite.config.js                      # Vite + 代理（/api -> :8080，strictPort；按需调整）
│   └── vitest.config.js                    # Vitest 单元测试配置
│
├── .trae/                                  # AI 助手技能（项目级 Skills 库）
│   ├── rules/project_rules.md              # ★ 项目铁律（本 README 同级 SSOT）
│   └── skills/                             # 14 个领域技能卡
│       ├── api-dev/  architecture-designer/  database-designer/  db-schema/
│       ├── encoding-fix-zh/  frontend-design/  frontend-slides/  prd/
│       ├── skill-creator/  spec-workflow-guide/  sql-toolkit/  superpowers/
│       ├── system-design/  ui-ux/  update-docs/  …
│
├── screenshot-tool/                        # 截图自动化工具（Playwright headless）
│   ├── capture.mjs / gen_index.mjs / debug*.mjs
│   └── package.json
│
├── screenshots/                            # 系统截图产物（README 引用源）
│   ├── admin-00-login.png ~ admin-28-notifications.png     # 28 张管理员视角
│   ├── student-00-login.png ~ student-08-mine-profile.png  # 9 张学生视角
│   └── README.md / README-admin.md / README-student.md
│
├── AGENTS.md                               # AI 助手协作入口（Trae/Cursor/Claude Code…）
├── Dockerfile / docker-compose.yml         # 多阶段构建 + 单容器编排
├── .dockerignore  .gitignore
├── docs/                                   # 项目"宪法"（SSOT）
│   ├── 01_PRD.md                           # 产品需求（业务唯一事实来源）
│   ├── 02_ADR.md                           # 架构决策与工程规范
│   ├── 03_database_design_spec.md          # 数据库结构与 GORM 模型规范
│   ├── 04_SRD_api_specifications.md        # API 接口契约
│   ├── 05_superpowers_iteration_plan.md    # 12 步迭代路线图
│   └── 06_StudentHub_roadshow.html         # 路演/汇报版本（独立 HTML）
├── deploy/                                 # 部署资产
│   ├── .env.example  deploy.sh  entrypoint.sh
├── .github/
│   ├── SECRETS.md                          # 凭据清单
│   └── workflows/ci.yml / pr-check.yml / release.yml
│
└── README.md                               # 本文件
```

---

## 6. 快速开始

> **建议环境**：Windows 10/11 + PowerShell（5.1 或 7+）。macOS / Linux 同样支持，命令请按需替换路径分隔符。

### 6.1 环境要求

| 工具 | 最低版本 | 用途 | 备注 |
| --- | --- | --- | --- |
| Go | **1.25.0** | 编译/运行后端 | `go version` |
| Node.js | **18.x** | 构建/运行前端 | `node -v` |
| pnpm（推荐）| **9.x** | 前端包管理 | `npm i -g pnpm` |
| C 编译器 | — | `mattn/go-sqlite3` CGO 依赖 | Windows 需安装 [MinGW-w64](https://www.mingw-w64.org/) / TDM-GCC；macOS 自带 clang；Linux 安装 `build-essential` |
| Git | 2.x | 版本管理 | — |

> **提示**：Windows 平台若仅运行已构建好的 [`backend/student-system.exe`](file:///d:/Teach/AI_Coding/StudentHub/backend/student-system.exe)，可跳过 Go 与 C 编译器。

### 6.2 克隆代码

```powershell
git clone <repo-url> StudentHub
cd StudentHub
```

### 6.3 启动后端

后端监听端口由 [`defaultConfig()`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/boot.go#L78-L90) 中的 `getEnvInt("APP_PORT", 8080)` 决定（**默认 `:8080`**，与 [ADR](file:///d:/Teach/AI_Coding/StudentHub/docs/02_ADR.md) / `config.yaml` / `vite.config.js` 代理目标完全一致；历史日志中的 `server-8088.log` 等为开发期临时端口调整留下的痕迹，可通过 `APP_PORT` 环境变量任意覆盖）。数据库文件位于 `backend/data/studenthub.db`。

```powershell
# 进入后端目录
cd backend

# 方式一：源码直接运行（首次会拉取依赖，默认监听 :8080）
go mod tidy
go run ./cmd/server

# 方式二：自定义端口启动（与前端 vite.config.js 代理目标需同步）
$Env:APP_PORT = 9080
go run ./cmd/server

# 方式三：编译为可执行文件
go build -o student-system.exe ./cmd/server
.\student-system.exe
```

启动过程会自动完成：

1. 加载默认配置（仅 `APP_ENV` / `APP_PORT` / `JWT_SECRET` 三个环境变量生效；[`configs/config.yaml`](file:///d:/Teach/AI_Coding/StudentHub/backend/configs/config.yaml) 当前仅作归档参考，未被 `boot.Run()` 读取）
2. 初始化 Zap 日志（默认输出到 stderr；如需落盘可结合 `tee` 重定向到 `backend/logs/server-<port>.log`）
3. 打开 SQLite 并启用 WAL / `foreign_keys` / `busy_timeout`
4. `AutoMigrate` 建表（按 [`docs/03_database_design_spec.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md)）
5. 执行启动种子（admin、角色、字典、菜单、院系、学生、团支部、审批用户、SQ/QG/CMP/TY 演示数据）
6. 注册路由、启动 Gin + Scheduler
7. 监听 `:${APP_PORT}`

启动后可通过 `http://localhost:<port>/api/v1/healthz` 验证。

### 6.4 启动前端

前端开发服务器默认监听 `http://127.0.0.1:5173`（`strictPort`，占用即报错），通过 [`vite.config.js`](file:///d:/Teach/AI_Coding/StudentHub/frontend/vite.config.js) 将 `/api/**` 代理到后端 `http://localhost:8080`（**与 `defaultConfig()` 默认端口完全一致，无需任何调整**）。若后端临时切换端口（如 `APP_PORT=9080`），需同步修改 `server.proxy['/api'].target`，否则请求会打到未启动的旧端口。

```powershell
# 进入前端目录
cd frontend

# 安装依赖（推荐 pnpm）
pnpm install
# 或：npm install

# 启动开发服务器
pnpm dev
# 或：npm run dev
```

打开浏览器访问 `http://127.0.0.1:5173`，输入默认账号登录即可（见下节）。

### 6.5 首次访问与默认账号

> **登录账号命名铁律**：所有 `idx_student` 学生的登录 username = **`student_no`（学号）**，共 10 位数字（格式 `{年级4}{专业号2}{班级号2}{序号2}`，例如 `2022010101`），由 [`SeedStudentUser`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/seed.go#L99-L168) 启动时**自动**为每个未绑定的学生创建 sys_user。**无需手动运行任何脚本**。

| 账号 | 密码 | 角色 | 权限范围 |
| --- | --- | --- | --- |
| `admin` | `admin@123` | `R-SY-ADMIN` | 全模块、所有菜单、所有操作 |
| `<学号>`（如 `2023010101`） | `student@123` | `R-STU-NORM` | 普通学生视角（自助） |
| `T001` | `pwd@123` | `R-COL-COUN`（辅导员） | 院系内社团初审、社区/楼层日常管理、勤工岗位初审 |
| `T002` | `pwd@123` | `R-COL-LEAGUE`（院系团委书记）| 院系内团员发展 / 社团活动初审 |
| `T003` | `pwd@123` | `R-SY-LEAGUE`（校团委管理员）| 团员发展、社团活动全局规则与终审 |

> - 实际学号以 `idx_student.student_no` 为准，可在「系统管理 → 组织管理 → 学生」或 DB 中查询。
> - 三级审批教师账号（`T001/T002/T003`）由 [`SeedApprovalUsers`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/seed.go#L171-L245) 启动时自动创建。
> - 旧版 `seedstudent` 脚本（`go run ./cmd/seedstudent`，生成 `student01…NN`）已**被 `SeedStudentUser` 自动回填为学号**（仅在 username 仍为 `student01` 时执行一次），正常情况下不必再手动执行。
> - **生产环境务必**：① 修改 `admin` / 教师 / 学生密码；② 替换 `config.yaml` 中 `jwt.secret`。

---

## 7. 常用命令

### 7.1 后端

| 位置 | 命令 | 说明 |
| --- | --- | --- |
| `backend/` | `go run ./cmd/server` | 启动后端服务（端口由 `APP_PORT` 决定，**默认 `:8080`**） |
| `backend/` | `go build -o student-system.exe ./cmd/server` | 编译为 Windows 可执行文件 |
| `backend/` | `go run ./cmd/seedall` | **推荐**：一站式灌入 TY/ST/SQ/QG 四大业务模块演示数据（仅追加，不清空） |
| `backend/` | `go run ./cmd/seedstudent` | **旧版**：批量生成 `student{NN}` 账号；启动时 `SeedStudentUser` 会自动回填为学号，正常情况下**无需执行** |
| `backend/` | `go run ./cmd/seedst` | 社团活动示例数据（已被 `seedall` 涵盖） |
| `backend/` | `go run ./cmd/testseed` | 测试种子 |
| `backend/` | `go run ./cmd/dbcheck` | 数据库自检（表结构 / 索引 / 软删） |
| `backend/` | `go run ./cmd/dbcheck_encoding` | 中文/全角标点乱码自检（CSV 导出、对账脚本） |
| `backend/` | `go run ./cmd/dbcheckmenu` | sys_menu 孤儿 / 重名排查（`--title "..."` / `--cleanup --code <code>`） |
| `backend/` | `go run ./cmd/renamedict` | 字典项重命名工具 |
| `backend/` | `go run ./cmd/fixactivity` | 活动数据修复工具 |
| `backend/` | `go run ./cmd/fixinspbizno` | 巡查业务编号（`sq_inspection.biz_no`）修复 |
| `backend/` | `go vet ./...` | 静态检查 |
| `backend/` | `go test ./...` | 单元测试（按需） |
| `backend/` | `go mod tidy` | 整理依赖 |

### 7.2 前端

| 位置 | 命令 | 说明 |
| --- | --- | --- |
| `frontend/` | `pnpm dev` | 启动 Vite 开发服务器（`http://127.0.0.1:5173`，`strictPort` 占用即报错） |
| `frontend/` | `pnpm build` | 生产构建到 `frontend/dist/` |
| `frontend/` | `pnpm preview` | 预览生产构建（`http://127.0.0.1:4173`） |
| `frontend/` | `pnpm lint` | ESLint 全量检查（`--max-warnings 0`） |
| `frontend/` | `pnpm lint:fix` | ESLint 自动修复 |
| `frontend/` | `pnpm format` | Prettier 格式化 `src/**/*.{js,vue}` |
| `frontend/` | `pnpm test` | Vitest 单元测试一次性运行 |
| `frontend/` | `pnpm test:watch` | Vitest watch 模式 |
| `frontend/` | `pnpm install` | 安装依赖 |

### 7.3 联调小贴士

- 前端 `vite.config.js` 已将 `/api` 代理到 `:8080`（与后端默认端口完全一致），**开发期请保持后端在运行**。
- 后端 `r.NoRoute` 兜底逻辑：非 `/api` 路径优先返回 `frontend/dist/index.html`（生产部署时生效），方便同一域下提供 SPA + API。
- 浏览器 Network 面板 → 看 `request_id` 可在 `backend/logs/server-8088.log` 关联到对应请求的 Zap 日志。

---

## 8. 配置说明

主配置 [`backend/configs/config.yaml`](file:///d:/Teach/AI_Coding/StudentHub/backend/configs/config.yaml)：

```yaml
app:
  env: dev                  # dev | prod；prod 会切换 Gin 为 ReleaseMode
  port: 8080                # 监听端口（与 defaultConfig 默认一致）
  name: student-system
db:
  path: ./data/studenthub.db
log:
  level: info               # debug | info | warn | error
jwt:
  secret: "studenthub-dev-jwt-secret-change-in-prod"
  access_ttl: 15m           # 访问令牌有效期
  refresh_ttl: 168h         # 刷新令牌有效期（7 天）
  issuer: studenthub
```

> 后端启动时通过 [`boot.defaultConfig()`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/boot.go#L77-L89) 加载，并支持环境变量覆盖（`APP_ENV`、`JWT_SECRET`）。**生产环境部署前请务必修改 `jwt.secret`**。

---

## 9. API 约定

### 9.1 统一响应封包

所有 `/api/v1/**` 接口均返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": { /* 业务对象 */ },
  "request_id": "req-xxxxxx"
}
```

- `code = 0` 表示成功；非 0 为业务错误码（详见 §9.4）。
- 前端 [`http.js`](file:///d:/Teach/AI_Coding/StudentHub/frontend/src/api/http.js) 拦截器自动解包 `data`，失败时 `ElMessage.error(message)`。

### 9.2 鉴权流程

1. **登录** `POST /api/v1/auth/login`（公开）→ 返回 `access_token`（15min），`refresh_token` 通过 `HttpOnly Cookie` 下发（7 天）。
2. 前端将 `access_token` 写入 `localStorage`，请求拦截器自动注入 `Authorization: Bearer <token>`。
3. 收到 **401** 时拦截器调用 `POST /api/v1/auth/refresh` 刷新，**排队重试** 期间其他请求。
4. 刷新失败 → 清理登录态 → 跳转 `/login`。

> 中间件：JWT 解析与验签由 [`internal/middleware/auth_middleware.go`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/middleware/auth_middleware.go) 提供；RBAC 角色卡控见 [`internal/middleware/rbac_middleware.go`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/middleware/rbac_middleware.go)。

### 9.3 健康检查

| 端点 | 用途 |
| --- | --- |
| `GET /api/v1/healthz` | 进程存活 |
| `GET /api/v1/readyz`  | DB 可达性（`sqlDB.Ping()`） |

### 9.4 错误码

| 范围 | 含义 | 示例 |
| --- | --- | --- |
| `0` | 成功 | — |
| `1xxx` | 通用业务错误 | 1001 参数非法、1003 数据不存在 |
| `12xx` | 鉴权 | 1201 未登录、1203 无权限、1204 Token 过期 |
| `13xx` | 业务校验 | 1301 状态机不合法、1302 工时超限、1305 推优到会率不足 |
| `14xx` | 资源/路由 | 1404 接口不存在 |
| `15xx` | 系统/依赖 | 1500 数据库不可用、1599 内部异常 |

> 具体错误码定义以 [`docs/04_SRD_api_specifications.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/04_SRD_api_specifications.md) 为准。

---

## 10. 数据库

- **引擎**：SQLite3，单文件 `backend/data/studenthub.db`。
- **性能开关**（启动时由 [`initDB`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/boot.go#L143-L181) 启用）：

  ```sql
  PRAGMA journal_mode = WAL;     -- 写前日志，提升并发读
  PRAGMA synchronous = NORMAL;   -- 兼顾安全与性能
  PRAGMA foreign_keys = ON;      -- 强制外键
  PRAGMA busy_timeout = 5000;    -- 写锁等待 5s
  PRAGMA temp_store = MEMORY;    -- 临时表放内存
  ```

- **迁移**：GORM `AutoMigrate`（模型见 [`backend/internal/models`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/models)），表结构、字段、类型、长度、索引、外键严格对齐 [`docs/03_database_design_spec.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md)。
- **软删**：约定 `is_deleted` 字段 + 唯一索引 `WHERE is_deleted=0`。
- **备份建议**：
  - 定期复制 `studenthub.db`（+`-wal`、`-shm`）三件套；
  - 仓库已可见 `studenthub.db.bak.YYYYMMDD_HHMMSS` 形式的快照示例。

---

## 11. 开发指南

### 11.1 垂直击穿

按 [`docs/05_superpowers_iteration_plan.md`](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) 的 **12 步** 节奏推进，每个步长内打通 `DB Migration → Model → Repo → Service → API → 前端联调` 全链路，不跨步长、不省略环节。

| 步长 | 名称 | 关键交付 |
| --- | --- | --- |
| S01 | 工程脚手架 + 数据底座 | 后端可启动、DB 自动迁移到位 |
| S02 | 鉴权登录 + JWT + 角色 | 登录闭环、Token 颁发与刷新 |
| S03 | 通用布局 + 路由守卫 + 字典 | 前端骨架、动态菜单、字典下拉 |
| S04 | IDX 学生身份库 + 组织树 | 4 大模块共同依赖 |
| S05–S06 | TY 团员发展 | 入团申请 CRUD + 三级审批流状态机 |
| S07–S08 | ST 社团活动 | 社团 + 活动 CRUD + 分级审批 + 签到 |
| S09 | SQ 学生社区 | 楼栋寝室 + 巡查事件 |
| S10 | QG 勤工助学 | 困难认定 + 岗位 + 工时 |
| S11 | 基础设施 | 异步文件归档 + 通知中心 + 定时任务 |
| S12 | 量化与看板 | CMP 综合素质量化 + 统计报表 |

### 11.2 后端目录铁律

```
backend/
├── cmd/<entry>/main.go                 # 每个可执行入口一个子包
├── internal/
│   ├── boot/                           # 装配（只放配置/路由/中间件）
│   ├── eventx/                         # 进程内事件总线
│   ├── idgen/                          # 业务编号生成器
│   ├── middleware/                     # JWT / RBAC
│   ├── models/                         # 所有 GORM 模型集中管理
│   ├── modules/<name>/                 # 业务模块
│   │   ├── api/        # HTTP Handler（参数绑定 + 调用 service）
│   │   ├── service/    # 业务编排（含状态机/事件发布）
│   │   ├── repository/ # GORM 数据访问
│   │   ├── event/      # 事件订阅（CMP 使用）
│   │   └── statemachine/  # 复杂审批流状态机（ty/st）
│   ├── notifyx/                        # 通知通道实现
│   ├── scheduler/                      # cron 任务
│   └── statem/                         # 通用状态机引擎
├── pkg/                                # 通用工具（无业务依赖）
├── storage/                            # 本地文件存储
├── configs/
├── data/
└── logs/
```

### 11.3 前端目录铁律

```
frontend/src/
├── api/<module>.js                     # 与后端模块一一对应
├── components/                         # 跨模块通用组件
├── layouts/                            # 布局（含顶栏 / 侧栏 / 通知铃铛）
├── router/index.js                     # 静态路由 + 动态菜单挂载
├── stores/                             # Pinia（auth / dict / menu）
├── utils/                              # 工具（日期、ECharts）
└── views/
    ├── <module>/                       # 模块内页面
    └── Dashboard.vue / Login.vue / ...
```

### 11.4 提交前自检

```powershell
# 后端
cd backend
go vet ./...
golangci-lint run ./...         # 完整门禁：vet + staticcheck + errcheck + revive + ...
go build ./...

# 前端
cd ../frontend
pnpm install
pnpm lint            # ESLint --max-warnings 0，零警告才能合入
pnpm build           # 含 Vite 静态检查
pnpm test            # Vitest 单元测试一次性运行
```

---

## 12. 生产部署

项目支持 **两种部署形态**：

| 形态 | 适用场景 | 入口 | 产物 |
| --- | --- | --- | --- |
| **Docker 单容器** | 教学/演示/小规模生产，**默认推荐** | [`docker-compose.yml`](file:///d:/Teach/AI_Coding/StudentHub/docker-compose.yml) | [`Dockerfile`](file:///d:/Teach/AI_Coding/StudentHub/Dockerfile) 多阶段构建（Node 20 + Go 1.25 → Alpine 3.20） |
| **裸金属/虚拟机** | 自定义 K8s / 物理机 / 高级调优 | 自行 systemd + Nginx | `student-system.exe` + `frontend/dist/` |

CI/CD：[`.github/workflows/`](file:///d:/Teach/AI_Coding/StudentHub/.github/workflows) 三套工作流（`ci.yml` → develop 自动部署 dev；`pr-check.yml` → PR 必跑 lint/build/test；`release.yml` → 推送 tag 自动发布），凭据清单见 [`.github/SECRETS.md`](file:///d:/Teach/AI_Coding/StudentHub/.github/SECRETS.md)。

### 12.0 Docker 一键部署（推荐）

```powershell
# 1. 准备环境变量（必须）
Copy-Item deploy/.env.example deploy/.env
notepad deploy/.env   # 至少修改 JWT_SECRET、CRYPTOX_KEY

# 2. 构建并启动
docker compose build
docker compose up -d
docker compose logs -f

# 3. 反向代理层（Nginx / 云 LB）将 :80/:443 转给宿主机 :8080 即可
```

镜像内布局：
- `/app/server` —— 后端二进制
- `/app/frontend/dist/` —— 前端静态资源（Gin NoRoute SPA fallback）
- `/app/data/` —— SQLite + WAL/SHM（持久化卷）
- `/app/storage/` —— 上传文件（持久化卷）
- `/app/logs/` —— 日志输出（持久化卷）

### 12.1 后端部署

```powershell
# 1) 交叉编译（Linux 示例）
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"
go build -o studenthub-server ./cmd/server

# 2) 上传至服务器，配置 systemd（示例 /etc/systemd/system/studenthub.service）
# [Unit]
# Description=StudentHub Backend
# After=network.target
# [Service]
# WorkingDirectory=/opt/studenthub
# ExecStart=/opt/studenthub/studenthub-server
# Restart=always
# Environment=APP_ENV=prod
# Environment=JWT_SECRET=<强随机密钥>
# [Install]
# WantedBy=multi-user.target

sudo systemctl daemon-reload
sudo systemctl enable --now studenthub
```

### 12.2 前端部署

```powershell
cd frontend
pnpm build           # 产物输出到 frontend/dist/
```

将 `frontend/dist/` 内容上传到 Nginx 静态目录（如 `/var/www/studenthub/`），并由 Nginx 托管（见下节）。

### 12.3 Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name studenthub.example.com;

    # 前端静态资源
    root /var/www/studenthub;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反代
    location /api/ {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade           $http_upgrade;
        proxy_set_header   Connection        "upgrade";
    }

    # 上传文件大小限制
    client_max_body_size 50m;
}
```

> 后端 `r.NoRoute` 已对非 `/api` 路径兜底返回 `frontend/dist/index.html`（生产部署时有效），如纯 Nginx 部署可不依赖该兜底。

---

## 13. 常见问题（FAQ）

**Q1：后端启动报 `mattn/go-sqlite3` 需要 CGO。**
A：安装 C 编译器并启用 CGO：
- Windows：安装 MinGW-w64 / TDM-GCC，并将 `gcc.exe` 所在目录加入 `PATH`；
- macOS：`xcode-select --install`；
- Linux：`sudo apt install build-essential`（Debian/Ubuntu）或 `sudo yum groupinstall "Development Tools"`（CentOS）。
若仅运行已构建的 `student-system.exe`，可忽略。

**Q2：前端 `pnpm dev` 后访问 401。**
A：① 确认后端已启动并监听 `:8080`（默认）：`curl http://localhost:8080/api/v1/healthz` 应返回 `{"status":"healthy"}`；② 确认 [`vite.config.js`](file:///d:/Teach/AI_Coding/StudentHub/frontend/vite.config.js) 的 `server.proxy['/api'].target` 与后端实际端口一致（两者默认都是 `:8080`）；③ 用登录接口正确返回的 `access_token` 调用受保护接口（前端已自动注入 `Authorization: Bearer <token>`）。

**Q3：登录成功但部分菜单点击提示 403。**
A：当前账号未授予对应角色。管理员账号拥有全部权限（`R-SY-ADMIN`）；学生账号仅 `R-STU-NORM` 自助权限；三级审批教师账号 `T001/T002/T003` 分别具备院系辅导员 / 院系团委 / 校团委权限。需扩展权限时，进入「系统管理 → 用户管理」调整角色绑定，或直接修改 `sys_user_role` 表。

**Q4：修改了 `config.yaml` 不生效？**
A：[`backend/configs/config.yaml`](file:///d:/Teach/AI_Coding/StudentHub/backend/configs/config.yaml) 的 `port: 8080` 与 [`defaultConfig()`](file:///d:/Teach/AI_Coding/StudentHub/backend/internal/boot/boot.go#L78-L90) 默认值**已对齐**，但**文件本身不会被 `boot.Run()` 读取**，仅作归档参考。所有运行期参数由 `defaultConfig()` 内的默认值 + `APP_ENV` / `APP_PORT` / `JWT_SECRET` 三个环境变量决定。如需扩展配置项，请修改 `boot.go` 并同步本文 §8。

**Q5：数据库被锁住 / 写不进？**
A：检查是否存在残留进程未释放文件锁；删除 `data/studenthub.db-shm` 与 `data/studenthub.db-wal` 后重启（建议先备份三件套再操作）。同时确认 SQLite 已启用 `busy_timeout`（参见 §10）。

**Q6：如何重置数据？**
A：删除 `backend/data/studenthub.db*` 三件套后重启 `go run ./cmd/server`，启动种子会重新注入 admin、字典、菜单、院系学生、演示数据。

**Q7：生产环境 JWT 被泄露怎么办？**
A：① 立即在 `config.yaml` / 环境变量中替换 `jwt.secret` 并重启服务；② 通知所有在线用户重新登录；③ 排查日志中可疑 `request_id` 关联操作（`backend/logs/server-8088.log`）。

**Q8：想加新模块怎么起步？**
A：参照 [docs/05 §0.3 通用约定](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) —— 后端新增 `internal/modules/<name>/{api,service,repository}`，前端新增 `src/api/<name>.js` + `src/views/<name>/`，并在 `internal/models/` 与 `boot.go` 中注册模型 / 路由 / 种子。

---

## 14. 文档导航

| 文档 | 用途 |
| --- | --- |
| [01_PRD.md](file:///d:/Teach/AI_Coding/StudentHub/docs/01_PRD.md) | 产品需求文档（业务唯一事实来源，SSOT） |
| [02_ADR.md](file:///d:/Teach/AI_Coding/StudentHub/docs/02_ADR.md) | 架构决策记录（ADR-001 ~ 020）与工程规范 |
| [03_database_design_spec.md](file:///d:/Teach/AI_Coding/StudentHub/docs/03_database_design_spec.md) | 数据库结构、GORM 模型、索引与约束规范 |
| [04_SRD_api_specifications.md](file:///d:/Teach/AI_Coding/StudentHub/docs/04_SRD_api_specifications.md) | API 接口契约（URL / 参数 / 响应 / 错误码） |
| [05_superpowers_iteration_plan.md](file:///d:/Teach/AI_Coding/StudentHub/docs/05_superpowers_iteration_plan.md) | 12 步垂直击穿迭代路线图 |
| [06_StudentHub_roadshow.html](file:///d:/Teach/AI_Coding/StudentHub/docs/06_StudentHub_roadshow.html) | 路演/汇报版本（独立 HTML，浏览器直开） |

### 14.1 协作与凭据

| 文件 | 用途 |
| --- | --- |
| [AGENTS.md](file:///d:/Teach/AI_Coding/StudentHub/AGENTS.md) | AI 编码助手协作入口（Trae/Cursor/Claude Code/Aider/Continue…） |
| [`.github/SECRETS.md`](file:///d:/Teach/AI_Coding/StudentHub/.github/SECRETS.md) | CI/CD 所需 GitHub Secrets 清单 |

---

## 15. 许可证

本项目为教学/演示用途的内部工程，许可证信息以仓库 `LICENSE` 为准（若未提供，请按内部约定处理）。
