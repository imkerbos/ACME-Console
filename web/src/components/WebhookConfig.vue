<template>
  <div class="webhook-config">
    <div class="section-header">
      <h3>{{ $t('notification.webhookConfig') }}</h3>
      <button class="btn btn-primary btn-sm" @click="showAddModal = true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 5v14M5 12h14"/>
        </svg>
        {{ $t('notification.addWebhook') }}
      </button>
    </div>

    <!-- Webhook List -->
    <div v-if="webhooks.length > 0" class="webhook-list">
      <div v-for="webhook in webhooks" :key="webhook.id" class="webhook-item">
        <div class="webhook-info">
          <div class="webhook-header">
            <span class="webhook-type">{{ getTypeLabel(webhook.type) }}</span>
            <span :class="['status-badge', webhook.enabled ? 'status-enabled' : 'status-disabled']">
              {{ webhook.enabled ? $t('notification.enabled') : $t('notification.disabled') }}
            </span>
          </div>
          <div class="webhook-url">{{ getDisplayUrl(webhook) }}</div>
          <div class="webhook-meta">
            {{ $t('notification.notifyDays') }}: {{ webhook.notify_days }} {{ $t('common.days') }}
          </div>
        </div>
        <div class="webhook-actions">
          <button class="btn btn-ghost btn-sm" @click="testWebhook(webhook)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            {{ $t('notification.testWebhook') }}
          </button>
          <button class="btn btn-ghost btn-sm" @click="editWebhook(webhook)">
            {{ $t('common.edit') }}
          </button>
          <button class="btn btn-ghost btn-sm btn-danger-text" @click="confirmDelete(webhook)">
            {{ $t('common.delete') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
        </svg>
      </div>
      <p class="empty-text">{{ $t('notification.noWebhooks') }}</p>
      <p class="empty-desc">{{ $t('notification.noWebhooksDesc') }}</p>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal || showEditModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ showEditModal ? $t('notification.editWebhook') : $t('notification.addWebhook') }}</h3>
          <button class="modal-close" @click="closeModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <form @submit.prevent="handleSubmit" class="modal-body">
          <div class="form-group">
            <label class="form-label">{{ $t('notification.type') }} <span class="required">*</span></label>
            <select v-model="form.type" class="form-select" required @change="handleTypeChange">
              <option value="generic">{{ $t('notification.typeGeneric') }}</option>
              <option value="telegram">{{ $t('notification.typeTelegram') }}</option>
              <option value="lark">{{ $t('notification.typeLark') }}</option>
            </select>
          </div>

          <!-- Telegram 配置 -->
          <template v-if="form.type === 'telegram'">
            <div class="form-group">
              <label class="form-label">{{ $t('notification.telegramBotToken') }} <span class="required">*</span></label>
              <input v-model="form.telegram_token" type="text" class="form-input" :placeholder="$t('notification.telegramBotTokenPlaceholder')" required />
              <p class="form-hint">{{ $t('notification.telegramBotTokenHint') }}</p>
            </div>

            <div class="form-group">
              <label class="form-label">{{ $t('notification.telegramChatId') }} <span class="required">*</span></label>
              <input v-model="form.telegram_chat_id" type="text" class="form-input" :placeholder="$t('notification.telegramChatIdPlaceholder')" required />
              <p class="form-hint">{{ $t('notification.telegramChatIdHint') }}</p>
            </div>
          </template>

          <!-- Lark 配置 -->
          <template v-if="form.type === 'lark'">
            <div class="form-group">
              <label class="form-label">{{ $t('notification.webhookUrl') }} <span class="required">*</span></label>
              <input v-model="form.webhook_url" type="url" class="form-input" :placeholder="$t('notification.larkWebhookPlaceholder')" required />
              <p class="form-hint">{{ $t('notification.larkWebhookHint') }}</p>
            </div>
          </template>

          <!-- Generic Webhook 配置 -->
          <template v-if="form.type === 'generic'">
            <div class="form-group">
              <label class="form-label">{{ $t('notification.webhookUrl') }} <span class="required">*</span></label>
              <input v-model="form.webhook_url" type="url" class="form-input" :placeholder="$t('notification.genericWebhookPlaceholder')" required />
            </div>

            <div class="form-group">
              <label class="form-label">{{ $t('notification.extraConfig') }}</label>
              <textarea v-model="form.webhook_config" class="form-textarea" rows="4" :placeholder="$t('notification.extraConfigPlaceholder')"></textarea>
              <p class="form-hint">{{ $t('notification.extraConfigHint') }}</p>
            </div>
          </template>

          <div class="form-group">
            <label class="form-label">{{ $t('notification.notifyDays') }} <span class="required">*</span></label>
            <input v-model.number="form.notify_days" type="number" class="form-input" min="1" max="90" required />
            <p class="form-hint">{{ $t('notification.notifyDaysHint') }}</p>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="form.enabled" type="checkbox" class="form-checkbox" />
              <span>{{ $t('notification.enabled') }}</span>
            </label>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="closeModal">{{ $t('common.cancel') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? $t('common.loading') : $t('common.save') }}
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
          <p>{{ $t('notification.deleteConfirm') }}</p>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">{{ $t('common.cancel') }}</button>
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
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { notificationApi } from '../api'

