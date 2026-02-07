<template>
  <div class="settings-page">
    <main class="settings-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>SETTINGS</h1>
            <p class="subtitle">System and user management</p>
          </div>
        </div>

        <div class="settings-grid">
          <!-- Account Section -->
          <Card class="section-card accent-bar">
            <h2 class="section-title">ACCOUNT</h2>
            <div class="section-content">
              <div class="info-row">
                <span class="label">USERNAME</span>
                <span class="value">{{ currentUser?.username || '--' }}</span>
              </div>

              <div class="form-divider"></div>

              <h3 class="subsection-title">CHANGE PASSWORD</h3>
              <div class="form-group">
                <Input
                  v-model="passwordForm.current"
                  label="CURRENT PASSWORD"
                  type="password"
                  placeholder="Enter current password"
                />
              </div>
              <div class="form-group">
                <Input
                  v-model="passwordForm.new"
                  label="NEW PASSWORD"
                  type="password"
                  placeholder="Enter new password"
                />
              </div>
              <div class="form-group">
                <Input
                  v-model="passwordForm.confirm"
                  label="CONFIRM PASSWORD"
                  type="password"
                  placeholder="Confirm new password"
                />
              </div>
              <Button
                variant="primary"
                @click="handleChangePassword"
                :loading="changingPassword"
              >
                {{ changingPassword ? 'UPDATING...' : 'UPDATE PASSWORD' }}
              </Button>
            </div>
          </Card>

          <!-- System Section -->
          <Card class="section-card accent-bar">
            <h2 class="section-title">SYSTEM</h2>
            <div class="section-content">
              <div class="form-group">
                <Input
                  v-model="systemForm.hostname"
                  label="HOSTNAME / DISPLAY NAME"
                  type="text"
                  placeholder="Enter system hostname"
                />
              </div>
              <Button
                variant="primary"
                @click="handleSaveSystemSettings"
                :loading="savingSystem"
              >
                {{ savingSystem ? 'SAVING...' : 'SAVE SETTINGS' }}
              </Button>
            </div>
          </Card>

          <!-- User Management Section -->
          <Card class="section-card accent-bar full-width">
            <h2 class="section-title">USER MANAGEMENT</h2>
            <div class="section-content">
              <div v-if="loading" class="loading-state-inline">
                <div class="spinner"></div>
                <p>Loading users...</p>
              </div>

              <div v-else-if="users.length === 0" class="empty-state-inline">
                <p class="text-secondary">No additional users</p>
              </div>

              <table v-else class="user-table">
                <thead>
                  <tr>
                    <th>USERNAME</th>
                    <th>CREATED</th>
                    <th>ACTIONS</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="user in users" :key="user.id">
                    <td class="font-mono">{{ user.username }}</td>
                    <td class="text-secondary">{{ formatDate(user.created_at) }}</td>
                    <td>
                      <Button
                        variant="danger"
                        size="sm"
                        @click="handleDeleteUserClick(user)"
                        :disabled="users.length === 1"
                      >
                        DELETE
                      </Button>
                    </td>
                  </tr>
                </tbody>
              </table>

              <div class="form-divider"></div>

              <h3 class="subsection-title">ADD USER</h3>
              <div class="form-row">
                <Input
                  v-model="newUser.username"
                  placeholder="Username"
                  type="text"
                />
                <Input
                  v-model="newUser.password"
                  placeholder="Password"
                  type="password"
                />
                <Button
                  variant="primary"
                  @click="handleAddUser"
                  :loading="addingUser"
                >
                  {{ addingUser ? 'ADDING...' : 'ADD USER' }}
                </Button>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </main>

    <!-- Delete User Confirmation Modal -->
    <Modal v-model="showDeleteModal" title="CONFIRM DELETE USER">
      <p>Are you sure you want to delete user <strong>{{ userToDelete?.username }}</strong>?</p>
      <p class="text-secondary mt-sm">This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showDeleteModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="confirmDeleteUser" :loading="deletingUser">
          {{ deletingUser ? 'DELETING...' : 'DELETE USER' }}
        </Button>
      </template>
    </Modal>

    <!-- Toast Notification -->
    <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const settingsStore = useSettingsStore()

const passwordForm = ref({
  current: '',
  new: '',
  confirm: ''
})

const systemForm = ref({
  hostname: ''
})

const newUser = ref({
  username: '',
  password: ''
})

const changingPassword = ref(false)
const savingSystem = ref(false)
const addingUser = ref(false)
const deletingUser = ref(false)
const showDeleteModal = ref(false)
const userToDelete = ref(null)
const toast = ref({ show: false, message: '', type: 'success' })

const users = ref([])
const currentUser = ref(null)
const loading = ref(false)

