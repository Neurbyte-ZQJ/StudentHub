import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useMenuStore } from '@/stores/menu'

// 静态路由
const staticRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', guest: true }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '工作台', requiresAuth: true }
      },
      {
        path: 'sys/dict',
        name: 'DictManage',
        component: () => import('@/views/sys/DictManage.vue'),
        meta: { title: '字典管理', requiresAuth: true }
      },
      // TY 入团申请页面（非菜单路由，由列表页跳转）
      {
        path: 'ty/application/new',
        name: 'TyApplicationCreate',
        component: () => import('@/views/ty/ApplicationForm.vue'),
        meta: { title: '新增入团申请', requiresAuth: true }
      },
      {
        path: 'ty/application/:id/edit',
        name: 'TyApplicationEdit',
        component: () => import('@/views/ty/ApplicationForm.vue'),
        meta: { title: '编辑入团申请', requiresAuth: true }
      },
      {
        path: 'ty/application/:id',
        name: 'TyApplicationDetail',
        component: () => import('@/views/ty/ApplicationDetail.vue'),
        meta: { title: '入团申请详情', requiresAuth: true }
      },
      // ST 社团活动页面
      {
        path: 'st/association/new',
        name: 'StAssociationCreate',
        component: () => import('@/views/st/AssociationForm.vue'),
        meta: { title: '新建社团', requiresAuth: true }
      },
      {
        path: 'st/association/:id/edit',
        name: 'StAssociationEdit',
        component: () => import('@/views/st/AssociationForm.vue'),
        meta: { title: '编辑社团', requiresAuth: true }
      },
      {
        path: 'st/association/:id',
        name: 'StAssociationDetail',
        component: () => import('@/views/st/AssociationDetail.vue'),
        meta: { title: '社团详情', requiresAuth: true }
      },
      {
        path: 'st/activity/new',
        name: 'StActivityCreate',
        component: () => import('@/views/st/ActivityForm.vue'),
        meta: { title: '新建活动', requiresAuth: true }
      },
      {
        path: 'st/activity/:id/edit',
        name: 'StActivityEdit',
        component: () => import('@/views/st/ActivityForm.vue'),
        meta: { title: '编辑活动', requiresAuth: true }
      },
      {
        path: 'st/activity/approval',
        name: 'StActivityApproval',
        component: () => import('@/views/st/ActivityApproval.vue'),
        meta: { title: '活动审批', requiresAuth: true }
      },
      {
        path: 'st/activity/:id',
        name: 'StActivityDetail',
        component: () => import('@/views/st/ActivityDetail.vue'),
        meta: { title: '活动详情', requiresAuth: true }
      },
      {
        path: 'st/activity/:id/checkin',
        name: 'StActivityCheckin',
        component: () => import('@/views/st/ActivityCheckin.vue'),
        meta: { title: '活动签到', requiresAuth: true }
      },
      {
        path: 'st/activity/:id/summary',
        name: 'StActivitySummary',
        component: () => import('@/views/st/ActivitySummary.vue'),
        meta: { title: '活动总结', requiresAuth: true }
      },
      // 通知中心
      {
        path: 'notifications',
        name: 'NotificationCenter',
        component: () => import('@/views/notifications/NotificationCenter.vue'),
        meta: { title: '通知中心', requiresAuth: true }
      },
      // SQ 学生社区页面
      {
        path: 'sq/inspection/new',
        name: 'SqInspectionCreate',
        component: () => import('@/views/sq/InspectionForm.vue'),
        meta: { title: '新增巡查', requiresAuth: true }
      },
      {
        path: 'sq/inspection/:id',
        name: 'SqInspectionDetail',
        component: () => import('@/views/sq/InspectionDetail.vue'),
        meta: { title: '巡查详情', requiresAuth: true }
      },
      {
        path: 'sq/incident/new',
        name: 'SqIncidentCreate',
        component: () => import('@/views/sq/IncidentReport.vue'),
        meta: { title: '上报事件', requiresAuth: true }
      },
      {
        path: 'sq/incident/:id',
        name: 'SqIncidentDetail',
        component: () => import('@/views/sq/IncidentDetail.vue'),
        meta: { title: '事件详情', requiresAuth: true }
      },
      // CMP 综合看板兜底路由（菜单 store 也会动态注册；此处避免菜单 API 失败时无法访问）
      {
        path: 'cmp/dashboard',
        name: 'CmpDashboard',
        component: () => import('@/views/cmp/Dashboard.vue'),
        meta: { title: '管理驾驶舱', requiresAuth: true }
      },
      {
        path: 'cmp/ranking',
        name: 'CmpRanking',
        component: () => import('@/views/cmp/ScoreRanking.vue'),
        meta: { title: '综合分排行', requiresAuth: true }
      },
      {
        path: 'mine/score',
        name: 'MyScore',
        component: () => import('@/views/cmp/MyScore.vue'),
        meta: { title: '我的综合分', requiresAuth: true }
      },
      // ===== 菜单页面静态兜底路由（防刷新时动态路由未加载导致 404）=====
      // TY 团员发展
      { path: 'ty', name: 'TyHome', component: () => import('@/views/ty/ApplicationList.vue'), meta: { title: '团员发展', requiresAuth: true } },
      { path: 'ty/application', name: 'TyApplication', component: () => import('@/views/ty/ApplicationList.vue'), meta: { title: '入团申请', requiresAuth: true } },
      { path: 'ty/approval', name: 'TyApproval', component: () => import('@/views/ty/ApprovalCenter.vue'), meta: { title: '审批中心', requiresAuth: true } },
      // 推优大会
      { path: 'ty/recommendation-meeting/new', name: 'TyRecMeetingCreate', component: () => import('@/views/ty/RecommendationMeetingForm.vue'), meta: { title: '新建推优大会', requiresAuth: true } },
      { path: 'ty/recommendation-meeting', name: 'TyRecMeetingList', component: () => import('@/views/ty/RecommendationMeetingList.vue'), meta: { title: '推优大会', requiresAuth: true } },
      // 培养记录
      { path: 'ty/cultivation', name: 'TyCultivation', component: () => import('@/views/ty/CultivationView.vue'), meta: { title: '培养记录管理', requiresAuth: true } },
      // 发展对象
      { path: 'ty/development-object', name: 'TyDevObject', component: () => import('@/views/ty/DevelopmentObjectView.vue'), meta: { title: '发展对象管理', requiresAuth: true } },
      // 政审
      { path: 'ty/political-review', name: 'TyPoliticalReview', component: () => import('@/views/ty/PoliticalReviewView.vue'), meta: { title: '政审管理', requiresAuth: true } },
      // 发展大会
      { path: 'ty/development-meeting', name: 'TyDevMeeting', component: () => import('@/views/ty/DevelopmentMeetingView.vue'), meta: { title: '发展大会', requiresAuth: true } },
      // 转正流程
      { path: 'ty/probationary', name: 'TyProbationary', component: () => import('@/views/ty/ProbationaryView.vue'), meta: { title: '转正流程', requiresAuth: true } },
      // 团员花名册
      { path: 'ty/member-roster', name: 'TyMemberRoster', component: () => import('@/views/ty/MemberRoster.vue'), meta: { title: '团员花名册', requiresAuth: true } },
      // 团员发展轨迹
      { path: 'ty/students/:id/development-track', name: 'TyDevelopmentTrack', component: () => import('@/views/ty/DevelopmentTrackView.vue'), meta: { title: '发展轨迹', requiresAuth: true } },
      // ST 社团活动
      { path: 'st', name: 'StHome', component: () => import('@/views/st/AssociationList.vue'), meta: { title: '社团活动', requiresAuth: true } },
      { path: 'st/association', name: 'StAssociation', component: () => import('@/views/st/AssociationList.vue'), meta: { title: '社团管理', requiresAuth: true } },
      { path: 'st/activity', name: 'StActivity', component: () => import('@/views/st/ActivityList.vue'), meta: { title: '活动管理', requiresAuth: true } },
      // SQ 学生社区
      { path: 'sq', name: 'SqHome', component: () => import('@/views/sq/InspectionList.vue'), meta: { title: '学生社区', requiresAuth: true } },
      { path: 'sq/building', name: 'SqBuilding', component: () => import('@/views/sq/BuildingTree.vue'), meta: { title: '楼栋管理', requiresAuth: true } },
      { path: 'sq/inspection', name: 'SqInspection', component: () => import('@/views/sq/InspectionList.vue'), meta: { title: '巡查记录', requiresAuth: true } },
      { path: 'sq/incident', name: 'SqIncident', component: () => import('@/views/sq/IncidentList.vue'), meta: { title: '异常事件', requiresAuth: true } },
      // QG 勤工助学
      { path: 'qg', name: 'QgHome', component: () => import('@/views/qg/PositionList.vue'), meta: { title: '勤工助学', requiresAuth: true } },
      { path: 'qg/difficulty', name: 'QgDifficulty', component: () => import('@/views/qg/DifficultyList.vue'), meta: { title: '困难认定', requiresAuth: true } },
      { path: 'qg/position', name: 'QgPosition', component: () => import('@/views/qg/PositionList.vue'), meta: { title: '岗位管理', requiresAuth: true } },
      { path: 'qg/attendance', name: 'QgAttendance', component: () => import('@/views/qg/AttendanceRecord.vue'), meta: { title: '工时打卡', requiresAuth: true } },
      // 我的申请
      { path: 'mine', name: 'MineHome', component: () => import('@/views/cmp/MyScore.vue'), meta: { title: '我的申请', requiresAuth: true } },
      { path: 'mine/ty-development', name: 'MineTyDevelopment', component: () => import('@/views/ty/MyDevelopment.vue'), meta: { title: '我的团员发展', requiresAuth: true } },
      { path: 'mine/ty-application', name: 'MineTyApplication', component: () => import('@/views/ty/ApplicationList.vue'), meta: { title: '我的入团申请', requiresAuth: true } },
      { path: 'mine/thought-report', name: 'MineThoughtReport', component: () => import('@/views/ty/MyThoughtReport.vue'), meta: { title: '我的思想汇报', requiresAuth: true } },
      { path: 'mine/activity', name: 'MineActivity', component: () => import('@/views/st/ActivityList.vue'), meta: { title: '我的社团', requiresAuth: true } },
      { path: 'mine/work', name: 'MineWork', component: () => import('@/views/qg/AttendanceRecord.vue'), meta: { title: '我的勤工', requiresAuth: true } },
      { path: 'mine/profile', name: 'MineProfile', component: () => import('@/views/idx/MyProfile.vue'), meta: { title: '我的档案', requiresAuth: true } },
      // IDX 学生管理
      { path: 'idx/student', name: 'IdxStudent', component: () => import('@/views/idx/StudentList.vue'), meta: { title: '学生列表', requiresAuth: true } },
      { path: 'idx/import', name: 'IdxImport', component: () => import('@/views/idx/StudentImport.vue'), meta: { title: '学生导入', requiresAuth: true } },
      // SYS 系统管理
      { path: 'sys/user', name: 'SysUser', component: () => import('@/views/sys/UserManage.vue'), meta: { title: '用户管理', requiresAuth: true } },
      { path: 'sys/org', name: 'SysOrg', component: () => import('@/views/sys/OrgManage.vue'), meta: { title: '组织管理', requiresAuth: true } },
      { path: 'sys/job', name: 'SysJob', component: () => import('@/views/sys/JobMonitor.vue'), meta: { title: '任务监控', requiresAuth: true } },
      {
        path: '403',
        name: 'Forbidden',
        component: () => import('@/views/Forbidden.vue'),
        meta: { title: '无权限' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '页面不存在' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: staticRoutes
})

// 路由守卫：未登录 → /login，已登录 → 禁止访问 /login
// 登录后首次加载菜单 → 拉取 → 注册动态路由 → next()
router.beforeEach(async (to, _from, next) => {
  const token = localStorage.getItem('access_token')

  // 未登录访问需认证页 → 跳转登录
  if (to.matched.some(r => r.meta.requiresAuth) && !token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }

  // 已登录访问 /login → 跳转首页
  if (to.matched.some(r => r.meta.guest) && token) {
    next({ name: 'Dashboard' })
    return
  }

  // 已登录：确保菜单和用户信息已加载
  if (token) {
    const menuStore = useMenuStore()
    const authStore = useAuthStore()
    if (!menuStore.isLoaded) {
      try {
        await menuStore.fetchMenus(router)
      } catch {
        // 菜单加载失败不阻塞
      }
    }
    // 用户信息未恢复时主动拉取（解决刷新页面后 user 为 null 导致显示默认"用户"的问题）
    if (!authStore.user) {
      try {
        await authStore.fetchUser()
      } catch {
        // 用户信息拉取失败不阻塞导航
      }
    }
  }

  next()
})

export default router
