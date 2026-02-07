<template>
  <div class="certificate-list">
    <!-- Action Bar -->
    <div class="action-bar">
      <div class="filters">
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <path d="M21 21l-4.35-4.35"/>
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            :placeholder="$t('certificate.searchPlaceholder')"
          />
        </div>
        <select v-model="workspaceFilter" class="workspace-filter" @change="loadCertificates()">
          <option value="">{{ $t('certificate.allCertificates') }}</option>
          <option v-for="ws in workspaces" :key="ws.id" :value="ws.id">
            {{ ws.name }}
          </option>
        </select>
      </div>
      <router-link to="/certificates/new" class="btn btn-primary">
        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 5v14M5 12h14"/>
        </svg>
        {{ $t('certificate.newCertificate') }}
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

    <!-- Certificate Table -->
    <div v-else class="table-card">
      <table v-if="filteredCertificates.length > 0" class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ $t('certificate.domains') }}</th>
            <th>{{ $t('certificate.keyType') }}</th>
            <th>{{ $t('certificate.status') }}</th>
            <th>{{ $t('certificate.created') }}</th>
            <th>{{ $t('certificate.expires') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cert in filteredCertificates" :key="cert.id">
            <td class="cell-id">#{{ cert.id }}</td>
            <td>
              <div class="domain-list">
                <span v-for="domain in parseDomains(cert.domains).slice(0, 2)" :key="domain" class="domain-tag">
                  {{ domain }}
                </span>
                <span v-if="parseDomains(cert.domains).length > 2" class="domain-more">
                  +{{ parseDomains(cert.domains).length - 2 }} {{ $t('certificate.more') }}
                </span>
              </div>
            </td>
            <td>
              <span class="key-type-badge">{{ cert.key_type }}</span>
            </td>
            <td>
              <span :class="['status-badge', `status-${cert.status}`]">
                <span class="status-dot"></span>
                {{ $t(`certificate.${cert.status}`) }}
              </span>
            </td>
            <td class="cell-date">{{ formatDate(cert.created_at) }}</td>
            <td class="cell-date">
              <span v-if="cert.expires_at" :class="{ 'text-warning': isExpiringSoon(cert.expires_at) }">
                {{ formatDate(cert.expires_at) }}
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="cell-actions">
              <router-link :to="`/certificates/${cert.id}`" class="btn btn-ghost btn-sm">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                  <path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                </svg>
                {{ $t('common.view') }}
              </router-link>
              <button class="btn btn-ghost btn-sm btn-danger-text" @click="confirmDelete(cert)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3,6 5,6 21,6"/>
                  <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                </svg>
                {{ $t('common.delete') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
        </div>
        <h3 class="empty-title">{{ $t('certificate.noCertificates') }}</h3>
        <p class="empty-description">{{ $t('certificate.noCertificatesDesc') }}</p>
        <router-link to="/certificates/new" class="btn btn-primary">
          {{ $t('certificate.createCertificate') }}
        </router-link>
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
          <p>{{ $t('certificate.deleteConfirm') }}</p>
          <div class="delete-cert-info">
            <span v-for="domain in parseDomains(deleteCert?.domains || '[]').slice(0, 2)" :key="domain" class="domain-tag">
              {{ domain }}
            </span>
            <span v-if="parseDomains(deleteCert?.domains || '[]').length > 2" class="domain-more">
              +{{ parseDomains(deleteCert?.domains || '[]').length - 2 }} {{ $t('certificate.more') }}
            </span>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">
              {{ $t('common.cancel') }}
            </button>
            <button type="button" class="btn btn-danger" :disabled="deleting" @click="handleDelete">
              {{ deleting ? $t('common.loading') : $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { certificateApi, workspaceApi } from '../api'

const { t } = useI18n()
const route = useRoute()

const certificates = ref([])
const workspaces = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const workspaceFilter = ref('')
const showDeleteModal = ref(false)
const deleteCert = ref(null)
const deleting = ref(false)
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0,
  totalPages: 0
})

const filteredCertificates = computed(() => {
  if (!searchQuery.value) return certificates.value
  const query = searchQuery.value.toLowerCase()
  return certificates.value.filter(cert => {
    const domains = parseDomains(cert.domains)
    return domains.some(d => d.toLowerCase().includes(query))
  })
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

async function loadWorkspaces() {
  try {
    const response = await workspaceApi.list()
    workspaces.value = response.data || []
  } catch (e) {
    console.error('Failed to load workspaces:', e)
  }
}

async function loadCertificates(page = 1) {
  loading.value = true
  error.value = null

  try {
    const params = { page, page_size: 20 }
    if (workspaceFilter.value) {
      params.workspace_id = workspaceFilter.value
    }

    const response = await certificateApi.list(params)
    const data = response.data

    certificates.value = data.items || []
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
  loadCertificates(page)
}

function parseDomains(domainsStr) {
  try {
    return JSON.parse(domainsStr)
  } catch {
    return [domainsStr]
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

function isExpiringSoon(dateStr) {
  if (!dateStr) return false
  const expires = new Date(dateStr)
  const now = new Date()
  const daysUntilExpiry = (expires - now) / (1000 * 60 * 60 * 24)
  return daysUntilExpiry < 30
}

function confirmDelete(cert) {
  deleteCert.value = cert
  showDeleteModal.value = true
}

async function handleDelete() {
  deleting.value = true

  try {
    await certificateApi.delete(deleteCert.value.id)
    showDeleteModal.value = false
    loadCertificates(pagination.value.page)
  } catch (e) {
    error.value = e.message
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  // Check if workspace_id is in query params
  if (route.query.workspace_id) {
    workspaceFilter.value = route.query.workspace_id
  }
  loadWorkspaces()
  loadCertificates()
})
</script>

<style scoped>
.certificate-list {
  max-width: 100%;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

.filters {
  display: flex;
  gap: 1rem;
  flex: 1;
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 320px;
}

.workspace-filter {
  padding: 0.625rem 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  background: white;
  min-width: 180px;
  cursor: pointer;
  transition: all 0.2s;
}

.workspace-filter:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
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

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
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

.domain-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.domain-tag {
  background: #EEF2FF;
  color: #4F46E5;
  padding: 0.25rem 0.625rem;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-family: monospace;
}

.domain-more {
  color: #6B7280;
  font-size: 0.8125rem;
  padding: 0.25rem 0;
}

.key-type-badge {
  background: #F3F4F6;
  color: #374151;
  padding: 0.25rem 0.625rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-pending {
  background: #FEF3C7;
  color: #92400E;
}

.status-pending .status-dot {
  background: #F59E0B;
}

.status-ready {
  background: #D1FAE5;
  color: #065F46;
}

.status-ready .status-dot {
  background: #10B981;
}

.status-failed {
  background: #FEE2E2;
  color: #991B1B;
}

.status-failed .status-dot {
  background: #EF4444;
}

.cell-date {
  color: #6B7280;
  font-size: 0.875rem;
  white-space: nowrap;
}

.text-warning {
  color: #D97706;
  font-weight: 500;
}

.text-muted {
  color: #9CA3AF;
}

.cell-actions {
  text-align: right;
}

.btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: transparent;
  color: #6B7280;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.2s;
}

.btn-ghost:hover {
  background: #F3F4F6;
  color: #111827;
}

.btn-ghost svg {
  width: 16px;
  height: 16px;
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
  margin-bottom: 0.5rem;
}

.empty-description {
  color: #6B7280;
  font-size: 0.875rem;
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

.btn-danger-text {
  color: #DC2626;
}

.btn-danger-text:hover {
  background: #FEE2E2;
  color: #991B1B;
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

.delete-cert-info {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  margin-top: 0.75rem;
  padding: 0.75rem;
  background: #FEE2E2;
  border-radius: 8px;
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
</style>
