<template>
  <el-container class="layout-container">
    <!-- 顶部通栏 Header -->
    <el-header class="layout-header" :height="headerHeight">
      <div class="header-left">
        <!-- Logo -->
        <div class="header-logo" @click="$router.push('/dashboard')">
          <img src="/images/logo.png" alt="Logo" class="logo-img" />
          <span class="logo-text">学生一站式管理平台</span>
        </div>
      </div>
      <div class="header-center" />
      <div class="header-right">
        <!-- 首页按钮 -->
        <div class="header-nav-item" :class="{ active: isHome }" @click="$router.push('/dashboard')">
          <el-icon :size="16"><HomeFilled /></el-icon>
          <span>首页</span>
        </div>
        <!-- 通知铃铛 -->
        <NotificationBell class="header-bell" />
        <div class="header-divider" />
        <el-dropdown trigger="click" @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" icon="UserFilled" class="user-avatar" />
            <span class="user-name">{{ authStore.displayName || '用户' }}</span>
            <el-icon class="user-arrow"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>
                <div class="role-tags">
                  <el-tag size="small" v-for="role in authStore.roles" :key="role" effect="plain">
                    {{ role }}
                  </el-tag>
                </div>
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <!-- 下方区域：侧边栏 + 主内容 -->
    <el-container class="layout-body">
      <!-- 侧边栏 -->
      <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
        <div class="sidebar-toggle" @click="isCollapse = !isCollapse">
          <el-icon :size="16">
            <component :is="isCollapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          background-color="transparent"
          text-color="var(--sh-text-secondary)"
          active-text-color="var(--sh-primary)"
          @select="(index) => console.log('[Menu] select ->', index, '| current route =', $route.path)"
        >
          <template v-for="menu in menuStore.menuList" :key="menu.code">
            <!-- 有子菜单 -->
            <el-sub-menu v-if="menu.children && menu.children.length" :index="menu.path">
              <template #title>
                <el-icon><component :is="menu.icon" /></el-icon>
                <span>{{ menu.title }}</span>
              </template>
              <el-menu-item
                v-for="child in menu.children"
                :key="child.code"
                :index="child.path"
              >
                {{ child.title }}
              </el-menu-item>
            </el-sub-menu>
            <!-- 无子菜单 -->
            <el-menu-item v-else :index="menu.path">
              <el-icon><component :is="menu.icon" /></el-icon>
              <template #title>{{ menu.title }}</template>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <!-- 主区域 -->
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useMenuStore } from '@/stores/menu'
import { HomeFilled, ArrowDown, SwitchButton, Expand, Fold } from '@element-plus/icons-vue'
import NotificationBell from '@/layouts/components/NotificationBell.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const menuStore = useMenuStore()

const isCollapse = ref(false)
const headerHeight = '64px'
const activeMenu = computed(() => route.path)
const isHome = computed(() => route.path === '/dashboard')

async function handleCommand(cmd) {
  if (cmd === 'logout') {
    await authStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  flex-direction: column;
  background: var(--sh-bg-base);
}

/* ── 顶部通栏 Header ── */
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--sh-bg-white);
  border-bottom: 1px solid var(--sh-border-light);
  padding: 0 var(--sh-space-lg);
  z-index: 10;
  box-shadow: var(--sh-shadow-sm);
}

/* Header 内 Logo */
.header-logo {
  display: flex;
  align-items: center;
  cursor: pointer;
  gap: 10px;
  padding: 4px 8px;
  margin-left: -16px;
  border-radius: var(--sh-radius-md);
  transition: background var(--sh-duration-fast) var(--sh-ease-out);
}
.header-logo:hover {
  background: var(--sh-primary-lighter);
}
.header-logo .logo-img {
  height: 36px;
  flex-shrink: 0;
}
.header-logo .logo-text {
  color: var(--sh-primary-dark);
  font-size: var(--sh-text-lg);
  font-weight: 700;
  white-space: nowrap;
  letter-spacing: 0.02em;
}

