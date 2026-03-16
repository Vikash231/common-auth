<template>
  <div class="page">
    <div class="page-header">
      <h1>Welcome, {{ authStore.user?.name }} 👋</h1>
      <p>Choose a service to get started</p>
    </div>

    <div class="cards-grid">
      <RouterLink to="/weather" class="service-card">
        <div class="card-icon weather-icon">🌤️</div>
        <div class="card-body">
          <h2>Weather</h2>
          <p>Get real-time weather conditions for any place in the world.</p>
          <span :class="['perm-badge', hasWeather ? 'allowed' : 'denied']">
            {{ hasWeather ? '✓ Access granted' : '✗ No access' }}
          </span>
        </div>
        <span class="card-arrow">→</span>
      </RouterLink>

      <RouterLink to="/distance" class="service-card">
        <div class="card-icon distance-icon">📍</div>
        <div class="card-body">
          <h2>Distance</h2>
          <p>Calculate travel time and distance between two locations.</p>
          <span :class="['perm-badge', hasDistance ? 'allowed' : 'denied']">
            {{ hasDistance ? '✓ Access granted' : '✗ No access' }}
          </span>
        </div>
        <span class="card-arrow">→</span>
      </RouterLink>

      <RouterLink to="/settings" class="service-card">
        <div class="card-icon settings-icon">⚙️</div>
        <div class="card-body">
          <h2>Settings</h2>
          <p>View user permissions and manage groups{{ authStore.isAdmin ? ' (admin)' : '' }}.</p>
          <span class="perm-badge allowed">✓ Available</span>
        </div>
        <span class="card-arrow">→</span>
      </RouterLink>
    </div>

    <div class="info-box">
      <h3>Your Account</h3>
      <div class="info-grid">
        <div><span class="label">Email</span><span>{{ authStore.user?.email }}</span></div>
        <div><span class="label">Role</span><span>{{ authStore.isAdmin ? 'Administrator' : 'User' }}</span></div>
        <div>
          <span class="label">Permissions</span>
          <span>
            <span v-for="p in authStore.permissions" :key="p" class="perm-tag">{{ p }}</span>
            <span v-if="!authStore.permissions.length" class="perm-tag denied">none</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()

const hasWeather = computed(() => authStore.permissions.includes('weather:read'))
const hasDistance = computed(() => authStore.permissions.includes('distance:read'))
</script>

<style scoped>
.page { padding: 2rem; max-width: 1100px; margin: 0 auto; }
.page-header { margin-bottom: 2rem; }
.page-header h1 { font-size: 1.75rem; color: #1a1a2e; margin-bottom: 0.25rem; }
.page-header p { color: #718096; }
.cards-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
.service-card {
  background: white;
  border-radius: 14px;
  padding: 1.75rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  text-decoration: none;
  color: inherit;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  border: 2px solid transparent;
  transition: all 0.25s;
  position: relative;
}
.service-card:hover { border-color: #0f3460; box-shadow: 0 8px 24px rgba(0,0,0,0.12); transform: translateY(-2px); }
.card-icon { font-size: 2.5rem; }
.card-body h2 { font-size: 1.2rem; color: #1a1a2e; margin-bottom: 0.4rem; }
.card-body p { color: #718096; font-size: 0.875rem; line-height: 1.5; margin-bottom: 0.75rem; }
.card-arrow { position: absolute; top: 1.75rem; right: 1.75rem; font-size: 1.2rem; color: #a0aec0; transition: transform 0.2s; }
.service-card:hover .card-arrow { transform: translateX(4px); color: #0f3460; }
.perm-badge { font-size: 0.75rem; font-weight: 600; padding: 3px 10px; border-radius: 12px; }
.perm-badge.allowed { background: #c6f6d5; color: #276749; }
.perm-badge.denied { background: #fed7d7; color: #c53030; }
.info-box { background: white; border-radius: 14px; padding: 1.75rem; box-shadow: 0 2px 12px rgba(0,0,0,0.08); }
.info-box h3 { font-size: 1rem; color: #4a5568; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 1rem; }
.info-grid { display: flex; flex-direction: column; gap: 0.75rem; }
.info-grid > div { display: flex; gap: 1rem; align-items: center; font-size: 0.9rem; }
.label { color: #718096; min-width: 100px; font-weight: 500; }
.perm-tag { background: #ebf8ff; color: #2b6cb0; font-size: 0.75rem; padding: 2px 10px; border-radius: 12px; margin-right: 4px; }
.perm-tag.denied { background: #fff5f5; color: #c53030; }
</style>