const props = defineProps({
  workspaceId: {
    type: Number,
    default: null
  },
  certificateId: {
    type: Number,
    default: null
  }
})

const { t } = useI18n()

const webhooks = ref([])
const showAddModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const testing = ref(false)
const editingWebhook = ref(null)
const deleteWebhook = ref(null)

const form = ref({
  type: 'generic',
  webhook_url: '',
  webhook_config: '',
  telegram_token: '',
  telegram_chat_id: '',
  notify_days: 7,
  enabled: true
})

function getTypeLabel(type) {
  const labels = {
    generic: t('notification.typeGeneric'),
    telegram: t('notification.typeTelegram'),
    lark: t('notification.typeLark')
  }
  return labels[type] || type
}

function handleTypeChange() {
  // 切换类型时清空相关字段
  form.value.webhook_url = ''
  form.value.webhook_config = ''
  form.value.telegram_token = ''
  form.value.telegram_chat_id = ''
}

function getDisplayUrl(webhook) {
  if (webhook.type === 'telegram') {
    return `Telegram Bot (Chat ID: ${webhook.telegram_chat_id || '-'})`
  }
  return webhook.webhook_url
}

async function loadWebhooks() {
  try {
    const params = {}
    if (props.workspaceId) params.workspace_id = props.workspaceId
    if (props.certificateId) params.certificate_id = props.certificateId

    const response = await notificationApi.list(params)
    webhooks.value = response.data || []
  } catch (e) {
    console.error('Failed to load webhooks:', e)
  }
}

function editWebhook(webhook) {
  editingWebhook.value = webhook

  // 解析 webhook_config 以获取 telegram 配置
  let telegramToken = ''
  let telegramChatId = ''

  if (webhook.type === 'telegram' && webhook.webhook_config) {
    try {
      const config = JSON.parse(webhook.webhook_config)
      telegramToken = config.token || ''
      telegramChatId = config.chat_id || ''
    } catch (e) {
      console.error('Failed to parse webhook config:', e)
    }
  }

  form.value = {
    type: webhook.type,
    webhook_url: webhook.webhook_url || '',
    webhook_config: webhook.type === 'generic' ? (webhook.webhook_config || '') : '',
    telegram_token: telegramToken,
    telegram_chat_id: telegramChatId,
    notify_days: webhook.notify_days,
    enabled: webhook.enabled
  }
  showEditModal.value = true
}

function confirmDelete(webhook) {
  deleteWebhook.value = webhook
  showDeleteModal.value = true
}

