<template>
  <div class="page">
    <div class="page-header">
      <h1>🌤️ Weather Service</h1>
      <p>Search for current weather conditions by city or location name</p>
    </div>

    <div class="search-card">
      <form @submit.prevent="fetchWeather" class="search-form">
        <input
          v-model="place"
          type="text"
          placeholder="Enter city name (e.g. London, Tokyo, New York)..."
          class="search-input"
        />
        <button type="submit" class="btn-search" :disabled="loading || !place.trim()">
          {{ loading ? '...' : 'Search' }}
        </button>
      </form>
    </div>

    <div v-if="error" class="error-box">{{ error }}</div>

    <div v-if="weather" class="weather-card">
      <div class="weather-header">
        <div>
          <h2>{{ weather.place }}</h2>
          <p class="weather-condition">{{ weather.condition }}</p>
        </div>
        <div class="temperature">{{ weather.temperature_celsius }}°C</div>
      </div>

      <div class="weather-grid">
        <div class="weather-stat">
          <span class="stat-icon">💧</span>
          <span class="stat-label">Humidity</span>
          <span class="stat-value">{{ weather.humidity_percent }}%</span>
        </div>
        <div class="weather-stat">
          <span class="stat-icon">💨</span>
          <span class="stat-label">Wind Speed</span>
          <span class="stat-value">{{ weather.wind_speed_kmh }} km/h</span>
        </div>
        <div class="weather-stat">
          <span class="stat-icon">👁️</span>
          <span class="stat-label">Visibility</span>
          <span class="stat-value">{{ weather.visibility_km }} km</span>
        </div>
        <div class="weather-stat">
          <span class="stat-icon">🕐</span>
          <span class="stat-label">Updated</span>
          <span class="stat-value">{{ formatTime(weather.fetched_at) }}</span>
        </div>
      </div>
    </div>

    <div v-else-if="!loading && !error" class="empty-state">
      <span>🔍</span>
      <p>Enter a location above to see its weather</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../api'

const place = ref('')
const weather = ref(null)
const loading = ref(false)
const error = ref('')

async function fetchWeather() {
  if (!place.value.trim()) return
  error.value = ''
  weather.value = null
  loading.value = true
  try {
    const res = await api.get('/api/weather', { params: { place: place.value } })
    weather.value = res.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to fetch weather data'
  } finally {
    loading.value = false
  }
}

function formatTime(iso) {
  return new Date(iso).toLocaleTimeString()
}
</script>

<style scoped>
.page { padding: 2rem; max-width: 800px; margin: 0 auto; }
.page-header { margin-bottom: 2rem; }
.page-header h1 { font-size: 1.75rem; color: #1a1a2e; margin-bottom: 0.25rem; }
.page-header p { color: #718096; }
.search-card { background: white; border-radius: 14px; padding: 1.5rem; box-shadow: 0 2px 12px rgba(0,0,0,0.08); margin-bottom: 1.5rem; }
.search-form { display: flex; gap: 0.75rem; }
.search-input {
  flex: 1;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.65rem 1rem;
  font-size: 0.95rem;
  outline: none;
  transition: border-color 0.2s;
}
.search-input:focus { border-color: #0f3460; }
.btn-search {
  background: #0f3460;
  color: white;
  border: none;
  padding: 0.65rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}
.btn-search:hover:not(:disabled) { background: #16213e; }
.btn-search:disabled { opacity: 0.6; cursor: not-allowed; }
.error-box { background: #fed7d7; color: #c53030; padding: 1rem; border-radius: 10px; margin-bottom: 1.5rem; }
.weather-card { background: linear-gradient(135deg, #0f3460, #16213e); color: white; border-radius: 14px; padding: 2rem; box-shadow: 0 8px 24px rgba(15,52,96,0.3); }
.weather-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 2rem; }
.weather-header h2 { font-size: 1.5rem; margin-bottom: 0.3rem; }
.weather-condition { color: #a0c4ff; font-size: 0.95rem; }
.temperature { font-size: 3.5rem; font-weight: 700; color: #fff; }
.weather-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
.weather-stat { background: rgba(255,255,255,0.1); border-radius: 10px; padding: 1rem; }
.stat-icon { font-size: 1.2rem; display: block; margin-bottom: 0.4rem; }
.stat-label { display: block; font-size: 0.75rem; color: #a0c4ff; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 0.25rem; }
.stat-value { font-size: 1rem; font-weight: 600; }
.empty-state { text-align: center; padding: 4rem 2rem; color: #718096; }
.empty-state span { font-size: 3rem; display: block; margin-bottom: 1rem; }
</style>
