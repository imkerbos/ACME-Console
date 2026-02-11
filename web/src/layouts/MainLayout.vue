<template>
  <div class="layout">
    <!-- Sidebar -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo">
          <svg viewBox="0 0 100 100" class="logo-icon">
            <rect width="100" height="100" rx="16" fill="#10B981"/>
            <path d="M25 70 L50 30 L75 70 M35 55 L65 55" stroke="white" stroke-width="8" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <div v-if="!sidebarCollapsed" class="logo-info">
            <span class="logo-text">{{ siteTitle }}</span>
            <span v-if="siteSubtitle" class="logo-subtitle">{{ siteSubtitle }}</span>
          </div>
        </div>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/certificates" class="nav-item">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
          <span v-if="!sidebarCollapsed" class="nav-label">{{ $t('nav.certificates') }}</span>
        </router-link>

        <router-link to="/workspaces" class="nav-item">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
          </svg>
          <span v-if="!sidebarCollapsed" class="nav-label">{{ $t('nav.workspaces') }}</span>
        </router-link>

        <!-- System Management Group (Admin Only) -->
        <div v-if="isAdmin" class="nav-group">
          <button class="nav-item nav-group-toggle" :class="{ active: isSystemRoute }" @click="toggleSystemMenu">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            <span v-if="!sidebarCollapsed" class="nav-label">{{ $t('nav.system') }}</span>
            <svg v-if="!sidebarCollapsed" class="nav-arrow" :class="{ expanded: systemMenuOpen }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 5l7 7-7 7"/>
            </svg>
          </button>
          <div v-if="systemMenuOpen && !sidebarCollapsed" class="nav-submenu">
            <router-link to="/users" class="nav-item nav-sub-item">
              <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 00-3-3.87"/>
                <path d="M16 3.13a4 4 0 010 7.75"/>
              </svg>
              <span class="nav-label">{{ $t('nav.users') }}</span>
            </router-link>
            <router-link to="/settings" class="nav-item nav-sub-item">
              <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/>
              </svg>
              <span class="nav-label">{{ $t('nav.settings') }}</span>
            </router-link>
            <router-link to="/about" class="nav-item nav-sub-item">
              <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 16v-4M12 8h.01"/>
              </svg>
              <span class="nav-label">{{ $t('nav.about') }}</span>
            </router-link>
          </div>
        </div>
      </nav>

      <div class="sidebar-footer">
        <button @click="toggleSidebar" class="collapse-btn">
          <svg class="collapse-icon" :class="{ rotated: sidebarCollapsed }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M15 19l-7-7 7-7"/>
          </svg>
        </button>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="main-wrapper">
      <!-- Header -->
      <header class="header">
        <div class="header-left">
          <h2 class="page-title">{{ pageTitle }}</h2>
        </div>
        <div class="header-right">
          <!-- Language Switcher -->
          <div class="lang-switcher">
            <button
              :class="['lang-btn', { active: currentLocale === 'zh-CN' }]"
              @click="setLocale('zh-CN')"
            >
              中
            </button>
            <button
              :class="['lang-btn', { active: currentLocale === 'en-US' }]"
              @click="setLocale('en-US')"
            >
              EN
            </button>
          </div>

          <div class="user-menu" @click="showUserMenu = !showUserMenu">
            <div class="user-avatar">
              {{ userInitial }}
            </div>
            <span class="user-name">{{ user?.nickname || user?.username }}</span>
            <span v-if="isAdmin" class="user-role">{{ $t('user.admin') }}</span>
            <svg class="dropdown-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 9l-7 7-7-7"/>
            </svg>

            <div v-if="showUserMenu" class="dropdown-menu">
              <router-link to="/profile" class="dropdown-item">{{ $t('nav.profile') }}</router-link>
              <button @click="handleLogout" class="dropdown-item danger">{{ $t('nav.logout') }}</button>
            </div>
          </div>
        </div>
      </header>

      <!-- Content -->
      <main class="content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../stores/auth'
import { useSite } from '../stores/site'
import { setLocale as setAppLocale, getLocale } from '../locales'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const { getUser, isAdmin: checkAdmin, logout } = useAuth()
const site = useSite()

site.load()

const sidebarCollapsed = ref(false)
const showUserMenu = ref(false)
const systemMenuOpen = ref(true)
const user = computed(() => getUser())
const isAdmin = computed(() => checkAdmin())
const currentLocale = computed(() => locale.value)
const isSystemRoute = computed(() => ['/users', '/settings', '/about'].includes(route.path))
const siteTitle = computed(() => site.getTitle())
const siteSubtitle = computed(() => site.getSubtitle())

const userInitial = computed(() => {
  const name = user.value?.nickname || user.value?.username || 'U'
  return name.charAt(0).toUpperCase()
})

