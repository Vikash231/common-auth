<template>
  <div class="callback-page">
    <div class="callback-card">
      <span class="spinner">⟳</span>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import api from '../api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const message = ref('Completing sign in...')

onMounted(async () => {
  const token = route.query.token
  if (!token) {
    message.value = 'Authentication failed. Redirecting...'
    setTimeout(() => router.push('/login'), 2000)
    return
  }

  authStore.setToken(token)
  try {
    const res = await api.get('/api/auth/me')
    authStore.setUser(res.data)
    router.push('/dashboard')
  } catch {
    authStore.logout()
    router.push('/login')
  }
})
</script>

<style scoped>
.callback-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}
.callback-card {
  background: white;
  border-radius: 16px;
  padding: 3rem;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.spinner {
  font-size: 3rem;
  display: block;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
p { color: #4a5568; font-size: 1rem; }
</style>
