<template>
  <div id="app">
    <Navbar v-if="authStore.isAuthenticated" />
    <main :class="authStore.isAuthenticated ? 'with-nav' : 'full-page'">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { RouterView } from 'vue-router'
import Navbar from './components/Navbar.vue'
import { useAuthStore } from './store/auth'

const authStore = useAuthStore()
authStore.initFromStorage()
</script>

<style>
.with-nav { padding: 0; }
.full-page { min-height: 100vh; }
</style>
