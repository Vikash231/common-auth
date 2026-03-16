import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/auth'

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', component: () => import('../views/Login.vue'), meta: { guest: true } },
  { path: '/signup', component: () => import('../views/Signup.vue'), meta: { guest: true } },
  { path: '/set-password', component: () => import('../views/SetPassword.vue'), meta: { guest: true } },
  { path: '/auth/callback', component: () => import('../views/AuthCallback.vue'), meta: { guest: true } },
  {
    path: '/dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/weather',
    component: () => import('../views/Weather.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/distance',
    component: () => import('../views/Distance.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    component: () => import('../views/Settings.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  authStore.initFromStorage()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next('/login')
  }
  if (to.meta.guest && authStore.isAuthenticated) {
    return next('/dashboard')
  }
  next()
})

export default router