/* 中间导航 */
.header-center {
  display: flex;
  align-items: center;
  gap: var(--sh-space-xs);
}
.header-nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 8px 16px;
  border-radius: var(--sh-radius-md);
  color: var(--sh-text-secondary);
  font-size: var(--sh-text-sm);
  font-weight: 500;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}
.header-nav-item:hover {
  color: var(--sh-primary);
  background: var(--sh-primary-lighter);
}
.header-nav-item.active {
  color: var(--sh-primary);
  background: var(--sh-primary-lighter);
  font-weight: 600;
}

/* 右侧区域 */
.header-left {
  display: flex;
  align-items: center;
}
.header-right {
  display: flex;
  align-items: center;
  gap: var(--sh-space-sm);
}
.header-divider {
  width: 1px;
  height: 24px;
  background: var(--sh-border-light);
  margin: 0 var(--sh-space-xs);
}
.header-bell {
  margin-right: 4px;
}
.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  gap: 8px;
  padding: 4px 8px;
  border-radius: var(--sh-radius-md);
  transition: background var(--sh-duration-fast) var(--sh-ease-out);
}
.user-info:hover {
  background: var(--sh-bg-elevated);
}
.user-avatar {
  background: var(--sh-primary);
}
.user-name {
  font-size: var(--sh-text-sm);
  color: var(--sh-text-regular);
  font-weight: 500;
}
.user-arrow {
  font-size: 12px;
  color: var(--sh-text-placeholder);
  transition: transform var(--sh-duration-fast) var(--sh-ease-out);
}
.role-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

/* ── 下方 body 区域 ── */
.layout-body {
  flex: 1;
  overflow: hidden;
}

/* ── 侧边栏 ── */
.layout-aside {
  background: var(--sh-bg-white);
  border-right: 1px solid var(--sh-border-light);
  transition: width 0.3s var(--sh-ease-out);
  overflow-y: auto;
  overflow-x: hidden;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}
.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  cursor: pointer;
  color: var(--sh-text-placeholder);
  border-bottom: 1px solid var(--sh-border-light);
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
  flex-shrink: 0;
}
.sidebar-toggle:hover {
  color: var(--sh-primary);
  background: var(--sh-primary-lighter);
}

/* ── 主区域 ── */
.layout-main {
  background: var(--sh-bg-base);
  overflow-y: auto;
  padding: 0;
}

/* ── Element Plus 菜单样式覆盖 ── */
:deep(.el-menu) {
  border-right: none;
  width: 100%;
  padding: var(--sh-space-xs);
}
:deep(.el-menu-item) {
  border-radius: var(--sh-radius-md);
  margin: 2px 0;
  height: 44px;
  line-height: 44px;
  font-weight: 500;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}
:deep(.el-menu-item:hover) {
  background: var(--sh-primary-lighter) !important;
  color: var(--sh-primary) !important;
}
:deep(.el-menu-item.is-active) {
  background: var(--sh-primary-lighter) !important;
  color: var(--sh-primary) !important;
  font-weight: 600;
  position: relative;
}
:deep(.el-menu-item.is-active::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  background: var(--sh-primary);
  border-radius: 0 2px 2px 0;
}
:deep(.el-sub-menu__title) {
  border-radius: var(--sh-radius-md);
  margin: 2px 0;
  height: 44px;
  line-height: 44px;
  font-weight: 500;
  transition: all var(--sh-duration-fast) var(--sh-ease-out);
}
:deep(.el-sub-menu__title:hover) {
  background: var(--sh-primary-lighter) !important;
  color: var(--sh-primary) !important;
}
:deep(.el-sub-menu .el-menu-item) {
  padding-left: 52px !important;
  height: 40px;
  line-height: 40px;
  font-size: var(--sh-text-sm);
}
</style>
