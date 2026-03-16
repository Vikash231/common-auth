<template>
  <div class="page">
    <div class="page-header">
      <h1>⚙️ Settings — User Permissions</h1>
      <p v-if="authStore.isAdmin">As admin, you can manage user permissions and invite new users.</p>
      <p v-else>View users and their permission groups. Contact an admin to change permissions.</p>
    </div>

    <!-- Admin: Invite User -->
    <div v-if="authStore.isAdmin" class="section-card">
      <h2>Invite New User</h2>
      <form @submit.prevent="inviteUser" class="invite-form">
        <div class="form-row">
          <input v-model="invite.name" type="text" placeholder="Full Name" required class="form-input" />
          <input v-model="invite.email" type="email" placeholder="Email address" required class="form-input" />
          <button type="submit" class="btn-primary" :disabled="inviteLoading">
            {{ inviteLoading ? 'Inviting...' : 'Send Invite' }}
          </button>
        </div>
        <div v-if="inviteSuccess" class="success-inline">✓ Invitation sent to {{ invite.email }}</div>
        <div v-if="inviteError" class="error-inline">{{ inviteError }}</div>
      </form>
    </div>

    <!-- Admin: Permission Groups -->
    <div v-if="authStore.isAdmin" class="section-card">
      <div class="section-header">
        <h2>Permission Groups</h2>
        <button class="btn-secondary" @click="showNewGroup = !showNewGroup">
          {{ showNewGroup ? 'Cancel' : '+ New Group' }}
        </button>
      </div>

      <form v-if="showNewGroup" @submit.prevent="createGroup" class="new-group-form">
        <input v-model="newGroup.name" type="text" placeholder="Group name" required class="form-input" />
        <label class="checkbox-label">
          <input type="checkbox" v-model="newGroup.can_use_weather_api" />
          Weather API
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="newGroup.can_use_distance_api" />
          Distance API
        </label>
        <button type="submit" class="btn-primary">Create Group</button>
        <div v-if="groupError" class="error-inline">{{ groupError }}</div>
      </form>

      <div class="groups-grid">
        <div v-for="group in groups" :key="group.id" class="group-card">
          <div class="group-name">{{ group.name }}</div>
          <div class="group-perms">
            <span :class="['perm-pill', group.can_use_weather_api ? 'on' : 'off']">
              {{ group.can_use_weather_api ? '✓' : '✗' }} Weather
            </span>
            <span :class="['perm-pill', group.can_use_distance_api ? 'on' : 'off']">
              {{ group.can_use_distance_api ? '✓' : '✗' }} Distance
            </span>
          </div>
          <div class="group-actions">
            <button class="btn-xs" @click="toggleGroupPerm(group, 'weather')">Toggle Weather</button>
            <button class="btn-xs" @click="toggleGroupPerm(group, 'distance')">Toggle Distance</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Users Table -->
    <div class="section-card">
      <div class="section-header">
        <h2>Users</h2>
        <span class="user-count">{{ users.length }} users</span>
      </div>

      <div v-if="usersLoading" class="loading">Loading users...</div>

      <table v-else class="users-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Status</th>
            <th>Permission Group</th>
            <th>Permissions</th>
            <th v-if="authStore.isAdmin">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>
              <span class="user-name">{{ user.name }}</span>
              <span v-if="user.is_admin" class="admin-tag">Admin</span>
            </td>
            <td class="email-cell">{{ user.email }}</td>
            <td>
              <span :class="['status-badge', user.is_active ? 'active' : 'inactive']">
                {{ user.is_active ? 'Active' : 'Pending' }}
              </span>
            </td>
            <td>
              <select
                v-if="authStore.isAdmin && !user.is_admin"
                :value="user.group?.id"
                @change="updateUserGroup(user.id, $event.target.value)"
                class="group-select"
              >
                <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
              <span v-else>{{ user.group?.name || '—' }}</span>
            </td>
            <td>
              <span v-if="user.group?.can_use_weather_api" class="perm-pill on small">Weather</span>
              <span v-if="user.group?.can_use_distance_api" class="perm-pill on small">Distance</span>
              <span v-if="!user.group?.can_use_weather_api && !user.group?.can_use_distance_api" class="perm-pill off small">None</span>
            </td>
            <td v-if="authStore.isAdmin">
              <button v-if="!user.is_admin" class="btn-danger-xs" @click="deleteUser(user)">Remove</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../store/auth'
import api from '../api'

const authStore = useAuthStore()

const users = ref([])
const groups = ref([])
const usersLoading = ref(false)

const invite = ref({ name: '', email: '' })
const inviteLoading = ref(false)
const inviteSuccess = ref(false)
const inviteError = ref('')

const showNewGroup = ref(false)
const newGroup = ref({ name: '', can_use_weather_api: true, can_use_distance_api: true })
const groupError = ref('')

async function loadData() {
  usersLoading.value = true
  try {
    const [u, g] = await Promise.all([
      api.get('/api/auth/users'),
      api.get('/api/auth/groups')
    ])
    users.value = u.data
    groups.value = g.data
  } catch (e) {
    console.error(e)
  } finally {
    usersLoading.value = false
  }
}

async function inviteUser() {
  inviteError.value = ''
  inviteSuccess.value = false
  inviteLoading.value = true
  try {
    await api.post('/api/auth/admin/invite', invite.value)
    inviteSuccess.value = true
    invite.value = { name: '', email: '' }
    await loadData()
  } catch (e) {
    inviteError.value = e.response?.data?.error || 'Failed to invite user'
  } finally {
    inviteLoading.value = false
  }
}

