<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <span class="auth-logo">🔐</span>
        <h1>Sign In</h1>
        <p>Welcome back to Common Auth</p>
      </div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label>Email</label>
          <input v-model="form.email" type="email" placeholder="you@example.com" required />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="form.password" type="password" placeholder="••••••••" required />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <div class="divider"><span>or continue with</span></div>

      <div class="social-logins">
        <a :href="oidcURL" class="btn-social btn-google">
          <svg width="18" height="18" viewBox="0 0 48 48"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.18 1.48-4.97 2.29-8.16 2.29-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/></svg>
          Sign in with Google
        </a>
        <a :href="samlURL" class="btn-social btn-okta">
          <svg width="18" height="18" viewBox="0 0 48 48" fill="white"><circle cx="24" cy="24" r="20" fill="#007DC1"/><text x="24" y="30" text-anchor="middle" fill="white" font-size="16" font-weight="bold">O</text></svg>
          Sign in with Okta (SAML)
        </a>
      </div>

      <p class="auth-footer">
        Don't have an account?
        <RouterLink to="/auth/signup">Create one</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const router = useRouter()

const form = ref({ email: '', password: '' })
const loading = ref(false)
const error = ref('')

const oidcURL = '/api/auth/oidc/login'
const samlURL = '/api/auth/saml/login'

async function handleLogin() {
  error.value = ''
  loading.value = true
  const ok = await authStore.login(form.value.email, form.value.password)
  loading.value = false
  if (ok) {
    router.push('/dashboard')
  } else {
    error.value = authStore.error || 'Login failed'
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
.divider { text-align: center; margin: 1.5rem 0; position: relative; }
.divider::before { content: ''; position: absolute; top: 50%; left: 0; right: 0; height: 1px; background: #e2e8f0; }
.divider span { background: white; padding: 0 0.75rem; color: #718096; font-size: 0.85rem; position: relative; }
.social-logins { display: flex; flex-direction: column; gap: 0.75rem; }
.btn-social {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  padding: 0.65rem;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.2s;
  border: 2px solid #e2e8f0;
  color: #4a5568;
  background: white;
}
.btn-social:hover { border-color: #cbd5e0; background: #f7fafc; }
.btn-google:hover { border-color: #EA4335; }
.btn-okta { background: #007DC1; color: white; border-color: #007DC1; }
.btn-okta:hover { background: #0066a0; border-color: #0066a0; }
.auth-footer { text-align: center; margin-top: 1.5rem; font-size: 0.875rem; color: #718096; }
.auth-footer a { color: #0f3460; font-weight: 600; text-decoration: none; }
.auth-footer a:hover { text-decoration: underline; }
</style>
