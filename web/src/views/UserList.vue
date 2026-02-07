<template>
  <div class="user-list">
    <!-- Action Bar -->
    <div class="action-bar">
      <div class="search-box">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <path d="M21 21l-4.35-4.35"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          :placeholder="$t('user.searchPlaceholder')"
          @input="handleSearch"
        />
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 5v14M5 12h14"/>
        </svg>
        {{ $t('user.createUser') }}
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <p class="loading-text">{{ $t('common.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="alert alert-error">
      <svg class="alert-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <path d="M12 8v4M12 16h.01"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <!-- User Table -->
    <div v-else class="table-card">
      <table v-if="users.length > 0" class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ $t('user.username') }}</th>
            <th>{{ $t('user.nickname') }}</th>
            <th>{{ $t('user.email') }}</th>
            <th>{{ $t('user.role') }}</th>
            <th>{{ $t('user.status') }}</th>
            <th>{{ $t('user.lastLogin') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td class="cell-id">#{{ user.id }}</td>
            <td class="cell-username">
              <span class="username">{{ user.username }}</span>
            </td>
            <td>{{ user.nickname || '-' }}</td>
            <td>{{ user.email || '-' }}</td>
            <td>
              <span :class="['role-badge', `role-${user.role}`]">
                {{ user.role === 'admin' ? $t('user.admin') : $t('user.regularUser') }}
              </span>
            </td>
            <td>
              <span :class="['status-badge', user.status === 1 ? 'status-active' : 'status-inactive']">
                <span class="status-dot"></span>
                {{ user.status === 1 ? $t('user.active') : $t('user.inactive') }}
              </span>
            </td>
            <td class="cell-date">{{ user.last_login || '-' }}</td>
            <td class="cell-actions">
              <button class="btn-icon-only" @click="editUser(user)" :title="$t('common.edit')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
              </button>
              <button class="btn-icon-only" @click="resetPasswordModal(user)" :title="$t('user.resetPassword')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                  <path d="M7 11V7a5 5 0 0110 0v4"/>
                </svg>
              </button>
              <button
                class="btn-icon-only btn-danger"
                @click="confirmDelete(user)"
                :title="$t('common.delete')"
                :disabled="user.id === currentUserId"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3,6 5,6 21,6"/>
                  <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 00-3-3.87"/>
            <path d="M16 3.13a4 4 0 010 7.75"/>
          </svg>
        </div>
        <h3 class="empty-title">No users found</h3>
        <button class="btn btn-primary" @click="showCreateModal = true">
          {{ $t('user.createUser') }}
        </button>
      </div>

      <!-- Pagination -->
      <div v-if="pagination.totalPages > 1" class="pagination-container">
        <div class="pagination-info">
          {{ $t('pagination.showing', { start: (pagination.page - 1) * pagination.pageSize + 1, end: Math.min(pagination.page * pagination.pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="pagination">
          <button
            class="pagination-btn"
            :disabled="pagination.page <= 1"
            @click="loadPage(pagination.page - 1)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M15 19l-7-7 7-7"/>
            </svg>
          </button>
          <button
            v-for="p in displayedPages"
            :key="p"
            :class="['pagination-btn', { active: p === pagination.page }]"
            @click="loadPage(p)"
          >
            {{ p }}
          </button>
          <button
            class="pagination-btn"
            :disabled="pagination.page >= pagination.totalPages"
            @click="loadPage(pagination.page + 1)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 5l7 7-7 7"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingUser" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ editingUser ? $t('user.editUser') : $t('user.createUser') }}</h3>
          <button class="modal-close" @click="closeModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleSubmit" class="modal-body">
          <div v-if="modalError" class="alert alert-error">{{ modalError }}</div>

          <div class="form-group" v-if="!editingUser">
            <label class="form-label">{{ $t('user.username') }} *</label>
            <input
              v-model="form.username"
              type="text"
              class="form-input"
              required
              minlength="3"
              maxlength="50"
            />
          </div>

          <div class="form-group">
            <label class="form-label">{{ $t('user.nickname') }}</label>
            <input v-model="form.nickname" type="text" class="form-input" />
          </div>

          <div class="form-group">
            <label class="form-label">{{ $t('user.email') }}</label>
            <input v-model="form.email" type="email" class="form-input" />
          </div>

          <div class="form-group" v-if="!editingUser">
            <label class="form-label">{{ $t('auth.password') }} *</label>
            <input v-model="form.password" type="password" class="form-input" required minlength="6" />
          </div>

          <div class="form-group">
            <label class="form-label">{{ $t('user.role') }}</label>
            <select v-model="form.role" class="form-select">
              <option value="user">{{ $t('user.regularUser') }}</option>
              <option value="admin">{{ $t('user.admin') }}</option>
            </select>
          </div>

          <div class="form-group" v-if="editingUser">
            <label class="form-label">{{ $t('user.status') }}</label>
            <select v-model="form.status" class="form-select">
              <option :value="1">{{ $t('user.active') }}</option>
              <option :value="0">{{ $t('user.inactive') }}</option>
            </select>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="closeModal">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? $t('common.loading') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Reset Password Modal -->
    <div v-if="showResetModal" class="modal-overlay" @click.self="showResetModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ $t('user.resetPassword') }}</h3>
          <button class="modal-close" @click="showResetModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleResetPassword" class="modal-body">
          <div v-if="modalError" class="alert alert-error">{{ modalError }}</div>

          <p class="reset-info">{{ resetUser?.username }}</p>

          <div class="form-group">
            <label class="form-label">{{ $t('user.newPassword') }} *</label>
            <input v-model="newPassword" type="password" class="form-input" required minlength="6" />
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showResetModal = false">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? $t('common.loading') : $t('common.confirm') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ $t('common.confirm') }}</h3>
          <button class="modal-close" @click="showDeleteModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <p>{{ $t('user.deleteConfirm') }}</p>
          <p class="delete-user-info">{{ deleteUser?.username }}</p>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">
              {{ $t('common.cancel') }}
            </button>
            <button type="button" class="btn btn-danger" :disabled="submitting" @click="handleDelete">
              {{ submitting ? $t('common.loading') : $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { userApi } from '../api'
import { useAuth } from '../stores/auth'

const { getUser } = useAuth()
const currentUserId = computed(() => getUser()?.id)

const users = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const searchTimeout = ref(null)
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0,
  totalPages: 0
})

// Modals
const showCreateModal = ref(false)
const showResetModal = ref(false)
const showDeleteModal = ref(false)
const editingUser = ref(null)
const resetUser = ref(null)
const deleteUser = ref(null)
const submitting = ref(false)
const modalError = ref(null)
const newPassword = ref('')

const form = ref({
  username: '',
  nickname: '',
  email: '',
  password: '',
  role: 'user',
  status: 1
})

const displayedPages = computed(() => {
  const pages = []
  const total = pagination.value.totalPages
  const current = pagination.value.page

  let start = Math.max(1, current - 2)
  let end = Math.min(total, current + 2)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

async function loadUsers(page = 1) {
  loading.value = true
  error.value = null

  try {
    const response = await userApi.list({
      page,
      page_size: 20,
      keyword: searchQuery.value
    })
    const data = response.data

    users.value = data.items || []
    pagination.value = {
      page: data.page,
      pageSize: data.page_size,
      total: data.total,
      totalPages: data.total_pages
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function loadPage(page) {
  loadUsers(page)
}

function handleSearch() {
  clearTimeout(searchTimeout.value)
  searchTimeout.value = setTimeout(() => {
    loadUsers(1)
  }, 300)
}

function editUser(user) {
  editingUser.value = user
  form.value = {
    username: user.username,
    nickname: user.nickname || '',
    email: user.email || '',
    password: '',
    role: user.role,
    status: user.status
  }
  modalError.value = null
}

function resetPasswordModal(user) {
  resetUser.value = user
  newPassword.value = ''
  modalError.value = null
  showResetModal.value = true
}

function confirmDelete(user) {
  deleteUser.value = user
  showDeleteModal.value = true
}

function closeModal() {
  showCreateModal.value = false
  editingUser.value = null
  modalError.value = null
  form.value = {
    username: '',
    nickname: '',
    email: '',
    password: '',
    role: 'user',
    status: 1
  }
}

async function handleSubmit() {
  modalError.value = null
  submitting.value = true

  try {
    if (editingUser.value) {
      await userApi.update(editingUser.value.id, {
        nickname: form.value.nickname,
        email: form.value.email,
        role: form.value.role,
        status: form.value.status
      })
    } else {
      await userApi.create({
        username: form.value.username,
        password: form.value.password,
        nickname: form.value.nickname,
        email: form.value.email,
        role: form.value.role
      })
    }

    closeModal()
    loadUsers(pagination.value.page)
  } catch (e) {
    modalError.value = e.message
  } finally {
    submitting.value = false
  }
}

async function handleResetPassword() {
  modalError.value = null
  submitting.value = true

  try {
    await userApi.resetPassword(resetUser.value.id, newPassword.value)
    showResetModal.value = false
  } catch (e) {
    modalError.value = e.message
  } finally {
    submitting.value = false
  }
}

async function handleDelete() {
  submitting.value = true

  try {
    await userApi.delete(deleteUser.value.id)
    showDeleteModal.value = false
    loadUsers(pagination.value.page)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-list {
  max-width: 100%;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 360px;
}

.search-icon {
  position: absolute;
  left: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: #9CA3AF;
}

.search-input {
  width: 100%;
  padding: 0.625rem 0.875rem 0.625rem 2.5rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  background: white;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.btn-icon {
  width: 18px;
  height: 18px;
  margin-right: 0.5rem;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
}

.loading-text {
  margin-top: 1rem;
  color: #6B7280;
  font-size: 0.875rem;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #E5E7EB;
  border-top-color: #10B981;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.alert-error {
  background: #FEE2E2;
  color: #991B1B;
  border: 1px solid #FECACA;
}

.alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.table-card {
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th {
  background: #F9FAFB;
  padding: 0.75rem 1rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #6B7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  text-align: left;
  border-bottom: 1px solid #E5E7EB;
}

.table td {
  padding: 1rem;
  border-bottom: 1px solid #F3F4F6;
  vertical-align: middle;
}

.table tbody tr:last-child td {
  border-bottom: none;
}

.table tbody tr:hover {
  background: #F9FAFB;
}

.cell-id {
  font-family: monospace;
  color: #6B7280;
  font-size: 0.875rem;
}

.cell-username .username {
  font-weight: 500;
  color: #111827;
}

.role-badge {
  display: inline-flex;
  padding: 0.25rem 0.625rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
}

.role-admin {
  background: #EEF2FF;
  color: #4F46E5;
}

.role-user {
  background: #F3F4F6;
  color: #374151;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-active {
  background: #D1FAE5;
  color: #065F46;
}

.status-active .status-dot {
  background: #10B981;
}

.status-inactive {
  background: #FEE2E2;
  color: #991B1B;
}

.status-inactive .status-dot {
  background: #EF4444;
}

.cell-date {
  color: #6B7280;
  font-size: 0.875rem;
  white-space: nowrap;
}

.cell-actions {
  text-align: right;
  white-space: nowrap;
}

.btn-icon-only {
  padding: 0.375rem;
  background: none;
  border: none;
  color: #6B7280;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.btn-icon-only:hover:not(:disabled) {
  background: #F3F4F6;
  color: #111827;
}

.btn-icon-only:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-icon-only.btn-danger:hover:not(:disabled) {
  background: #FEE2E2;
  color: #DC2626;
}

.btn-icon-only svg {
  width: 18px;
  height: 18px;
}

.empty-state {
  padding: 4rem 2rem;
  text-align: center;
}

.empty-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 1.5rem;
  background: #F3F4F6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-icon svg {
  width: 32px;
  height: 32px;
  color: #9CA3AF;
}

.empty-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 1.5rem;
}

.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-top: 1px solid #E5E7EB;
}

.pagination-info {
  color: #6B7280;
  font-size: 0.875rem;
}

.pagination {
  display: flex;
  gap: 0.25rem;
}

.pagination-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 36px;
  height: 36px;
  padding: 0 0.75rem;
  background: white;
  border: 1px solid #E5E7EB;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  background: #F3F4F6;
  border-color: #D1D5DB;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-btn.active {
  background: #10B981;
  border-color: #10B981;
  color: white;
}

.pagination-btn svg {
  width: 16px;
  height: 16px;
}

/* Modals */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow: auto;
}

.modal-sm {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #E5E7EB;
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.modal-close {
  padding: 0.375rem;
  background: none;
  border: none;
  color: #6B7280;
  cursor: pointer;
  border-radius: 6px;
}

.modal-close:hover {
  background: #F3F4F6;
  color: #111827;
}

.modal-close svg {
  width: 20px;
  height: 20px;
}

.modal-body {
  padding: 1.5rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-input,
.form-select {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.reset-info {
  font-weight: 500;
  color: #111827;
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: #F9FAFB;
  border-radius: 8px;
}

.delete-user-info {
  font-weight: 500;
  color: #DC2626;
  margin-top: 0.5rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #10B981;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #059669;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #F3F4F6;
  color: #374151;
}

.btn-secondary:hover {
  background: #E5E7EB;
}

.btn-danger {
  background: #EF4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #DC2626;
}
</style>
