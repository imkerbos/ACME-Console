import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import { useAuth } from './stores/auth'
import i18n from './locales'
import './styles/main.css'

// Views
import Login from './views/Login.vue'
import CertificateList from './views/CertificateList.vue'
import CertificateDetail from './views/CertificateDetail.vue'
import CertificateCreate from './views/CertificateCreate.vue'
import WorkspaceList from './views/WorkspaceList.vue'
import WorkspaceDetail from './views/WorkspaceDetail.vue'
import Profile from './views/Profile.vue'
import UserList from './views/UserList.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { guest: true }
  },
  {
    path: '/',
    redirect: '/certificates'
  },
  {
    path: '/certificates',
    name: 'CertificateList',
    component: CertificateList,
    meta: { requiresAuth: true }
  },
  {
    path: '/certificates/new',
    name: 'CertificateCreate',
    component: CertificateCreate,
    meta: { requiresAuth: true }
  },
  {
    path: '/certificates/:id',
    name: 'CertificateDetail',
    component: CertificateDetail,
    meta: { requiresAuth: true }
  },
  {
    path: '/workspaces',
    name: 'WorkspaceList',
    component: WorkspaceList,
    meta: { requiresAuth: true }
  },
  {
    path: '/workspaces/:id',
    name: 'WorkspaceDetail',
    component: WorkspaceDetail,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true }
  },
  {
    path: '/users',
    name: 'UserList',
    component: UserList,
    meta: { requiresAuth: true, requiresAdmin: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guards
router.beforeEach((to, from, next) => {
  const { isAuthenticated, isAdmin } = useAuth()
  const authenticated = isAuthenticated()

  if (to.meta.requiresAuth && !authenticated) {
    next('/login')
  } else if (to.meta.guest && authenticated) {
    next('/')
  } else if (to.meta.requiresAdmin && !isAdmin()) {
    next('/')
  } else {
    next()
  }
})

const app = createApp(App)
app.use(router)
app.use(i18n)
app.mount('#app')
