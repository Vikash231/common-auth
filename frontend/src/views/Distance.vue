<template>
  <div class="page">
    <div class="page-header">
      <h1>📍 Distance Service</h1>
      <p>Calculate travel time and distance between two locations</p>
    </div>

    <div class="search-card">
      <form @submit.prevent="fetchDistance" class="search-form">
        <div class="location-inputs">
          <div class="input-group">
            <span class="input-icon">🟢</span>
            <input v-model="from" type="text" placeholder="From (e.g. New York)" class="location-input" />
          </div>
          <div class="swap-btn" @click="swap">⇅</div>
          <div class="input-group">
            <span class="input-icon">🔴</span>
            <input v-model="to" type="text" placeholder="To (e.g. Los Angeles)" class="location-input" />
          </div>
        </div>
        <button type="submit" class="btn-search" :disabled="loading || !from.trim() || !to.trim()">
          {{ loading ? 'Calculating...' : 'Calculate Distance' }}
        </button>
      </form>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div v-if="result" class="result-card">
      <div class="route-header">
        <div class="route-places">
          <span class="route-from">{{ result.from }}</span>
          <span class="route-arrow">→</span>
          <span class="route-to">{{ result.to }}</span>
        </div>
        <div class="route-distance">{{ result.distance_km }} km</div>
      </div>

      <div class="travel-modes">
        <div class="travel-mode">
          <span class="mode-icon">🚗</span>
          <span class="mode-label">Driving</span>
          <span class="mode-time">{{ formatTime(result.driving_time_minutes) }}</span>
        </div>
        <div class="travel-mode">
          <span class="mode-icon">🚌</span>
          <span class="mode-label">Transit</span>
          <span class="mode-time">{{ formatTime(result.transit_time_minutes) }}</span>
        </div>
        <div class="travel-mode">
          <span class="mode-icon">🚶</span>
          <span class="mode-label">Walking</span>
          <span class="mode-time">{{ formatTime(result.walking_time_minutes) }}</span>
        </div>
      </div>

      <p class="fetched-at">Calculated at {{ new Date(result.fetched_at).toLocaleString() }}</p>
    </div>

    <div v-else-if="!loading && !error" class="empty-state">
      <span>🗺️</span>
      <p>Enter source and destination to calculate travel times</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../api'

const from = ref('')
const to = ref('')
const result = ref(null)
const loading = ref(false)
const error = ref('')

async function fetchDistance() {
  if (!from.value.trim() || !to.value.trim()) return
  error.value = ''
  result.value = null
  loading.value = true
  try {
    const res = await api.get('/api/distance', { params: { from: from.value, to: to.value } })
    result.value = res.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to calculate distance'
  } finally {
    loading.value = false
  }
}

function swap() {
  const tmp = from.value
  from.value = to.value
  to.value = tmp
}

function formatTime(minutes) {
  if (minutes < 60) return `${minutes} min`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}h ${m}min` : `${h}h`
}
</script>

<style scoped>
.page { padding: 2rem; max-width: 800px; margin: 0 auto; }
.page-header { margin-bottom: 2rem; }
.page-header h1 { font-size: 1.75rem; color: #1a1a2e; margin-bottom: 0.25rem; }
.page-header p { color: #718096; }
.search-card { background: white; border-radius: 14px; padding: 1.5rem; box-shadow: 0 2px 12px rgba(0,0,0,0.08); margin-bottom: 1.5rem; }
.search-form { display: flex; flex-direction: column; gap: 1rem; }
.location-inputs { display: flex; flex-direction: column; gap: 0.5rem; position: relative; }
.input-group { display: flex; align-items: center; gap: 0.75rem; }
.input-icon { font-size: 1rem; width: 20px; text-align: center; }
.location-input {
  flex: 1;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.65rem 1rem;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.2s;
}
.location-input:focus { border-color: #0f3460; }
.swap-btn { align-self: center; cursor: pointer; font-size: 1.2rem; color: #718096; padding: 0.25rem; transition: color 0.2s; margin-left: 1.75rem; }
.swap-btn:hover { color: #0f3460; }
.btn-search {
  background: #0f3460;
  color: white;
  border: none;
  padding: 0.75rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  font-size: 1rem;
  transition: background 0.2s;
}
.btn-search:hover:not(:disabled) { background: #16213e; }
.btn-search:disabled { opacity: 0.6; cursor: not-allowed; }
.error-box { background: #fed7d7; color: #c53030; padding: 1rem; border-radius: 10px; margin-bottom: 1.5rem; }
.result-card { background: white; border-radius: 14px; padding: 2rem; box-shadow: 0 2px 12px rgba(0,0,0,0.08); }
.route-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; padding-bottom: 1.5rem; border-bottom: 1px solid #e2e8f0; }
.route-places { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.route-from, .route-to { font-weight: 700; font-size: 1.1rem; color: #1a1a2e; }
.route-arrow { color: #718096; font-size: 1.2rem; }
.route-distance { font-size: 1.75rem; font-weight: 700; color: #0f3460; }
.travel-modes { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
.travel-mode { text-align: center; background: #f7fafc; border-radius: 10px; padding: 1.25rem 1rem; }
.mode-icon { font-size: 1.75rem; display: block; margin-bottom: 0.5rem; }
.mode-label { display: block; font-size: 0.8rem; color: #718096; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 0.4rem; }
.mode-time { font-size: 1.1rem; font-weight: 700; color: #1a1a2e; }
.fetched-at { margin-top: 1rem; font-size: 0.8rem; color: #a0aec0; text-align: right; }
.empty-state { text-align: center; padding: 4rem 2rem; color: #718096; }
.empty-state span { font-size: 3rem; display: block; margin-bottom: 1rem; }
</style>
