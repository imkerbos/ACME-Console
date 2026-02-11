<template>
  <div class="settings-page">
    <!-- Site Settings -->
    <div class="settings-card">
      <h3>{{ $t('system.siteSettings') }}</h3>
      <p class="card-description">{{ $t('system.siteSettingsDesc') }}</p>

      <div v-if="successMessage" class="alert alert-success">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        {{ successMessage }}
      </div>

      <div v-if="error" class="alert alert-error">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M12 8v4M12 16h.01"/>
        </svg>
        {{ error }}
      </div>

      <form @submit.prevent="handleSaveSite" class="settings-form">
        <div class="form-group">
          <label class="form-label">{{ $t('system.siteTitle') }}</label>
          <input
            v-model="siteTitle"
            type="text"
            class="form-input"
            :placeholder="$t('system.siteTitlePlaceholder')"
          />
          <p class="form-hint">{{ $t('system.siteTitleHint') }}</p>
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('system.siteSubtitle') }}</label>
          <input
            v-model="siteSubtitle"
            type="text"
            class="form-input"
            :placeholder="$t('system.siteSubtitlePlaceholder')"
          />
          <p class="form-hint">{{ $t('system.siteSubtitleHint') }}</p>
        </div>

        <button type="submit" class="btn btn-primary" :disabled="submitting">
          <span v-if="submitting" class="btn-spinner"></span>
          {{ submitting ? $t('system.saving') : $t('common.save') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { settingApi } from '../api'
import { useSite } from '../stores/site'

const { t } = useI18n()
const site = useSite()

// Site settings
const siteTitle = ref('')
const siteSubtitle = ref('')
const submitting = ref(false)
const error = ref(null)
const successMessage = ref(null)

onMounted(async () => {
  await loadSiteSettings()
})

async function loadSiteSettings() {
  try {
    const res = await settingApi.getSite()
    siteTitle.value = res.data?.title || ''
    siteSubtitle.value = res.data?.subtitle || ''
  } catch (e) {
    error.value = e.message
  }
}

async function handleSaveSite() {
  error.value = null
  successMessage.value = null
  submitting.value = true

  try {
    await settingApi.updateSite({
      title: siteTitle.value,
      subtitle: siteSubtitle.value
    })
    // Update the global site store so sidebar/title update immediately
    site.update(siteTitle.value, siteSubtitle.value)
    successMessage.value = t('system.settingsSaved')
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.settings-page {
  max-width: 640px;
  margin: 0 auto;
}

.settings-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.settings-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.card-description {
  color: #6B7280;
  font-size: 0.875rem;
  margin: 0 0 1.5rem 0;
}

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
}

.alert svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.alert-success {
  background: #D1FAE5;
  color: #065F46;
  border: 1px solid #A7F3D0;
}

.alert-error {
  background: #FEE2E2;
  color: #991B1B;
  border: 1px solid #FECACA;
}

.settings-form {
  display: flex;
  flex-direction: column;
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

.form-input {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #10B981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.form-hint {
  margin: 0.375rem 0 0 0;
  font-size: 0.75rem;
  color: #9CA3AF;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  align-self: flex-start;
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

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