function formatDate(dateString) {
  if (!dateString) return '--'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

async function handleChangePassword() {
  if (!passwordForm.value.current || !passwordForm.value.new || !passwordForm.value.confirm) {
    showToast('All password fields are required', 'error')
    return
  }

  if (passwordForm.value.new !== passwordForm.value.confirm) {
    showToast('New passwords do not match', 'error')
    return
  }

  if (passwordForm.value.new.length < 6) {
    showToast('New password must be at least 6 characters', 'error')
    return
  }

  changingPassword.value = true
  try {
    await settingsStore.changePassword(
      currentUser.value.id,
      passwordForm.value.current,
      passwordForm.value.new
    )
    showToast('Password updated successfully', 'success')
    passwordForm.value = { current: '', new: '', confirm: '' }
  } catch (err) {
    showToast(err.response?.data?.error || 'Failed to update password', 'error')
  } finally {
    changingPassword.value = false
  }
}

async function handleSaveSystemSettings() {
  if (!systemForm.value.hostname) {
    showToast('Hostname is required', 'error')
    return
  }

  savingSystem.value = true
  try {
    await settingsStore.updateSettings({ hostname: systemForm.value.hostname })
    showToast('System settings saved successfully', 'success')
  } catch (err) {
    showToast('Failed to save system settings', 'error')
  } finally {
    savingSystem.value = false
  }
}

async function handleAddUser() {
  if (!newUser.value.username || !newUser.value.password) {
    showToast('Username and password are required', 'error')
    return
  }

  if (newUser.value.password.length < 6) {
    showToast('Password must be at least 6 characters', 'error')
    return
  }

  addingUser.value = true
  try {
    await settingsStore.createUser(newUser.value.username, newUser.value.password)
    showToast('User added successfully', 'success')
    newUser.value = { username: '', password: '' }
    await loadUsers()
  } catch (err) {
    showToast(err.response?.data?.error || 'Failed to add user', 'error')
  } finally {
    addingUser.value = false
  }
}

function handleDeleteUserClick(user) {
  userToDelete.value = user
  showDeleteModal.value = true
}

async function confirmDeleteUser() {
  deletingUser.value = true
  try {
    await settingsStore.deleteUser(userToDelete.value.id)
    showToast('User deleted successfully', 'success')
    showDeleteModal.value = false
    await loadUsers()
  } catch (err) {
    showToast(err.response?.data?.error || 'Failed to delete user', 'error')
  } finally {
    deletingUser.value = false
    userToDelete.value = null
  }
}

async function loadUsers() {
  loading.value = true
  try {
    await settingsStore.fetchUsers()
    users.value = settingsStore.users
  } catch (err) {
    showToast('Failed to load users', 'error')
  } finally {
    loading.value = false
  }
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(async () => {
  await settingsStore.fetchCurrentUser()
  currentUser.value = settingsStore.currentUser

  await loadUsers()

  await settingsStore.fetchSettings()
  systemForm.value.hostname = settingsStore.settings.hostname || ''
})
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
}

.settings-main {
  padding-top: var(--space-2xl);
  padding-bottom: var(--space-3xl);
}

.page-header {
  margin-bottom: var(--space-2xl);
}

.page-header h1 {
  margin-bottom: var(--space-sm);
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--space-xl);
}

.section-card {
  padding: var(--space-xl);
}

.section-card.full-width {
  grid-column: 1 / -1;
}

.section-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  color: var(--reach-amber);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: var(--space-lg);
}

.subsection-title {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--reach-cyan);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-md);
}

.section-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
}

.info-row .label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.info-row .value {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--reach-amber);
}

.form-divider {
  height: 1px;
  background-color: rgba(74, 85, 104, 0.3);
  margin: var(--space-md) 0;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-row {
  display: flex;
  gap: var(--space-md);
  align-items: flex-end;
}

.form-row .input-wrapper {
  flex: 1;
}

.user-table {
  width: 100%;
  border-collapse: collapse;
  font-family: var(--font-mono);
  font-size: 0.875rem;
}

.user-table thead {
  border-bottom: 2px solid var(--reach-amber);
}

.user-table th {
  text-align: left;
  padding: var(--space-md);
  font-size: 0.75rem;
  color: var(--reach-amber);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.user-table td {
  padding: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.user-table tbody tr:hover {
  background-color: rgba(74, 85, 104, 0.1);
}

.loading-state-inline,
.empty-state-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-xl);
  text-align: center;
}

.loading-state-inline {
  gap: var(--space-md);
}

.toast {
  position: fixed;
  bottom: var(--space-xl);
  right: var(--space-xl);
  padding: var(--space-md) var(--space-lg);
  background-color: var(--reach-steel);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  z-index: var(--z-toast);
  animation: slideIn 0.3s ease-out;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  min-width: 250px;
}

.toast-success {
  border: 1px solid var(--reach-cyan);
  color: var(--reach-cyan);
}

.toast-error {
  border: 1px solid var(--reach-orange);
  color: var(--reach-orange);
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }

  .form-row {
    flex-direction: column;
    align-items: stretch;
  }

  .user-table {
    font-size: 0.75rem;
  }

  .user-table th,
  .user-table td {
    padding: var(--space-sm);
  }
}
</style>
