<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <span class="auth-logo">✉️</span>
        <h1>Create Account</h1>
        <p>You'll receive an email to set your password</p>
      </div>

      <div v-if="success" class="success-msg">
        <span>📬</span>
        <div>
          <strong>Verification email sent!</strong>
          <p>Check your inbox at <strong>{{ sentTo }}</strong> and click the link to set your password.</p>
        </div>
      </div>

      <form v-else @submit.prevent="handleSignup" class="auth-form">
        <div class="form-group">
          <label>Full Name</label>
          <input v-model="form.name" type="text" placeholder="John Doe" required />
        </div>
        <div class="form-group">
          <label>Email Address</label>
          <input v-model="form.email" type="email" placeholder="you@example.com" required />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? 'Sending...' : 'Create Account' }}
        </button>
      </form>

      <p class="auth-footer">
        Already have an account?
        <RouterLink to="/login">Sign in</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import api from '../api'

const form = ref({ name: '', email: '' })
const loading = ref(false)
const error = ref('')
const success = ref(false)
const sentTo = ref('')

async function handleSignup() {
  error.value = ''
  loading.value = true
  try {
    await api.post('/api/auth/signup', form.value)
    sentTo.value = form.value.email
    success.value = true
  } catch (err) {
    error.value = err.response?.data?.error || 'Signup failed'
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
  margin-bottom: 1rem;
}
.success-msg span { font-size: 1.5rem; flex-shrink: 0; }
.success-msg p { margin-top: 0.25rem; }
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
.auth-footer { text-align: center; margin-top: 1.5rem; font-size: 0.875rem; color: #718096; }
.auth-footer a { color: #0f3460; font-weight: 600; text-decoration: none; }
.auth-footer a:hover { text-decoration: underline; }
</style>