const pageTitle = computed(() => {
  const path = route.path

  // Use i18n for page titles
  const titles = {
    '/certificates': t('nav.certificates'),
    '/certificates/new': t('certificate.newCertificate'),
    '/workspaces': t('nav.workspaces'),
    '/profile': t('nav.profile'),
    '/users': t('nav.users'),
    '/settings': t('nav.settings'),
    '/about': t('nav.about')
  }

  if (titles[path]) {
    return titles[path]
  }

  // Dynamic routes
  if (path.match(/^\/certificates\/\d+$/)) {
    return `${t('certificate.certificates')} #${route.params.id}`
  }

  if (path.match(/^\/workspaces\/\d+$/)) {
    return t('workspace.workspaceDetail')
  }

  return 'Dashboard'
})

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function toggleSystemMenu() {
  if (sidebarCollapsed.value) return
  systemMenuOpen.value = !systemMenuOpen.value
}

function handleLogout() {
  logout()
  router.push('/login')
}

function setLocale(lang) {
  setAppLocale(lang)
}

// Close dropdown when clicking outside
document.addEventListener('click', (e) => {
  if (!e.target.closest('.user-menu')) {
    showUserMenu.value = false
  }
})
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
  background: #F3F4F6;
}

/* Sidebar */
.sidebar {
  width: 260px;
  background: #1F2937;
  color: white;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  position: fixed;
  height: 100vh;
  z-index: 100;
}

.sidebar.collapsed {
  width: 72px;
}

.sidebar-header {
  padding: 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
}

.logo-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.logo-text {
  font-weight: 700;
  font-size: 1.125rem;
  white-space: nowrap;
}

.logo-subtitle {
  font-size: 0.6875rem;
  color: #9CA3AF;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 0.125rem;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0.75rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  color: #9CA3AF;
  text-decoration: none;
  transition: all 0.2s;
  margin-bottom: 0.25rem;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.nav-item.router-link-active {
  background: #10B981;
  color: white;
}

.nav-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.nav-label {
  white-space: nowrap;
}

/* System menu group */
.nav-group {
  margin-top: 0.25rem;
}

.nav-group-toggle {
  width: 100%;
  background: none;
  border: none;
  font: inherit;
  cursor: pointer;
  position: relative;
}

.nav-group-toggle.active {
  color: #D1FAE5;
}

.nav-arrow {
  width: 14px;
  height: 14px;
  margin-left: auto;
  transition: transform 0.2s;
  flex-shrink: 0;
}

.nav-arrow.expanded {
  transform: rotate(90deg);
}

.nav-submenu {
  padding-left: 0.5rem;
}

.nav-sub-item {
  padding-left: 1.25rem !important;
  font-size: 0.875rem;
}

.nav-sub-item .nav-icon {
  width: 16px;
  height: 16px;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.collapse-btn {
  width: 100%;
  padding: 0.5rem;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 6px;
  color: #9CA3AF;
  cursor: pointer;
  display: flex;
  justify-content: center;
}

.collapse-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: white;
}

.collapse-icon {
  width: 20px;
  height: 20px;
  transition: transform 0.3s;
}

.collapse-icon.rotated {
  transform: rotate(180deg);
}

/* Main Wrapper */
.main-wrapper {
  flex: 1;
  margin-left: 260px;
  transition: margin-left 0.3s ease;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.sidebar.collapsed + .main-wrapper {
  margin-left: 72px;
}

/* Header */
.header {
  background: white;
  border-bottom: 1px solid #E5E7EB;
  padding: 0 1.5rem;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 50;
}

.page-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0;
  color: #111827;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

/* Language Switcher */
.lang-switcher {
  display: flex;
  background: #F3F4F6;
  border-radius: 6px;
  padding: 2px;
}

.lang-btn {
  padding: 0.375rem 0.625rem;
  background: none;
  border: none;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  color: #6B7280;
  cursor: pointer;
  transition: all 0.2s;
}

.lang-btn:hover {
  color: #111827;
}

.lang-btn.active {
  background: white;
  color: #10B981;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-radius: 8px;
  cursor: pointer;
  position: relative;
}

.user-menu:hover {
  background: #F3F4F6;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #10B981 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.875rem;
}

.user-name {
  font-weight: 500;
  color: #374151;
}

.user-role {
  padding: 0.125rem 0.5rem;
  background: #EEF2FF;
  color: #4F46E5;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.dropdown-icon {
  width: 16px;
  height: 16px;
  color: #9CA3AF;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.5rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  padding: 0.5rem;
  z-index: 100;
}

.dropdown-item {
  display: block;
  width: 100%;
  padding: 0.625rem 1rem;
  border: none;
  background: none;
  text-align: left;
  color: #374151;
  text-decoration: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
}

.dropdown-item:hover {
  background: #F3F4F6;
}

.dropdown-item.danger {
  color: #EF4444;
}

.dropdown-item.danger:hover {
  background: #FEE2E2;
}

/* Content */
.content {
  flex: 1;
  padding: 1.5rem;
}
</style>