async function updateUserGroup(userId, groupId) {
  try {
    await api.put(`/api/auth/admin/users/${userId}/group`, { group_id: groupId })
    await loadData()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to update group')
  }
}

async function deleteUser(user) {
  if (!confirm(`Remove user ${user.name}?`)) return
  try {
    await api.delete(`/api/auth/admin/users/${user.id}`)
    await loadData()
  } catch (e) {
    alert(e.response?.data?.error || 'Failed to delete user')
  }
}

async function createGroup() {
  groupError.value = ''
  try {
    await api.post('/api/auth/admin/groups', newGroup.value)
    showNewGroup.value = false
    newGroup.value = { name: '', can_use_weather_api: true, can_use_distance_api: true }
    await loadData()
  } catch (e) {
    groupError.value = e.response?.data?.error || 'Failed to create group'
  }
}

async function toggleGroupPerm(group, type) {
  const updates = {}
  if (type === 'weather') updates.can_use_weather_api = !group.can_use_weather_api
  if (type === 'distance') updates.can_use_distance_api = !group.can_use_distance_api
  try {
    await api.put(`/api/auth/admin/groups/${group.id}`, updates)
    await loadData()
  } catch (e) {
    alert('Failed to update group')
  }
}

onMounted(loadData)
</script>

<style scoped>
.page { padding: 2rem; max-width: 1100px; margin: 0 auto; }
.page-header { margin-bottom: 2rem; }
.page-header h1 { font-size: 1.75rem; color: #1a1a2e; margin-bottom: 0.25rem; }
.page-header p { color: #718096; }
.section-card { background: white; border-radius: 14px; padding: 1.75rem; box-shadow: 0 2px 12px rgba(0,0,0,0.08); margin-bottom: 1.5rem; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.25rem; }
.section-header h2 { font-size: 1.1rem; color: #1a1a2e; margin: 0; }
.user-count { font-size: 0.85rem; color: #718096; background: #f7fafc; padding: 3px 10px; border-radius: 12px; }
.invite-form { margin-top: 0.5rem; }
.form-row { display: flex; gap: 0.75rem; flex-wrap: wrap; }
.form-input {
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.6rem 0.9rem;
  font-size: 0.9rem;
  outline: none;
  flex: 1;
  min-width: 200px;
  transition: border-color 0.2s;
}
.form-input:focus { border-color: #0f3460; }
.btn-primary {
  background: #0f3460;
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  font-size: 0.9rem;
  white-space: nowrap;
  transition: background 0.2s;
}
.btn-primary:hover:not(:disabled) { background: #16213e; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary {
  background: white;
  border: 2px solid #e2e8f0;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-secondary:hover { border-color: #0f3460; color: #0f3460; }
.success-inline { color: #276749; background: #c6f6d5; padding: 0.5rem 0.75rem; border-radius: 6px; font-size: 0.875rem; margin-top: 0.5rem; }
.error-inline { color: #c53030; background: #fed7d7; padding: 0.5rem 0.75rem; border-radius: 6px; font-size: 0.875rem; margin-top: 0.5rem; }
.new-group-form { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; background: #f7fafc; padding: 1rem; border-radius: 10px; margin-bottom: 1rem; }
.checkbox-label { display: flex; align-items: center; gap: 0.4rem; font-size: 0.9rem; cursor: pointer; }
.groups-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1rem; }
.group-card { background: #f7fafc; border-radius: 10px; padding: 1rem; border: 2px solid #e2e8f0; }
.group-name { font-weight: 700; color: #1a1a2e; margin-bottom: 0.5rem; }
.group-perms { display: flex; gap: 0.5rem; margin-bottom: 0.75rem; flex-wrap: wrap; }
.group-actions { display: flex; gap: 0.4rem; flex-wrap: wrap; }
.btn-xs {
  background: white;
  border: 1px solid #cbd5e0;
  border-radius: 5px;
  padding: 3px 8px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-xs:hover { border-color: #0f3460; color: #0f3460; }
.btn-danger-xs {
  background: #fff5f5;
  border: 1px solid #fed7d7;
  color: #c53030;
  border-radius: 5px;
  padding: 3px 8px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-danger-xs:hover { background: #fed7d7; }
.loading { color: #718096; padding: 2rem; text-align: center; }
.users-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
.users-table th { text-align: left; padding: 0.75rem 1rem; border-bottom: 2px solid #e2e8f0; color: #4a5568; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.5px; }
.users-table td { padding: 0.75rem 1rem; border-bottom: 1px solid #f0f4f8; vertical-align: middle; }
.users-table tr:last-child td { border-bottom: none; }
.user-name { font-weight: 600; color: #1a1a2e; }
.admin-tag { background: #e53e3e; color: white; font-size: 0.65rem; padding: 2px 6px; border-radius: 10px; margin-left: 6px; font-weight: 600; vertical-align: middle; }
.email-cell { color: #718096; font-size: 0.875rem; }
.status-badge { font-size: 0.75rem; padding: 3px 10px; border-radius: 12px; font-weight: 600; }
.status-badge.active { background: #c6f6d5; color: #276749; }
.status-badge.inactive { background: #fefcbf; color: #744210; }
.perm-pill { font-size: 0.75rem; padding: 3px 10px; border-radius: 12px; font-weight: 600; margin-right: 4px; display: inline-block; }
.perm-pill.on { background: #c6f6d5; color: #276749; }
.perm-pill.off { background: #fed7d7; color: #c53030; }
.perm-pill.small { font-size: 0.7rem; padding: 2px 8px; }
.group-select {
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.3rem 0.5rem;
  font-size: 0.875rem;
  outline: none;
  cursor: pointer;
}
.group-select:focus { border-color: #0f3460; }
</style>
