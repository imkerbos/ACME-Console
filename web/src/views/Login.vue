<template>
  <div class="login-page">
    <div class="login-container">
      <!-- Language Switcher -->
      <div class="lang-switcher-top">
        <button
          :class="['lang-btn', { active: currentLocale === 'zh-CN' }]"
          @click="setLocale('zh-CN')"
        >
          中文
        </button>
        <button
          :class="['lang-btn', { active: currentLocale === 'en-US' }]"
          @click="setLocale('en-US')"
        >
          English
        </button>
      </div>

      <div class="login-header">
        <div class="logo">
          <svg viewBox="0 0 100 100" class="logo-icon">
            <rect width="100" height="100" rx="16" fill="#10B981"/>
            <path d="M25 70 L50 30 L75 70 M35 55 L65 55" stroke="white" stroke-width="8" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="title">{{ $t('auth.loginTitle') }}</h1>
        <p class="subtitle">{{ $t('auth.loginSubtitle') }}</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div v-if="error" class="alert alert-error">
          {{ error }}
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('auth.username') }}</label>
          <input
            v-model="username"
            type="text"
            class="form-input"
            :placeholder="$t('auth.username')"
            required
            autofocus
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ $t('auth.password') }}</label>
          <input
            v-model="password"
            type="password"
            class="form-input"
            :placeholder="$t('auth.password')"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
          <span v-if="loading" class="spinner-sm"></span>
          {{ loading ? $t('auth.signingIn') : $t('auth.signIn') }}
        </button>
      </form>

      <div class="login-footer">
        <p>{{ $t('auth.defaultCredentials') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../stores/auth'
import { setLocale as setAppLocale } from '../locales'

const router = useRouter()
const { locale } = useI18n()
const { login } = useAuth()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref(null)
const currentLocale = computed(() => locale.value)

function setLocale(lang) {
  setAppLocale(lang)
}

async function handleLogin() {
  error.value = null
  loading.value = true

  try {
    await login(username.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #10B981 100%);
  padding: 1rem;
}

.login-container {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  padding: 2.5rem;
  position: relative;
}

.lang-switcher-top {
  position: absolute;
  top: 1rem;
  right: 1rem;
  display: flex;
  gap: 0.5rem;
}

.lang-btn {
  padding: 0.375rem 0.75rem;
  background: #F3F4F6;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  color: #6B7280;
  cursor: pointer;
  transition: all 0.2s;
}

.lang-btn:hover {
  background: #E5E7EB;
  color: #374151;
}

.lang-btn.active {
  background: #10B981;
  color: white;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo {
  display: flex;
  justify-content: center;
  margin-bottom: 1rem;
}

.logo-icon {
  width: 64px;
  height: 64px;
}

.title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #111827;
  margin: 0 0 0.5rem 0;
}

.subtitle {
  color: #6B7280;
  margin: 0;
  font-size: 0.875rem;
}

.login-form {
  margin-bottom: 1.5rem;
}

.alert {
  padding: 0.875rem 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.875rem;
}

.alert-error {
  background: #FEE2E2;
  color: #991B1B;
  border: 1px solid #FECACA;
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
  padding: 0.75rem 1rem;
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
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
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

.btn-block {
  width: 100%;
}

.spinner-sm {
  display: inline-block;
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

.login-footer {
  text-align: center;
  color: #9CA3AF;
  font-size: 0.75rem;
}

.login-footer p {
  margin: 0;
}
</style>
