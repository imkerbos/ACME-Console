<template>
  <div class="certificate-detail">
    <div class="back-link">
      <router-link to="/certificates" class="link-back">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        {{ $t('certificate.backToCertificates') }}
      </router-link>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <p class="loading-text">{{ $t('certificate.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="alert alert-error">
      <svg class="alert-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <path d="M12 8v4M12 16h.01"/>
      </svg>
      <span>{{ error }}</span>
    </div>

    <template v-else-if="certificate">
      <!-- Header Card -->
      <div class="header-card">
        <div class="header-content">
          <div class="header-main">
            <div class="cert-icon" :class="`icon-${certificate.status}`">
              <svg v-if="certificate.status === 'ready'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <svg v-else-if="certificate.status === 'pending'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v6l4 2"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 8v4M12 16h.01"/>
              </svg>
            </div>
            <div class="header-info">
              <h1>{{ $t('certificate.certificateId', { id }) }}</h1>
              <div class="header-meta">
                <span :class="['status-badge', `status-${certificate.status}`]">
                  <span class="status-dot"></span>
                  {{ $t(`certificate.${certificate.status}`) }}
                </span>
                <span class="meta-divider"></span>
                <span class="key-type">{{ certificate.key_type }}</span>
                <span class="meta-divider"></span>
                <span class="created-at">{{ $t('certificate.created') }} {{ formatDate(certificate.created_at) }}</span>
              </div>
            </div>
          </div>
          <div class="header-actions">
            <button
              v-if="certificate.status === 'pending'"
              class="btn btn-secondary"
              :disabled="checking"
              @click="handlePreVerify"
            >
              <span v-if="checking" class="btn-spinner"></span>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
              </svg>
              {{ checking ? '检查中...' : '检查 DNS' }}
            </button>
            <button
              v-if="certificate.status === 'pending'"
              class="btn btn-primary"
              :disabled="verifying"
              @click="handleVerify"
            >
              <span v-if="verifying" class="btn-spinner"></span>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              {{ verifying ? $t('challenge.verifying') : $t('challenge.verifyAndIssue') }}
            </button>
            <button
              v-if="certificate.status === 'ready'"
              class="btn btn-primary"
              @click="handleDownload"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
              </svg>
              {{ $t('common.download') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Domains Card -->
      <div class="detail-card">
        <div class="card-title">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
          </svg>
          {{ $t('certificate.domains') }}
        </div>
        <div class="domain-list">
          <div v-for="domain in parseDomains(certificate.domains)" :key="domain" class="domain-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
            </svg>
            <span class="domain-name">{{ domain }}</span>
            <span v-if="domain.startsWith('*.')" class="domain-tag wildcard">{{ $t('certificate.wildcard') }}</span>
          </div>
        </div>
      </div>

      <!-- Timeline Card -->
      <div class="detail-card">
        <div class="card-title">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v6l4 2"/>
          </svg>
          {{ $t('certificate.timeline') }}
        </div>
        <div class="timeline">
          <div class="timeline-item completed">
            <div class="timeline-marker"></div>
            <div class="timeline-content">
              <span class="timeline-label">{{ $t('certificate.created') }}</span>
              <span class="timeline-date">{{ formatDateTime(certificate.created_at) }}</span>
            </div>
          </div>
          <div class="timeline-item" :class="{ completed: certificate.issued_at, pending: !certificate.issued_at && certificate.status === 'pending' }">
            <div class="timeline-marker"></div>
            <div class="timeline-content">
              <span class="timeline-label">{{ $t('certificate.issued') }}</span>
              <span class="timeline-date">{{ certificate.issued_at ? formatDateTime(certificate.issued_at) : $t('certificate.pendingVerification') }}</span>
            </div>
          </div>
          <div class="timeline-item" :class="{ completed: certificate.expires_at, future: certificate.expires_at }">
            <div class="timeline-marker"></div>
            <div class="timeline-content">
              <span class="timeline-label">{{ $t('certificate.expires') }}</span>
              <span class="timeline-date" :class="{ warning: isExpiringSoon }">
                {{ certificate.expires_at ? formatDateTime(certificate.expires_at) : '-' }}
                <span v-if="daysUntilExpiry !== null" class="expiry-days">
                  {{ $t('certificate.daysLeft', { days: daysUntilExpiry }) }}
                </span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- DNS Challenges Card -->
      <div v-if="certificate.challenges?.length > 0" class="detail-card">
        <div class="card-header-flex">
          <div class="card-title">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
            </svg>
            {{ $t('challenge.dnsChallenge') }}
          </div>
          <button class="btn btn-secondary btn-sm" @click="copyAllChallenges">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
              <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
            </svg>
            {{ $t('common.copyAll') }}
          </button>
        </div>

        <div v-if="certificate.status === 'pending'" class="challenge-hint">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 16v-4M12 8h.01"/>
          </svg>
          {{ $t('challenge.challengeHint') }}
        </div>

        <div class="challenge-list">
          <div v-for="challenge in certificate.challenges" :key="challenge.id" class="challenge-item">
            <div class="challenge-header">
              <span class="challenge-domain">{{ challenge.domain }}</span>
              <span :class="['status-badge', `status-${challenge.status}`]">
                <span class="status-dot"></span>
                {{ $t(`certificate.${challenge.status}`) }}
              </span>
            </div>
            <div class="challenge-records">
              <div class="record-row">
                <span class="record-label">{{ $t('challenge.txtHost') }}</span>
                <code class="record-value">{{ challenge.txt_host }}</code>
                <button class="copy-btn" @click="copyToClipboard(challenge.txt_host)" :title="$t('common.copy')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                    <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                  </svg>
                </button>
              </div>
              <div class="record-row">
                <span class="record-label">{{ $t('challenge.txtValue') }}</span>
                <code class="record-value">{{ challenge.txt_value }}</code>
                <button class="copy-btn" @click="copyToClipboard(challenge.txt_value)" :title="$t('common.copy')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                    <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- DNS Check Results Modal -->
      <div v-if="showDNSCheckModal" class="modal-overlay" @click="showDNSCheckModal = false">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <div class="modal-title">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
              </svg>
              DNS 检查结果
            </div>
            <button class="modal-close" @click="showDNSCheckModal = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <div v-if="allDNSMatched" class="alert alert-success">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <span>所有 DNS 记录验证通过！可以点击"验证并签发"按钮了。</span>
            </div>

            <div v-else class="alert alert-warning">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <span>部分 DNS 记录未正确配置，请检查后重试。</span>
            </div>

            <div class="dns-check-list">
              <div v-for="(result, index) in dnsCheckResults" :key="index" class="dns-check-item">
                <div class="dns-check-header">
                  <span class="dns-check-domain">{{ result.domain }}</span>
                  <span :class="['status-badge', result.matched ? 'status-verified' : 'status-failed']">
                    <span class="status-dot"></span>
                    {{ result.matched ? '已匹配' : '未匹配' }}
                  </span>
                </div>
                <div class="dns-check-details">
                  <div class="dns-check-row">
                    <span class="dns-check-label">TXT 主机:</span>
                    <code class="dns-check-value">{{ result.txt_host }}</code>
                  </div>
                  <div class="dns-check-row">
                    <span class="dns-check-label">期望值:</span>
                    <code class="dns-check-value">{{ result.expected_value }}</code>
                  </div>
                  <div v-if="result.found_values && result.found_values.length > 0" class="dns-check-row">
                    <span class="dns-check-label">查询到:</span>
                    <code class="dns-check-value">{{ result.found_values.join(', ') }}</code>
                  </div>
                  <div v-else class="dns-check-row">
                    <span class="dns-check-label">查询到:</span>
                    <span class="dns-check-value-empty">未查询到 TXT 记录</span>
                  </div>
                  <div v-if="result.error" class="dns-check-row error">
                    <span class="dns-check-label">错误:</span>
                    <span class="dns-check-error">{{ result.error }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showDNSCheckModal = false">
              关闭
            </button>
            <button v-if="allDNSMatched" class="btn btn-primary" @click="showDNSCheckModal = false; handleVerify()">
              验证并签发
            </button>
          </div>
        </div>
      </div>

      <!-- Download Card (when ready) -->
      <div v-if="certificate.status === 'ready'" class="detail-card success-card">
        <div class="success-header">
          <div class="success-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <div class="success-content">
            <h3>{{ $t('download.certificateReady') }}</h3>
            <p>{{ $t('download.certificateReadyDesc') }}</p>
          </div>
        </div>
        <div class="download-options">
          <button class="btn btn-primary" @click="handleDownload">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
            </svg>
            {{ $t('download.downloadBundle') }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { certificateApi } from '../api'

const route = useRoute()
const { t } = useI18n()
const id = route.params.id

const certificate = ref(null)
const loading = ref(true)
const error = ref(null)
const verifying = ref(false)
const checking = ref(false)
const dnsCheckResults = ref(null)
const showDNSCheckModal = ref(false)

const allDNSMatched = computed(() => {
  if (!dnsCheckResults.value) return false
  return dnsCheckResults.value.every(r => r.matched)
})

const isExpiringSoon = computed(() => {
  if (!certificate.value?.expires_at) return false
  const expires = new Date(certificate.value.expires_at)
  const now = new Date()
  const daysUntilExpiry = (expires - now) / (1000 * 60 * 60 * 24)
  return daysUntilExpiry < 30
})

const daysUntilExpiry = computed(() => {
  if (!certificate.value?.expires_at) return null
  const expires = new Date(certificate.value.expires_at)
  const now = new Date()
  return Math.ceil((expires - now) / (1000 * 60 * 60 * 24))
})

async function loadCertificate() {
  loading.value = true
  error.value = null

  try {
    const response = await certificateApi.get(id)
    certificate.value = response.data
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleVerify() {
  verifying.value = true
  error.value = null

  try {
    const response = await certificateApi.verify(id)
    certificate.value = response.data
    // Clear DNS check results after verification
    dnsCheckResults.value = null
  } catch (e) {
    error.value = e.message
  } finally {
    verifying.value = false
  }
}

async function handlePreVerify() {
  checking.value = true
  error.value = null

  try {
    const response = await certificateApi.preVerify(id)
    dnsCheckResults.value = response.data.results || []
    showDNSCheckModal.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    checking.value = false
  }
}

async function handleDownload() {
  try {
    const response = await certificateApi.download(id, 'zip')

    // Get domains for filename
    const domains = parseDomains(certificate.value.domains)
    const primaryDomain = domains[0] || 'certificate'
    const filename = `${primaryDomain}.zip`

    // Create download link for blob
    const blob = new Blob([response.data], { type: 'application/zip' })
    const url = URL.createObjectURL(blob)

    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()

    URL.revokeObjectURL(url)
  } catch (e) {
    error.value = e.message
  }
}

function copyAllChallenges() {
  if (!certificate.value?.challenges) return

  const text = certificate.value.challenges.map(c =>
    `# ${c.domain}\n${c.txt_host}. 300 IN TXT "${c.txt_value}"`
  ).join('\n\n')

  copyToClipboard(text)
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text).then(() => {
    // Could add a toast notification here
  })
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

function formatDateTime(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  loadCertificate()
})
</script>

<style scoped>
.certificate-detail {
  max-width: 800px;
  margin: 0 auto;
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
  width: 32px;
  height: 32px;
  border: 3px solid #E5E7EB;
  border-top-color: #10B981;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: #6B7280;
  font-size: 0.875rem;
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
  background: #FEE2E2;
  color: #991B1B;
  border: 1px solid #FECACA;
}

.alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

/* Header Card */
.header-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1.5rem;
}

.header-main {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.cert-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cert-icon svg {
  width: 28px;
  height: 28px;
}

.icon-ready {
  background: #D1FAE5;
  color: #059669;
}

.icon-pending {
  background: #FEF3C7;
  color: #D97706;
}

.icon-failed {
  background: #FEE2E2;
  color: #DC2626;
}

.header-info h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.5rem 0;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.meta-divider {
  width: 4px;
  height: 4px;
  background: #D1D5DB;
  border-radius: 50%;
}

.key-type,
.created-at {
  color: #6B7280;
  font-size: 0.875rem;
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

.status-verified {
  background: #DBEAFE;
  color: #1E40AF;
}

.status-verified .status-dot {
  background: #3B82F6;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

/* Detail Card */
.detail-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 1rem;
}

.card-title svg {
  width: 20px;
  height: 20px;
  color: #6B7280;
}

.card-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header-flex .card-title {
  margin-bottom: 0;
}

/* Domain List */
.domain-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.domain-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: #F9FAFB;
  border-radius: 8px;
}

.domain-item svg {
  width: 18px;
  height: 18px;
  color: #6B7280;
  flex-shrink: 0;
}

.domain-name {
  font-family: monospace;
  font-size: 0.875rem;
  color: #111827;
  flex: 1;
}

.domain-tag {
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.domain-tag.wildcard {
  background: #EEF2FF;
  color: #4F46E5;
}

/* Timeline */
.timeline {
  position: relative;
  padding-left: 1.5rem;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 7px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: #E5E7EB;
}

.timeline-item {
  position: relative;
  padding-bottom: 1.25rem;
}

.timeline-item:last-child {
  padding-bottom: 0;
}

.timeline-marker {
  position: absolute;
  left: -1.5rem;
  top: 4px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: white;
  border: 2px solid #E5E7EB;
}

.timeline-item.completed .timeline-marker {
  background: #10B981;
  border-color: #10B981;
}

.timeline-item.pending .timeline-marker {
  background: #F59E0B;
  border-color: #F59E0B;
}

.timeline-item.future .timeline-marker {
  border-color: #D1D5DB;
}

.timeline-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.timeline-label {
  font-weight: 500;
  color: #374151;
  font-size: 0.875rem;
}

.timeline-date {
  color: #6B7280;
  font-size: 0.875rem;
}

.timeline-date.warning {
  color: #D97706;
  font-weight: 500;
}

.expiry-days {
  font-weight: normal;
}

/* Challenge List */
.challenge-hint {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: #FEF3C7;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.875rem;
  color: #92400E;
}

.challenge-hint svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.challenge-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.challenge-item {
  border: 1px solid #E5E7EB;
  border-radius: 12px;
  overflow: hidden;
}

.challenge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: #F9FAFB;
  border-bottom: 1px solid #E5E7EB;
}

.challenge-domain {
  font-weight: 500;
  color: #111827;
  font-size: 0.875rem;
}

.challenge-records {
  padding: 1rem;
}

.record-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
}

.record-row:first-child {
  padding-top: 0;
}

.record-row:last-child {
  padding-bottom: 0;
}

.record-label {
  min-width: 80px;
  color: #6B7280;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  white-space: nowrap;
  flex-shrink: 0;
}

.record-value {
  flex: 1;
  font-family: 'Monaco', 'Menlo', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  color: #111827;
  background: #F3F4F6;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  white-space: nowrap;
  overflow-x: auto;
  line-height: 1.5;
  /* 美化滚动条 */
  scrollbar-width: thin;
  scrollbar-color: #D1D5DB #F3F4F6;
}

.record-value::-webkit-scrollbar {
  height: 6px;
}

.record-value::-webkit-scrollbar-track {
  background: #F3F4F6;
  border-radius: 3px;
}

.record-value::-webkit-scrollbar-thumb {
  background: #D1D5DB;
  border-radius: 3px;
}

.record-value::-webkit-scrollbar-thumb:hover {
  background: #9CA3AF;
}

.copy-btn {
  padding: 0.375rem;
  background: none;
  border: none;
  color: #9CA3AF;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #F3F4F6;
  color: #374151;
}

.copy-btn svg {
  width: 16px;
  height: 16px;
}

/* Success Card */
.success-card {
  background: linear-gradient(135deg, #F0FDF4 0%, #ECFDF5 100%);
  border-color: #A7F3D0;
}

.success-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.success-icon {
  width: 48px;
  height: 48px;
  background: #10B981;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.success-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.success-content h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #065F46;
  margin: 0 0 0.25rem 0;
}

.success-content p {
  color: #047857;
  font-size: 0.875rem;
  margin: 0;
}

.download-options {
  display: flex;
  gap: 0.75rem;
}

/* Buttons */
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
  background: white;
  color: #374151;
  border: 1px solid #E5E7EB;
}

.btn-secondary:hover {
  background: #F9FAFB;
  border-color: #D1D5DB;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.btn-sm svg {
  width: 16px;
  height: 16px;
}

.btn-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@media (max-width: 640px) {
  .header-content {
    flex-direction: column;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .btn {
    flex: 1;
    justify-content: center;
  }

  .timeline-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
  }
}

/* DNS Check Results */
.dns-check-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.dns-check-item {
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  padding: 1rem;
}

.dns-check-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #E5E7EB;
}

.dns-check-domain {
  font-weight: 600;
  color: #111827;
  font-size: 0.875rem;
}

.dns-check-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.dns-check-row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  font-size: 0.8125rem;
}

.dns-check-row.error {
  color: #DC2626;
}

.dns-check-label {
  min-width: 70px;
  color: #6B7280;
  flex-shrink: 0;
}

.dns-check-value {
  flex: 1;
  background: white;
  padding: 0.375rem 0.5rem;
  border-radius: 4px;
  border: 1px solid #E5E7EB;
  font-family: monospace;
  font-size: 0.75rem;
  word-break: break-all;
}

.dns-check-value-empty {
  color: #9CA3AF;
  font-style: italic;
}

.dns-check-error {
  color: #DC2626;
  font-size: 0.75rem;
}

.alert-warning {
  background: #FEF3C7;
  color: #92400E;
  border: 1px solid #FDE68A;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.alert-warning svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.alert-success {
  background: #D1FAE5;
  color: #065F46;
  border: 1px solid #A7F3D0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.alert-success svg {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.status-failed {
  background: #FEE2E2;
  color: #991B1B;
}

.status-failed .status-dot {
  background: #DC2626;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-secondary {
  background: #F3F4F6;
  color: #374151;
}

.btn-secondary:hover:not(:disabled) {
  background: #E5E7EB;
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Modal Styles */
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
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 16px;
  max-width: 700px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
  border-bottom: 1px solid #E5E7EB;
}

.modal-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
}

.modal-title svg {
  width: 24px;
  height: 24px;
  color: #10B981;
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: #6B7280;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
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
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid #E5E7EB;
}
</style>
