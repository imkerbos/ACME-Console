<template>
  <div class="profile-page">
    <div class="profile-card">
      <div class="profile-header">
        <div class="avatar">
          {{ userInitial }}
        </div>
        <div class="profile-info">
          <h2>{{ user?.nickname || user?.username }}</h2>
          <p class="username">@{{ user?.username }}</p>
        </div>
      </div>

      <div class="profile-details">
        <div class="detail-item">
          <span class="detail-label">{{ $t('user.email') }}</span>
          <span class="detail-value">{{ user?.email || '-' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">{{ $t('user.role') }}</span>
          <span class="detail-value">
            <span class="role-badge">{{ user?.role === 'admin' ? $t('user.admin') : $t('user.regularUser') }}</span>
          </span>
        </div>
        <div class="detail-item">
          <span class="detail-label">{{ $t('user.lastLogin') }}</span>
          <span class="detail-value">{{ formatDate(user?.last_login) }}</span>
        </div>
      </div>
    </div>

    <div class="profile-edit-card">
      <h3>{{ $t('profile.editProfile') }}</h3>
      <p class="card-description">{{ $t('profile.editProfileDesc') }}</p>

      <div v-if="profileSuccessMessage" class="alert alert-success">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        {{ profileSuccessMessage }}
      </div>

      <div v-if="profileError" class="alert alert-error">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M12 8v4M12 16h.01"/>
        </svg>
        {{ profileError }}
      </div>

      <form @submit.prevent="handleUpdateProfile" class="profile-form">
        <div class="form-group">
          <label class="form-label">{{ $t('user.nickname') }}</label>
          <input
            v-model="editNickname"
            type="text"
            class="form-input"
            :placeholder="$t('user.nickname')"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('user.email') }}</label>
          <input
            v-model="editEmail"
            type="email"
            class="form-input"
            :placeholder="$t('user.email')"
          />
        </div>

        <button type="submit" class="btn btn-primary" :disabled="profileSubmitting">
          <span v-if="profileSubmitting" class="btn-spinner"></span>
          {{ profileSubmitting ? $t('profile.updatingProfile') : $t('profile.updateProfile') }}
        </button>
      </form>
    </div>

    <div class="password-card">
      <h3>{{ $t('profile.changePassword') }}</h3>
      <p class="card-description">{{ $t('profile.changePasswordDesc') }}</p>

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

      <form @submit.prevent="handleChangePassword" class="password-form">
        <div class="form-group">
          <label class="form-label">{{ $t('profile.currentPassword') }}</label>
          <input
            v-model="oldPassword"
            type="password"
            class="form-input"
            :placeholder="$t('profile.currentPassword')"
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('profile.newPassword') }}</label>
          <input
            v-model="newPassword"
            type="password"
            class="form-input"
            :placeholder="$t('profile.newPassword')"
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('profile.confirmNewPassword') }}</label>
          <input
            v-model="confirmPassword"
            type="password"
            class="form-input"
            :placeholder="$t('profile.confirmNewPassword')"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary" :disabled="submitting">
          <span v-if="submitting" class="btn-spinner"></span>
          {{ submitting ? $t('profile.updatingPassword') : $t('profile.updatePassword') }}
        </button>
      </form>
    </div>

    <div class="notification-card">
      <h3>{{ $t('notification.personalCertNotification') }}</h3>
      <p class="card-description">{{ $t('notification.personalCertNotificationDesc') }}</p>
      <WebhookConfig />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../stores/auth'
import { authApi } from '../api'
import WebhookConfig from '../components/WebhookConfig.vue'

const { t } = useI18n()
const { getUser, setUser } = useAuth()

const user = computed(() => getUser())
const userInitial = computed(() => {
  const name = user.value?.nickname || user.value?.username || 'U'
  return name.charAt(0).toUpperCase()
})

// Profile edit form
const editNickname = ref('')
const editEmail = ref('')
const profileSubmitting = ref(false)
const profileError = ref(null)
const profileSuccessMessage = ref(null)

// Initialize edit form with current user data
watch(user, (newUser) => {
  if (newUser) {
    editNickname.value = newUser.nickname || ''
    editEmail.value = newUser.email || ''
  }
}, { immediate: true })

async function handleUpdateProfile() {
  profileError.value = null
  profileSuccessMessage.value = null

  profileSubmitting.value = true

  try {
    const data = {}
    if (editNickname.value) data.nickname = editNickname.value
    if (editEmail.value) data.email = editEmail.value

    const response = await authApi.updateProfile(data)
    profileSuccessMessage.value = t('profile.profileUpdated')

    // Update user in store
    if (response.data) {
      setUser(response.data)
    }
  } catch (e) {
    profileError.value = e.message
  } finally {
    profileSubmitting.value = false
  }
}

// Password change form
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const submitting = ref(false)
const error = ref(null)
const successMessage = ref(null)

async function handleChangePassword() {
  error.value = null
  successMessage.value = null

  if (newPassword.value !== confirmPassword.value) {
    error.value = t('user.passwordMismatch')
    return
  }

  if (newPassword.value.length < 6) {
    error.value = t('validation.minLength', { min: 6 })
    return
  }

  submitting.value = true

  try {
    await authApi.changePassword(oldPassword.value, newPassword.value)
    successMessage.value = t('profile.passwordUpdated')
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}

function formatDate(dateStr) {
  if (!dateStr) return t('common.no')
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.profile-page {
  max-width: 640px;
  margin: 0 auto;
}

.profile-card,
.password-card,
.profile-edit-card,
.notification-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #E5E7EB;
  margin-bottom: 1.5rem;
}

.avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #10B981 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.75rem;
  flex-shrink: 0;
}

.profile-info h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.username {
  color: #6B7280;
  font-size: 0.875rem;
  margin: 0;
}

.profile-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.detail-item {
  display: flex;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid #F3F4F6;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-label {
  width: 120px;
  color: #6B7280;
  font-size: 0.875rem;
}

.detail-value {
  flex: 1;
  color: #111827;
  font-size: 0.875rem;
}

.role-badge {
  display: inline-flex;
  padding: 0.25rem 0.625rem;
  background: #EEF2FF;
  color: #4F46E5;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.password-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.profile-edit-card h3 {
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

.password-form {
  display: flex;
  flex-direction: column;
}

.profile-form {
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
