<template>
  <div class="workspace-list">
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
          :placeholder="$t('workspace.searchPlaceholder')"
        />
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 5v14M5 12h14"/>
        </svg>
        {{ $t('workspace.newWorkspace') }}
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

    <!-- Workspace Grid -->
    <div v-else class="workspace-grid">
      <div v-if="filteredWorkspaces.length > 0" class="grid">
        <div
          v-for="workspace in filteredWorkspaces"
          :key="workspace.id"
          class="workspace-card"
          @click="$router.push(`/workspaces/${workspace.id}`)"
        >
          <div class="card-header">
            <div class="workspace-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
              </svg>
            </div>
            <span :class="['role-badge', `role-${workspace.role}`]">
              {{ $t(`workspace.role${capitalize(workspace.role)}`) }}
            </span>
          </div>
          <div class="card-body">
            <h3 class="workspace-name">{{ workspace.name }}</h3>
            <p class="workspace-desc">{{ workspace.description || '-' }}</p>
          </div>
          <div class="card-footer">
            <div class="stat">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
              </svg>
              <span>{{ workspace.member_count }}</span>
            </div>
            <div class="stat">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/>
                <line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
              <span>{{ formatDate(workspace.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
          </svg>
        </div>
        <h3 class="empty-title">{{ $t('workspace.noWorkspaces') }}</h3>
        <p class="empty-description">{{ $t('workspace.noWorkspacesDesc') }}</p>
        <button class="btn btn-primary" @click="showCreateModal = true">
          {{ $t('workspace.createWorkspace') }}
        </button>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('workspace.createWorkspace') }}</h3>
          <button class="modal-close" @click="showCreateModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleCreate" class="modal-body">
          <div class="form-group">
            <label class="form-label">
              {{ $t('workspace.name') }}
              <span class="required">*</span>
            </label>
            <input
              v-model="createForm.name"
              type="text"
              class="form-input"
              :placeholder="$t('workspace.namePlaceholder')"
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.description') }}</label>
            <textarea
              v-model="createForm.description"
              class="form-textarea"
              :placeholder="$t('workspace.descriptionPlaceholder')"
              rows="3"
            ></textarea>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="creating">
              {{ creating ? $t('common.loading') : $t('common.create') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { workspaceApi } from '../api'

const { t } = useI18n()

const workspaces = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const showCreateModal = ref(false)
const creating = ref(false)
const createForm = ref({
  name: '',
  description: ''
})

const filteredWorkspaces = computed(() => {
  if (!searchQuery.value) return workspaces.value
  const query = searchQuery.value.toLowerCase()
  return workspaces.value.filter(w =>
    w.name.toLowerCase().includes(query) ||
    (w.description && w.description.toLowerCase().includes(query))
  )
})

function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1)
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

async function loadWorkspaces() {
  loading.value = true
  error.value = null

  try {
    const response = await workspaceApi.list()
    workspaces.value = response.data || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  creating.value = true
  error.value = null

  try {
    await workspaceApi.create(createForm.value)
    showCreateModal.value = false
    createForm.value = { name: '', description: '' }
    loadWorkspaces()
  } catch (e) {
    error.value = e.message
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  loadWorkspaces()
})
</script>

<style scoped>
.workspace-list {
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
  max-width: 320px;
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

.btn-icon {
  width: 18px;
  height: 18px;
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

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #E5E7EB;
  border-top-color: #10B981;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  margin-top: 1rem;
  color: #6B7280;
  font-size: 0.875rem;
}

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
  background: #FEE2E2;
  color: #991B1B;
}

.alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.workspace-card {
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
}

.workspace-card:hover {
  border-color: #10B981;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: #F9FAFB;
  border-bottom: 1px solid #E5E7EB;
}

.workspace-icon {
  width: 40px;
  height: 40px;
  background: #10B981;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.workspace-icon svg {
  width: 22px;
  height: 22px;
  color: white;
}

.role-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.role-owner {
  background: #DBEAFE;
  color: #1E40AF;
}

.role-admin {
  background: #FEF3C7;
  color: #92400E;
}

.role-member {
  background: #E5E7EB;
  color: #374151;
}

.card-body {
  padding: 1.25rem;
}

.workspace-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.5rem 0;
}

.workspace-desc {
  font-size: 0.875rem;
  color: #6B7280;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  border-top: 1px solid #F3F4F6;
}

.stat {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #6B7280;
  font-size: 0.8125rem;
}

.stat svg {
  width: 16px;
  height: 16px;
}

.empty-state {
  padding: 4rem 2rem;
  text-align: center;
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
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
  margin-bottom: 0.5rem;
}

.empty-description {
  color: #6B7280;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
}

/* Modal */
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

.form-group {
  margin-bottom: 1.25rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.required {
  color: #EF4444;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
</style>
