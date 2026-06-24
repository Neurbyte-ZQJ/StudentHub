<template>
  <div class="workspace">
    <!-- ===== 学生视角 ===== -->
    <template v-if="overview.role_scope === 'student'">
      <header class="welcome-strip welcome-student">
        <div class="welcome-left">
          <div class="avatar-ring">
            <el-avatar :size="52" :src="overview.user?.avatar_url">
              {{ (overview.user?.display_name || '?').charAt(0) }}
            </el-avatar>
          </div>
          <div class="welcome-text">
            <h1>{{ greeting }}，{{ overview.user?.display_name || '同学' }}</h1>
            <p class="welcome-sub">
              <span class="id-badge">{{ overview.user?.username }}</span>
              <el-tag
                v-for="role in overview.user?.roles"
                :key="role.code"
                size="small"
                effect="plain"
                round
                class="role-tag"
              >
                {{ role.name }}
              </el-tag>
            </p>
          </div>
        </div>
        <div class="welcome-right">
          <span class="date-text">{{ currentDate }}</span>
        </div>
      </header>

      <!-- 入团进度卡片（学生专属） -->
      <section v-if="tyStatusInfo" class="progress-card" :style="{ '--accent': tyStatusInfo.color }">
        <div class="progress-label">入团发展进度</div>
        <div class="progress-body">
          <span class="progress-status">{{ tyStatusInfo.label }}</span>
          <el-button type="primary" size="small" round @click="$router.push('/mine/ty-application')">
            查看详情
          </el-button>
        </div>
      </section>

      <!-- 指标行：根据数量自适应宽度 -->
      <section class="metrics-row" :style="{ '--n': studentMetrics.length }">
        <div v-for="m in studentMetrics" :key="m.key" class="metric-card" :style="{ '--accent': m.color }">
          <div class="metric-value">{{ formatNum(m.value) }}</div>
          <div class="metric-label">{{ m.label }}</div>
          <div class="metric-bar" :style="{ background: m.color }"></div>
        </div>
      </section>

      <div class="main-grid">
        <section class="panel todo-panel">
          <h2 class="panel-title">我的待办</h2>
          <div v-if="overview.todo_items?.length" class="todo-list">
            <div
              v-for="(item, idx) in overview.todo_items.filter(i => i.path)"
              :key="idx"
              class="todo-item"
              @click="$router.push(item.path)"
            >
              <div class="todo-info">
                <span class="todo-title">{{ item.title }}</span>
              </div>
              <div class="todo-right">
                <span v-if="item.count > 0" class="todo-count" :style="{ background: item.color + '22', color: item.color }">{{ item.count }}</span>
                <el-icon :size="14"><ArrowRight /></el-icon>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">暂无待办事项</div>
        </section>

        <section class="panel links-panel">
          <h2 class="panel-title">常用功能</h2>
          <div class="links-grid">
            <div
              v-for="link in overview.quick_links"
              :key="link.title"
              class="link-card"
              :style="{ '--link-color': link.color }"
              @click="$router.push(link.path)"
            >
              <div class="link-icon-wrap">
                <el-icon :size="24"><component :is="iconMap[link.icon]" /></el-icon>
              </div>
              <span class="link-title">{{ link.title }}</span>
            </div>
          </div>
        </section>
      </div>
    </template>

    <!-- ===== 教师/管理员视角 ===== -->
    <template v-else>
      <header class="welcome-strip welcome-staff">
        <div class="welcome-left">
          <div class="avatar-ring">
            <el-avatar :size="52" :src="overview.user?.avatar_url">
              {{ (overview.user?.display_name || '?').charAt(0) }}
            </el-avatar>
          </div>
          <div class="welcome-text">
            <h1>{{ greeting }}，{{ overview.user?.display_name }}</h1>
            <p class="welcome-sub">
              <span class="id-badge">{{ overview.user?.username }}</span>
              <el-tag
                v-for="role in overview.user?.roles"
                :key="role.code"
                size="small"
                effect="plain"
                round
                class="role-tag"
              >
                {{ role.name }}
              </el-tag>
            </p>
          </div>
        </div>
        <div class="welcome-right">
          <span class="date-text">{{ currentDate }}</span>
        </div>
      </header>

      <!-- 指标行：根据数量自适应宽度 -->
      <section class="metrics-row" :style="{ '--n': staffMetrics.length }">
        <div v-for="m in staffMetrics" :key="m.key" class="metric-card" :style="{ '--accent': m.color }">
          <div class="metric-value">{{ formatNum(m.value) }}</div>
          <div class="metric-label">{{ m.label }}</div>
          <div class="metric-bar" :style="{ background: m.color }"></div>
        </div>
      </section>

      <div class="main-grid">
        <section class="panel todo-panel">
          <h2 class="panel-title">审批与待办</h2>
          <div v-if="overview.todo_items?.length && hasActiveTodos" class="todo-list">
            <div
              v-for="(item, idx) in activeTodos"
              :key="idx"
              class="todo-item"
              @click="$router.push(item.path)"
            >
              <div class="todo-info">
                <span class="todo-dot" :style="{ background: item.color || '#409eff' }"></span>
                <span class="todo-title">{{ item.title }}</span>
              </div>
              <div class="todo-right">
                <span v-if="item.count > 0" class="todo-count" :style="{ background: item.color + '22', color: item.color }">{{ item.count }}</span>
                <el-icon :size="14"><ArrowRight /></el-icon>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">暂无待办事项</div>
        </section>

        <section class="panel links-panel">
          <h2 class="panel-title">管理入口</h2>
          <div class="links-grid">
            <div
              v-for="link in overview.quick_links"
              :key="link.title"
              class="link-card"
              :style="{ '--link-color': link.color }"
              @click="$router.push(link.path)"
            >
              <div class="link-icon-wrap">
                <el-icon :size="24"><component :is="iconMap[link.icon]" /></el-icon>
              </div>
              <span class="link-title">{{ link.title }}</span>
            </div>
          </div>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { dashboardApi } from '@/api/dashboard'
