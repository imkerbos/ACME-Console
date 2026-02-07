<template>
  <div class="certificate-create">
    <div class="back-link">
      <router-link to="/certificates" class="link-back">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        {{ $t('certificate.backToCertificates') }}
      </router-link>
    </div>

    <div class="create-card">
      <div class="card-header">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
        </div>
        <div class="header-text">
          <h2>{{ $t('certificate.newCertificate') }}</h2>
          <p>{{ $t('certificate.createCertificateDesc') }}</p>
        </div>
      </div>

      <div v-if="error" class="alert alert-error">
        <svg class="alert-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M12 8v4M12 16h.01"/>
        </svg>
        <span>{{ error }}</span>
      </div>

      <form @submit.prevent="handleSubmit" class="create-form">
        <div class="form-section">
          <label class="form-label">
            <span class="label-text">{{ $t('certificate.domains') }}</span>
            <span class="label-required">*</span>
          </label>
          <div class="domain-input-wrapper">
            <textarea
              v-model="domainsText"
              class="form-textarea"
              :placeholder="$t('certificate.domainsPlaceholder')"
              required
            ></textarea>
            <div class="domain-count" v-if="domainCount > 0">
              {{ $t('certificate.domainCount', { count: domainCount }) }}
            </div>
          </div>
          <p class="form-hint">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            {{ $t('certificate.domainsHint', { code: '*.domain.com' }) }}
          </p>
        </div>

        <div class="form-section">
          <label class="form-label">
            <span class="label-text">{{ $t('user.email') }}</span>
            <span class="label-required">*</span>
          </label>
          <input
            v-model="email"
            type="email"
            class="form-input"
            :placeholder="$t('user.email')"
            required
          />
          <p class="form-hint">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            用于接收证书相关通知和 Let's Encrypt 账户注册
          </p>
        </div>

        <div class="form-section">
          <label class="form-label">
            <span class="label-text">{{ $t('certificate.keyType') }}</span>
          </label>
          <div class="key-type-options">
            <label class="key-type-option" :class="{ selected: keyType === 'RSA' }">
              <input type="radio" v-model="keyType" value="RSA" />
              <div class="option-content">
                <div class="option-header">
                  <span class="option-title">RSA</span>
                  <span class="option-badge">2048-bit</span>
                </div>
                <p class="option-description">{{ $t('certificate.rsaDesc') }}</p>
              </div>
            </label>
            <label class="key-type-option" :class="{ selected: keyType === 'ECC' }">
              <input type="radio" v-model="keyType" value="ECC" />
              <div class="option-content">
                <div class="option-header">
                  <span class="option-title">ECC</span>
                  <span class="option-badge recommended">P-256</span>
                </div>
                <p class="option-description">{{ $t('certificate.eccDesc') }}</p>
              </div>
            </label>
          </div>
        </div>

        <div class="form-section">
          <label class="form-label">
            <span class="label-text">{{ $t('certificate.workspace') }}</span>
          </label>
          <select v-model="workspaceId" class="form-select">
            <option :value="null">{{ $t('certificate.privateWorkspace') }}</option>
            <option v-for="ws in workspaces" :key="ws.id" :value="ws.id">
              {{ ws.name }}
            </option>
          </select>
          <p class="form-hint">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            {{ $t('certificate.workspaceOptional') }}
          </p>
        </div>

        <div class="form-actions">
          <router-link to="/certificates" class="btn btn-secondary">
            {{ $t('common.cancel') }}
          </router-link>
          <button type="submit" class="btn btn-primary" :disabled="submitting">
            <span v-if="submitting" class="btn-spinner"></span>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 5v14M5 12h14"/>
            </svg>
            {{ submitting ? $t('certificate.creating') : $t('certificate.createCertificate') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../stores/auth'
import { certificateApi, workspaceApi } from '../api'

const router = useRouter()
const { t } = useI18n()
const { getUser } = useAuth()

const domainsText = ref('')
const email = ref('')
const keyType = ref('RSA')
const workspaceId = ref(null)
const workspaces = ref([])
const submitting = ref(false)
const error = ref(null)

// Initialize email with current user's email
onMounted(async () => {
  const user = getUser()
  if (user?.email) {
    email.value = user.email
  }

  // Load workspaces
  try {
    const response = await workspaceApi.list()
    workspaces.value = response.data || []
  } catch (e) {
    console.error('Failed to load workspaces:', e)
  }
})

const domainCount = computed(() => {
  return domainsText.value
    .split('\n')
    .map(d => d.trim())
    .filter(d => d.length > 0).length
})

async function handleSubmit() {
  error.value = null

  const domains = domainsText.value
    .split('\n')
    .map(d => d.trim())
    .filter(d => d.length > 0)

  if (domains.length === 0) {
    error.value = t('certificate.domainsRequired')
    return
  }

  submitting.value = true

  try {
    const payload = {
      domains,
      email: email.value,
      key_type: keyType.value
    }

    if (workspaceId.value) {
      payload.workspace_id = parseInt(workspaceId.value)
    }

    const response = await certificateApi.create(payload)

    router.push(`/certificates/${response.data.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.certificate-create {
  max-width: 640px;
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

.create-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, #F0FDF4 0%, #ECFDF5 100%);
  border-bottom: 1px solid #E5E7EB;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: #10B981;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.header-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.header-text h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.header-text p {
  font-size: 0.875rem;
  color: #6B7280;
  margin: 0;
}

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  margin: 1.5rem 1.5rem 0;
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

.create-form {
  padding: 1.5rem;
}

.form-section {
  margin-bottom: 1.5rem;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
}

.label-text {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.label-required {
  color: #EF4444;
}

.domain-input-wrapper {
  position: relative;
}

.form-textarea {
  width: 100%;
  min-height: 140px;
  padding: 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  font-family: monospace;
  resize: vertical;
  transition: all 0.2s;
}

.form-input {
  width: 100%;
  padding: 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-select {
  width: 100%;
  padding: 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  background: white;
  transition: all 0.2s;
  cursor: pointer;
}

.form-textarea:focus,
.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.domain-count {
  position: absolute;
  right: 0.75rem;
  bottom: 0.75rem;
  background: #F3F4F6;
  color: #6B7280;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
}

.form-hint {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-top: 0.5rem;
  font-size: 0.8125rem;
  color: #6B7280;
}

.form-hint svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.form-hint code {
  background: #F3F4F6;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-size: 0.8125rem;
}

.key-type-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.key-type-option {
  position: relative;
  cursor: pointer;
}

.key-type-option input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.option-content {
  padding: 1rem;
  border: 2px solid #E5E7EB;
  border-radius: 12px;
  transition: all 0.2s;
}

.key-type-option:hover .option-content {
  border-color: #D1D5DB;
}

.key-type-option.selected .option-content {
  border-color: #10B981;
  background: #F0FDF4;
}

.option-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.375rem;
}

.option-title {
  font-weight: 600;
  color: #111827;
}

.option-badge {
  background: #F3F4F6;
  color: #6B7280;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.option-badge.recommended {
  background: #D1FAE5;
  color: #065F46;
}

.option-description {
  font-size: 0.8125rem;
  color: #6B7280;
  margin: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid #E5E7EB;
  margin-top: 1.5rem;
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

.btn-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 480px) {
  .key-type-options {
    grid-template-columns: 1fr;
  }
}
</style>
