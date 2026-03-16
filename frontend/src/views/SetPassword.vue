<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <span class="auth-logo">🔑</span>
        <h1>Set Your Password</h1>
        <p>Create a strong password for your account</p>
      </div>

      <div v-if="!token" class="error-msg">
        Invalid or missing token. Please use the link from your email.
      </div>

      <div v-else-if="success" class="success-msg">
        <span>✅</span>
        <div>
          <strong>Password set successfully!</strong>
          <p>Your account is ready. <RouterLink to="/auth/login">Sign in now</RouterLink></p>
        </div>
      </div>

      <form v-else @submit.prevent="handleSetPassword" class="auth-form">
        <div class="form-group">
          <label>New Password</label>
          <input v-model="form.password" type="password" placeholder="At least 8 characters" required minlength="8" />
        </div>
        <div class="form-group">
          <label>Confirm Password</label>
          <input v-model="form.confirm" type="password" placeholder="Repeat your password" required />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? 'Setting password...' : 'Set Password & Activate Account' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import api from '../api'

const route = useRoute()
const token = ref('')
const form = ref({ password: '', confirm: '' })
const loading = ref(false)
const error = ref('')
const success = ref(false)

onMounted(() => {
  token.value = route.query.token || ''
})

async function handleSetPassword() {
  error.value = ''
  if (form.value.password !== form.value.confirm) {
    error.value = 'Passwords do not match'
    return
  }
  loading.value = true
  try {
    await api.post('/api/auth/set-password', {
      token: token.value,
      password: form.value.password
    })
    success.value = true
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to set password'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  padding: 1rem;
}
.auth-card {
  background: white;
  border-radius: 16px;
  padding: 2.5rem;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.auth-header { text-align: center; margin-bottom: 2rem; }
.auth-logo { font-size: 3rem; display: block; margin-bottom: 0.5rem; }
.auth-header h1 { font-size: 1.75rem; color: #1a1a2e; margin-bottom: 0.25rem; }
.auth-header p { color: #718096; font-size: 0.9rem; }
.auth-form { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.4rem; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: #4a5568; }
.form-group input {
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.65rem 0.9rem;
  font-size: 0.95rem;
  transition: border-color 0.2s;
  outline: none;
}
.form-group input:focus { border-color: #0f3460; }
.error-msg { background: #fed7d7; color: #c53030; padding: 0.75rem; border-radius: 8px; font-size: 0.875rem; }
.success-msg {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: #c6f6d5;
  color: #276749;
  padding: 1.25rem;
  border-radius: 10px;
  font-size: 0.9rem;
}
.success-msg span { font-size: 1.5rem; flex-shrink: 0; }
.success-msg a { color: #276749; font-weight: bold; }
.btn-primary {
  background: #0f3460;
  color: white;
  border: none;
  padding: 0.75rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 0.5rem;
}
.btn-primary:hover:not(:disabled) { background: #16213e; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