import {
  Flag, Trophy, House, Briefcase, Document, TrendCharts, Setting,
  ArrowRight
} from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const overview = ref({
  user: null,
  role_scope: 'student',
  stats: {},
  todo_items: [],
  quick_links: []
})

const iconMap = { Flag, Trophy, House, Briefcase, Document, TrendCharts, Setting }

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const currentDate = computed(() => {
  const d = new Date()
  const weekdays = ['日', '一', '二', '三', '四', '五', '六']
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 星期${weekdays[d.getDay()]}`
})

// ---- 学生指标 ----
const studentMetrics = computed(() => {
  const s = overview.value.stats || {}
  return [
    { key: 'activity', label: '参加活动', value: s.my_activity_count || 0, color: '#67c23a' },
    { key: 'noti', label: '未读通知', value: s.unread_noti_count || 0, color: '#4a90a4' },
    { key: 'cmp', label: '综合分', value: s.my_cmp_score || 0, color: '#9b59b6' },
    { key: 'recruit', label: '招新中计划', value: s.recruiting_plan_count || 0, color: '#e6a23c' },
  ]
})

// ---- 教师指标 ----
const staffMetrics = computed(() => {
  const s = overview.value.stats || {}
  return [
    { key: 'student', label: '在校学生', value: s.student_count || 0, color: '#3a5ba0' },
    { key: 'ty', label: '入团待审', value: s.ty_pending_count || 0, color: '#c78c46' },
    { key: 'incident', label: '待处理事件', value: s.incident_open_count || 0, color: '#c05050' },
    { key: 'qg', label: '勤工岗位', value: s.qg_position_count || 0, color: '#7a6aae' },
    { key: 'assoc', label: '活跃社团', value: s.active_assoc_count || 0, color: '#5a9a5a' },
    { key: 'noti', label: '未读通知', value: s.unread_noti_count || 0, color: '#4a90a4' },
  ]
})

// ---- 入团状态映射（与后端 TyApplication.Status 短码对齐）----
const tyStatusMap = {
  S1: { label: '申请已提交，等待审核', color: '#e6a23c' },
  S2: { label: '推优通过，进入培养考察', color: '#e6a23c' },
  S3: { label: '培养考察中', color: '#409eff' },
  S4: { label: '已列为发展对象', color: '#67c23a' },
  S5: { label: '政审已完成', color: '#9b59b6' },
  S6: { label: '已被接收为预备团员', color: '#67c23a' },
  S7_MEMBER:       { label: '已是正式团员', color: '#67c23a' },
  // 兼容长码格式
  S1_SUBMITTED:   { label: '申请已提交，等待审核', color: '#e6a23c' },
  S2_RECOMMENDED: { label: '推优通过，进入培养考察', color: '#e6a23c' },
  S3_CULTIVATING: { label: '培养考察中', color: '#409eff' },
  S4_DEVELOPING:  { label: '已列为发展对象', color: '#67c23a' },
  S5_POLITICED:   { label: '政审已完成', color: '#9b59b6' },
  S6_ADMITTED:    { label: '已被接收为预备团员', color: '#67c23a' },
}

const tyStatusInfo = computed(() => {
  const status = overview.value.stats?.my_ty_status
  if (!status || status === 'WITHDRAWN') return null
  return tyStatusMap[status] || null
})

// ---- 过滤有效待办 ----
const activeTodos = computed(() => {
  return (overview.value.todo_items || []).filter(item => item.path)
})
const hasActiveTodos = computed(() => activeTodos.value.length > 0)

function formatNum(n) {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  return n.toLocaleString()
}

onMounted(async () => {
  if (!authStore.user && authStore.isLoggedIn) {
    try {
      await authStore.fetchUser()
    } catch {
      authStore.clearAuth()
      router.push('/login')
      return
    }
  }
  try {
    const data = await dashboardApi.overview()
    overview.value = data
  } catch {
    // 降级：仅展示用户信息
    if (authStore.user) {
      overview.value.user = {
        username: authStore.user.username,
        display_name: authStore.user.display_name,
        roles: authStore.user.roles
      }
    }
  }
})
</script>

<style scoped>
/* ===== 设计令牌 ===== */
.workspace {
  max-width: 1120px;
  margin: 0 auto;
  padding: var(--sh-space-lg);
}

/* ===== 顶部欢迎区 —— 共用基础 ===== */
.welcome-strip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 12px;
  padding: 24px 28px;
  color: #fff;
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
}
.welcome-strip::before {
  content: '';
  position: absolute;
  top: -40%;
  right: -10%;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
}
.welcome-strip::after {
  content: '';
  position: absolute;
  bottom: -30%;
  right: 15%;
  width: 180px;
  height: 180px;
  border-radius: 50%;
}
/* 学生视角：蓝紫渐变 */
.welcome-student {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d5a8e 50%, #3a6ba5 100%);
}
.welcome-student::after {
  background: rgba(155, 89, 182, 0.08);
}
/* 教师视角：深青渐变 */
.welcome-staff {
  background: linear-gradient(135deg, #1a3a3a 0%, #1e5555 50%, #267373 100%);
}
.welcome-staff::after {
  background: rgba(199, 140, 70, 0.07);
}

.welcome-left {
  display: flex;
  align-items: center;
  gap: 18px;
  position: relative;
  z-index: 1;
}
.avatar-ring {
  padding: 3px;
  border-radius: 50%;
  background: linear-gradient(135deg, #c78c46, #e4c28a);
  flex-shrink: 0;
  box-shadow: 0 2px 12px rgba(199, 140, 70, 0.35);
}
.avatar-ring :deep(.el-avatar) {
  background: #1e3a5f;
  color: #e4c28a;
  font-size: 20px;
  font-weight: 700;
}
.welcome-text h1 {
  font-size: 20px;
  font-weight: 700;
  margin: 0 0 6px;
  letter-spacing: 0.5px;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}
.welcome-sub {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.88);
}
.id-badge {
  background: rgba(255, 255, 255, 0.18);
  border: 1px solid rgba(255, 255, 255, 0.22);
  padding: 2px 10px;
  border-radius: 10px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  letter-spacing: 1px;
  color: #fff;
}
.role-tag {
  background: rgba(255, 255, 255, 0.2) !important;
  border-color: rgba(255, 255, 255, 0.35) !important;
  color: #fff !important;
}
.welcome-right {
  text-align: right;
  position: relative;
  z-index: 1;
}
.date-text {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
  letter-spacing: 0.5px;
}

/* ===== 入团进度卡片（学生专属） ===== */
.progress-card {
  background: #fff;
  border-radius: 10px;
  padding: 16px 20px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-left: 4px solid var(--accent);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
.progress-label {
  font-size: 13px;
  color: var(--slate-400, #8e94ab);
}
.progress-body {
  display: flex;
  align-items: center;
  gap: 12px;
}
.progress-status {
  font-size: 14px;
  font-weight: 600;
  color: var(--accent);
}

/* ===== 指标卡片行：根据数量自动均分 ===== */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(var(--n, 3), 1fr);
  gap: 12px;
  margin-bottom: 20px;
}
.metric-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px 16px 16px;
  position: relative;
  overflow: hidden;
  transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1), box-shadow 0.2s ease;
}
.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}
.metric-value {
  font-size: 28px;
  font-weight: 700;
  color: #1e3a5f;
  line-height: 1;
  margin-bottom: 6px;
  font-variant-numeric: tabular-nums;
}
.metric-label {
  font-size: 12px;
  color: #8e94ab;
  letter-spacing: 0.3px;
}
.metric-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  opacity: 0.7;
  border-radius: 0 0 10px 10px;
}

/* ===== 主体双栏 ===== */
.main-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.panel {
  background: #fff;
  border-radius: 10px;
  padding: 24px;
}
.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #1e3a5f;
  margin: 0 0 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eceef5;
}

/* ===== 待办事项 ===== */
.todo-list {
  display: flex;
  flex-direction: column;
}
.todo-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid #eceef5;
  cursor: pointer;
  transition: background 0.15s ease;
}
.todo-item:last-child {
  border-bottom: none;
}
.todo-item:hover {
  background: #f6f7fb;
  margin: 0 -24px;
  padding: 14px 24px;
}
.todo-info {
  display: flex;
  align-items: center;
  gap: 10px;
}
.todo-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.todo-title {
  font-size: 14px;
  color: #2e3347;
}
.todo-right {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #8e94ab;
}
.todo-count {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  min-width: 24px;
  text-align: center;
}
.empty-state {
  text-align: center;
  padding: 32px;
  color: #8e94ab;
  font-size: 13px;
}

/* ===== 快捷入口 ===== */
.links-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.link-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px 8px;
  border-radius: 10px;
  cursor: pointer;
  transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1), background 0.15s ease;
}
.link-card:hover {
  transform: translateY(-2px);
  background: #f6f7fb;
}
.link-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--link-color) 10%, transparent);
  color: var(--link-color);
  transition: background 0.15s ease;
}
.link-card:hover .link-icon-wrap {
  background: color-mix(in srgb, var(--link-color) 18%, transparent);
}
.link-title {
  font-size: 13px;
  color: #5a6078;
  font-weight: 500;
}

/* ===== 响应式 ===== */
@media (max-width: 1024px) {
  .metrics-row {
    grid-template-columns: repeat(min(var(--n, 3), 3), 1fr);
  }
}
@media (max-width: 768px) {
  .welcome-strip {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 20px;
  }
  .welcome-right {
    text-align: left;
  }
  .metrics-row {
    grid-template-columns: repeat(min(var(--n, 3), 2), 1fr);
  }
  .main-grid {
    grid-template-columns: 1fr;
  }
  .links-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 480px) {
  .metrics-row {
    grid-template-columns: repeat(min(var(--n, 3), 2), 1fr);
  }
  .links-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
