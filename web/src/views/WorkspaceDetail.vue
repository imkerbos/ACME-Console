<template>
  <div class="workspace-detail">
    <div class="back-link">
      <router-link to="/workspaces" class="link-back">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        {{ $t('workspace.backToWorkspaces') }}
      </router-link>
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

    <template v-else-if="workspace">
      <!-- Workspace Header -->
      <div class="workspace-header">
        <div class="header-info">
          <div class="workspace-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
            </svg>
          </div>
          <div>
            <h1 class="workspace-name">{{ workspace.name }}</h1>
            <p class="workspace-desc">{{ workspace.description || '-' }}</p>
          </div>
        </div>
        <div class="header-actions">
          <span :class="['role-badge', `role-${workspace.role}`]">
            {{ $t(`workspace.role${capitalize(workspace.role)}`) }}
          </span>
          <router-link :to="`/certificates?workspace_id=${workspace.id}`" class="btn btn-secondary">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
            </svg>
            {{ $t('workspace.viewCertificates') }}
          </router-link>
          <button v-if="canManage" class="btn btn-secondary" @click="showEditModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            {{ $t('common.edit') }}
          </button>
          <button v-if="isOwner" class="btn btn-danger-outline" @click="showDeleteModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3,6 5,6 21,6"/>
              <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
            {{ $t('common.delete') }}
          </button>
        </div>
      </div>

      <!-- Members Section -->
      <div class="section">
        <div class="section-header">
          <h2>{{ $t('workspace.memberManagement') }}</h2>
          <button v-if="canManage" class="btn btn-primary btn-sm" @click="showAddMemberModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
              <circle cx="8.5" cy="7" r="4"/>
              <line x1="20" y1="8" x2="20" y2="14"/>
              <line x1="23" y1="11" x2="17" y2="11"/>
            </svg>
            {{ $t('workspace.addMember') }}
          </button>
        </div>

        <div class="members-table">
          <table class="table">
            <thead>
              <tr>
                <th>{{ $t('user.username') }}</th>
                <th>{{ $t('user.nickname') }}</th>
                <th>{{ $t('user.email') }}</th>
                <th>{{ $t('workspace.role') }}</th>
                <th>{{ $t('workspace.createdAt') }}</th>
                <th v-if="canManage"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="member in members" :key="member.id">
                <td class="cell-username">{{ member.username }}</td>
                <td>{{ member.nickname || '-' }}</td>
                <td>{{ member.email || '-' }}</td>
                <td>
                  <span :class="['role-badge', `role-${member.role}`]">
                    {{ $t(`workspace.role${capitalize(member.role)}`) }}
                  </span>
                </td>
                <td class="cell-date">{{ formatDate(member.created_at) }}</td>
                <td v-if="canManage" class="cell-actions">
                  <template v-if="member.role !== 'owner'">
                    <button class="btn btn-ghost btn-sm" @click="openUpdateRoleModal(member)">
                      {{ $t('workspace.updateRole') }}
                    </button>
                    <button class="btn btn-ghost btn-sm btn-danger-text" @click="confirmRemoveMember(member)">
                      {{ $t('workspace.removeMember') }}
                    </button>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Notifications Section -->
      <div class="section" style="margin-top: 1.5rem;">
        <WebhookConfig :workspace-id="workspace.id" />
      </div>
    </template>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('workspace.editWorkspace') }}</h3>
          <button class="modal-close" @click="showEditModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleUpdate" class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.name') }} <span class="required">*</span></label>
            <input v-model="editForm.name" type="text" class="form-input" required />
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.description') }}</label>
            <textarea v-model="editForm.description" class="form-textarea" rows="3"></textarea>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showEditModal = false">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="updating">
              {{ updating ? $t('common.loading') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Add Member Modal -->
    <div v-if="showAddMemberModal" class="modal-overlay" @click.self="showAddMemberModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('workspace.addMember') }}</h3>
          <button class="modal-close" @click="showAddMemberModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleAddMember" class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.selectUser') }} <span class="required">*</span></label>
            <select v-model="addMemberForm.user_id" class="form-select" required>
              <option value="">{{ $t('workspace.selectUser') }}</option>
              <option v-for="user in availableUsers" :key="user.id" :value="user.id">
                {{ user.username }} ({{ user.nickname || user.email || '-' }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.selectRole') }} <span class="required">*</span></label>
            <select v-model="addMemberForm.role" class="form-select" required>
              <option value="member">{{ $t('workspace.roleMember') }}</option>
              <option value="admin">{{ $t('workspace.roleAdmin') }}</option>
            </select>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showAddMemberModal = false">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="addingMember">
              {{ addingMember ? $t('common.loading') : $t('common.confirm') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Update Role Modal -->
    <div v-if="showUpdateRoleModal" class="modal-overlay" @click.self="showUpdateRoleModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ $t('workspace.updateRole') }}</h3>
          <button class="modal-close" @click="showUpdateRoleModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleUpdateRole" class="modal-body">
          <p>{{ selectedMember?.username }}</p>
          <div class="form-group">
            <label class="form-label">{{ $t('workspace.selectRole') }}</label>
            <select v-model="updateRoleForm.role" class="form-select" required>
              <option value="member">{{ $t('workspace.roleMember') }}</option>
              <option value="admin">{{ $t('workspace.roleAdmin') }}</option>
            </select>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showUpdateRoleModal = false">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="updatingRole">
              {{ updatingRole ? $t('common.loading') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Workspace Modal -->
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
          <p>{{ $t('workspace.deleteConfirm') }}</p>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">{{ $t('common.cancel') }}</button>
            <button type="button" class="btn btn-danger" :disabled="deleting" @click="handleDelete">
              {{ deleting ? $t('common.loading') : $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Remove Member Modal -->
    <div v-if="showRemoveMemberModal" class="modal-overlay" @click.self="showRemoveMemberModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ $t('common.confirm') }}</h3>
          <button class="modal-close" @click="showRemoveMemberModal = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <p>{{ $t('workspace.removeMember') }}: {{ selectedMember?.username }}?</p>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showRemoveMemberModal = false">{{ $t('common.cancel') }}</button>
            <button type="button" class="btn btn-danger" :disabled="removingMember" @click="handleRemoveMember">
              {{ removingMember ? $t('common.loading') : $t('common.confirm') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { workspaceApi, userApi } from '../api'
import WebhookConfig from '../components/WebhookConfig.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspace = ref(null)
const members = ref([])
const allUsers = ref([])
const loading = ref(true)
const error = ref(null)

// Modals
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const showAddMemberModal = ref(false)
const showUpdateRoleModal = ref(false)
const showRemoveMemberModal = ref(false)

// Form states
const updating = ref(false)
const deleting = ref(false)
const addingMember = ref(false)
const updatingRole = ref(false)
const removingMember = ref(false)

const editForm = ref({ name: '', description: '' })
const addMemberForm = ref({ user_id: '', role: 'member' })
const updateRoleForm = ref({ role: 'member' })
const selectedMember = ref(null)

const canManage = computed(() => {
  return workspace.value?.role === 'owner' || workspace.value?.role === 'admin'
})

const isOwner = computed(() => {
  return workspace.value?.role === 'owner'
})

const availableUsers = computed(() => {
  const memberIds = members.value.map(m => m.user_id)
  return allUsers.value.filter(u => !memberIds.includes(u.id))
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

async function loadWorkspace() {
  loading.value = true
  error.value = null

  try {
    const id = route.params.id
    const [wsRes, membersRes] = await Promise.all([
      workspaceApi.get(id),
      workspaceApi.listMembers(id)
    ])
    workspace.value = wsRes.data
    members.value = membersRes.data || []
    editForm.value = {
      name: workspace.value.name,
      description: workspace.value.description || ''
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadUsers() {
  try {
    const res = await userApi.list({ page_size: 100 })
    allUsers.value = res.data?.items || []
  } catch (e) {
    console.error('Failed to load users:', e)
  }
}

async function handleUpdate() {
  updating.value = true
  try {
    await workspaceApi.update(workspace.value.id, editForm.value)
    showEditModal.value = false
    loadWorkspace()
  } catch (e) {
    error.value = e.message
  } finally {
    updating.value = false
  }
}

async function handleDelete() {
  deleting.value = true
  try {
    await workspaceApi.delete(workspace.value.id)
    router.push('/workspaces')
  } catch (e) {
    error.value = e.message
  } finally {
    deleting.value = false
  }
}

async function handleAddMember() {
  addingMember.value = true
  try {
    await workspaceApi.addMember(workspace.value.id, addMemberForm.value)
    showAddMemberModal.value = false
    addMemberForm.value = { user_id: '', role: 'member' }
    loadWorkspace()
  } catch (e) {
    error.value = e.message
  } finally {
    addingMember.value = false
  }
}

function openUpdateRoleModal(member) {
  selectedMember.value = member
  updateRoleForm.value.role = member.role
  showUpdateRoleModal.value = true
}

async function handleUpdateRole() {
  updatingRole.value = true
  try {
    await workspaceApi.updateMember(workspace.value.id, selectedMember.value.user_id, updateRoleForm.value)
    showUpdateRoleModal.value = false
    loadWorkspace()
  } catch (e) {
    error.value = e.message
  } finally {
    updatingRole.value = false
  }
}

function confirmRemoveMember(member) {
  selectedMember.value = member
  showRemoveMemberModal.value = true
}

async function handleRemoveMember() {
  removingMember.value = true
  try {
    await workspaceApi.removeMember(workspace.value.id, selectedMember.value.user_id)
    showRemoveMemberModal.value = false
    loadWorkspace()
  } catch (e) {
    error.value = e.message
  } finally {
    removingMember.value = false
  }
}

onMounted(() => {
  loadWorkspace()
  loadUsers()
})
</script>

<style scoped>
.workspace-detail {
  max-width: 100%;
}

.back-link {
  margin-bottom: 1.5rem;
}

.link-back {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: #6B7280;
  text-decoration: none;
  font-size: 0.875rem;
  transition: color 0.2s;
}

.link-back:hover {
  color: #111827;
}

.link-back svg {
  width: 18px;
  height: 18px;
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

.workspace-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1.5rem;
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  margin-bottom: 1.5rem;
}

.header-info {
  display: flex;
  gap: 1rem;
}

.workspace-icon {
  width: 56px;
  height: 56px;
  background: #10B981;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.workspace-icon svg {
  width: 28px;
  height: 28px;
  color: white;
}

.workspace-name {
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.workspace-desc {
  font-size: 0.875rem;
  color: #6B7280;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
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

.section {
  background: white;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #E5E7EB;
  background: #F9FAFB;
}

.section-header h2 {
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin: 0;
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

.cell-username {
  font-weight: 500;
  color: #111827;
}

.cell-date {
  color: #6B7280;
  font-size: 0.875rem;
}

.cell-actions {
  text-align: right;
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

.btn svg {
  width: 18px;
  height: 18px;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.btn-sm svg {
  width: 16px;
  height: 16px;
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

.btn-danger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-danger-outline {
  background: white;
  color: #DC2626;
  border: 1px solid #DC2626;
}

.btn-danger-outline:hover {
  background: #FEE2E2;
}

.btn-ghost {
  background: transparent;
  color: #6B7280;
  padding: 0.375rem 0.75rem;
}

.btn-ghost:hover {
  background: #F3F4F6;
  color: #111827;
}

.btn-danger-text {
  color: #DC2626;
}

.btn-danger-text:hover {
  background: #FEE2E2;
  color: #991B1B;
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
.form-textarea,
.form-select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-textarea:focus,
.form-select:focus {
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