async function testWebhook(webhook) {
  if (testing.value) return
  testing.value = true

  try {
    await notificationApi.test(webhook.id)
    alert(t('notification.testSuccess'))
  } catch (e) {
    const errorMsg = e.response?.data?.message || e.message || t('notification.testFailed')
    alert(t('notification.testFailed') + ': ' + errorMsg)
    console.error('Test webhook failed:', e)
  } finally {
    testing.value = false
  }
}

async function handleSubmit() {
  submitting.value = true

  try {
    // 根据类型构造数据
    let webhookUrl = form.value.webhook_url
    let webhookConfig = form.value.webhook_config

    if (form.value.type === 'telegram') {
      // Telegram: 使用 API URL，配置存储 token 和 chat_id
      webhookUrl = `https://api.telegram.org/bot${form.value.telegram_token}/sendMessage`
      webhookConfig = JSON.stringify({
        token: form.value.telegram_token,
        chat_id: form.value.telegram_chat_id
      })
    } else if (form.value.type === 'lark') {
      // Lark: 只需要 webhook_url
      webhookConfig = ''
    }

    const data = {
      type: form.value.type,
      webhook_url: webhookUrl,
      webhook_config: webhookConfig,
      notify_days: form.value.notify_days,
      enabled: form.value.enabled,
      workspace_id: props.workspaceId || null,
      certificate_id: props.certificateId || null
    }

    if (showEditModal.value) {
      await notificationApi.update(editingWebhook.value.id, data)
    } else {
      await notificationApi.create(data)
    }

    closeModal()
    loadWebhooks()
  } catch (e) {
    console.error('Failed to save webhook:', e)
  } finally {
    submitting.value = false
  }
}

async function handleDelete() {
  deleting.value = true

  try {
    await notificationApi.delete(deleteWebhook.value.id)
    showDeleteModal.value = false
    loadWebhooks()
  } catch (e) {
    console.error('Failed to delete webhook:', e)
  } finally {
    deleting.value = false
  }
}

function closeModal() {
  showAddModal.value = false
  showEditModal.value = false
  editingWebhook.value = null
  form.value = {
    type: 'generic',
    webhook_url: '',
    webhook_config: '',
    telegram_token: '',
    telegram_chat_id: '',
    notify_days: 7,
    enabled: true
  }
}

onMounted(() => {
  loadWebhooks()
})
</script>

<style scoped>
.webhook-config {
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

.section-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.webhook-list {
  padding: 1rem;
}

.webhook-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #F9FAFB;
  border-radius: 8px;
  margin-bottom: 0.75rem;
}

.webhook-item:last-child {
  margin-bottom: 0;
}

.webhook-info {
  flex: 1;
}

.webhook-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.webhook-type {
  font-weight: 600;
  color: #111827;
}

.status-badge {
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-enabled {
  background: #D1FAE5;
  color: #065F46;
}

.status-disabled {
  background: #F3F4F6;
  color: #6B7280;
}

.webhook-url {
  font-size: 0.875rem;
  color: #6B7280;
  font-family: monospace;
  margin-bottom: 0.25rem;
  word-break: break-all;
}

.webhook-meta {
  font-size: 0.8125rem;
  color: #9CA3AF;
}

.webhook-actions {
  display: flex;
  gap: 0.5rem;
}

.empty-state {
  padding: 3rem 2rem;
  text-align: center;
}

.empty-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 1rem;
  background: #F3F4F6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-icon svg {
  width: 24px;
  height: 24px;
  color: #9CA3AF;
}

.empty-text {
  font-weight: 500;
  color: #111827;
  margin-bottom: 0.25rem;
}

.empty-desc {
  font-size: 0.875rem;
  color: #6B7280;
  margin: 0;
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
  font-family: monospace;
}

.form-hint {
  font-size: 0.8125rem;
  color: #6B7280;
  margin-top: 0.25rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.form-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
</style>
