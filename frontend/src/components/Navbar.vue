<template>
  <nav class="navbar">
    <div class="nav-brand">
      <span class="logo">🔐</span>
      <span class="brand-name">Common Auth</span>
    </div>
    <div class="nav-links">
      <RouterLink to="/dashboard">Dashboard</RouterLink>
      <RouterLink to="/weather">Weather</RouterLink>
      <RouterLink to="/distance">Distance</RouterLink>
      <RouterLink to="/settings">Settings</RouterLink>
    </div>
    <div class="nav-user">
      <span class="user-name">{{ authStore.user?.name }}</span>
      <span v-if="authStore.isAdmin" class="admin-badge">Admin</span>
      <button class="logout-btn" @click="logout">Logout</button>
    </div>
  </nav>
</template>

<script setup>
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import api from '../api'

const authStore = useAuthStore()
const router = useRouter()

async function logout() {
  await api.post('/api/auth/logout').catch(() => {})
  authStore.logout()
  router.push('/auth/login')
}
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #1a1a2e;
  color: white;
  padding: 0 2rem;
  height: 60px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  position: sticky;
  top: 0;
  z-index: 100;
}
.nav-brand { display: flex; align-items: center; gap: 0.5rem; }
.logo { font-size: 1.4rem; }
.brand-name { font-weight: 700; font-size: 1.1rem; letter-spacing: 0.5px; }
.nav-links { display: flex; gap: 0.25rem; }
.nav-links a {
  color: #a0aec0;
  text-decoration: none;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: all 0.2s;
}
.nav-links a:hover, .nav-links a.router-link-active {
  color: white;
  background: rgba(255,255,255,0.1);
}
.nav-user { display: flex; align-items: center; gap: 0.75rem; }
.user-name { font-size: 0.9rem; color: #e2e8f0; }
.admin-badge {
  background: #e53e3e;
  color: white;
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 600;
  text-transform: uppercase;
}
.logout-btn {
  background: rgba(255,255,255,0.1);
  color: white;
  border: 1px solid rgba(255,255,255,0.2);
  padding: 0.35rem 0.9rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}
.logout-btn:hover { background: rgba(255,255,255,0.2); }
</style>
