<template>
  <div class="about-page">
    <div class="about-card">
      <div class="about-header">
        <svg viewBox="0 0 100 100" class="about-logo">
          <rect width="100" height="100" rx="16" fill="#10B981"/>
          <path d="M25 70 L50 30 L75 70 M35 55 L65 55" stroke="white" stroke-width="8" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <div class="about-title-area">
          <h2>ACME Console</h2>
          <p class="about-version">{{ $t('system.aboutVersion', { version: version }) }}</p>
        </div>
      </div>

      <p class="about-description">{{ $t('system.aboutDescription') }}</p>

      <div class="about-badges">
        <span class="badge badge-green">{{ $t('system.openSource') }}</span>
        <span class="badge badge-blue">{{ $t('system.freeLicense') }}</span>
        <span class="badge badge-purple">MIT License</span>
      </div>
    </div>

    <div class="about-card">
      <h3>{{ $t('system.features') }}</h3>
      <div class="feature-list">
        <div class="feature-item">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>{{ $t('system.featureACME') }}</span>
        </div>
        <div class="feature-item">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>{{ $t('system.featureDNS') }}</span>
        </div>
        <div class="feature-item">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>{{ $t('system.featureWildcard') }}</span>
        </div>
        <div class="feature-item">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>{{ $t('system.featureWorkspace') }}</span>
        </div>
        <div class="feature-item">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="#10B981" stroke-width="2">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>{{ $t('system.featureNotification') }}</span>
        </div>
      </div>
    </div>

    <div class="about-card">
      <h3>{{ $t('system.links') }}</h3>
      <div class="link-list">
        <a href="https://github.com/imkerbos/ACME-Console" target="_blank" rel="noopener noreferrer" class="link-item">
          <svg class="link-icon" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
          </svg>
          <div class="link-content">
            <span class="link-title">GitHub</span>
            <span class="link-url">github.com/imkerbos/ACME-Console</span>
          </div>
          <svg class="link-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 5l7 7-7 7"/>
          </svg>
        </a>
      </div>

      <div class="github-actions">
        <a href="https://github.com/imkerbos/ACME-Console" target="_blank" rel="noopener noreferrer" class="github-btn">
          <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
          </svg>
          Star
        </a>
        <a href="https://github.com/imkerbos/ACME-Console/subscription" target="_blank" rel="noopener noreferrer" class="github-btn">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
          </svg>
          Watch
        </a>
        <a href="https://github.com/imkerbos/ACME-Console/fork" target="_blank" rel="noopener noreferrer" class="github-btn">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <circle cx="12" cy="18" r="3"/>
            <circle cx="6" cy="6" r="3"/>
            <circle cx="18" cy="6" r="3"/>
            <path d="M18 9v1a2 2 0 01-2 2H8a2 2 0 01-2-2V9M12 12v3"/>
          </svg>
          Fork
        </a>
      </div>
    </div>

    <div class="about-card">
      <h3>{{ $t('system.license') }}</h3>
      <div class="license-content">
        <p>{{ $t('system.licenseText') }}</p>
      </div>
    </div>

    <div class="about-footer">
      <p>{{ $t('system.madeWith') }}</p>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const version = 'v0.1.0'
</script>

<style scoped>
.about-page {
  max-width: 640px;
  margin: 0 auto;
}

.about-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #E5E7EB;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.about-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
  margin: 0 0 1rem 0;
}

.about-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #E5E7EB;
  margin-bottom: 1.5rem;
}

.about-logo {
  width: 64px;
  height: 64px;
  flex-shrink: 0;
}

.about-title-area h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.about-version {
  color: #6B7280;
  font-size: 0.875rem;
  margin: 0;
}

.about-description {
  color: #4B5563;
  font-size: 0.9375rem;
  line-height: 1.6;
  margin: 0 0 1.25rem 0;
}

.about-badges {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.badge-green {
  background: #D1FAE5;
  color: #065F46;
}

.badge-blue {
  background: #DBEAFE;
  color: #1E40AF;
}

.badge-purple {
  background: #EDE9FE;
  color: #6D28D9;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.9375rem;
  color: #374151;
}

.feature-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.link-list {
  margin-bottom: 1.25rem;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 12px;
  text-decoration: none;
  color: inherit;
  transition: all 0.2s;
}

.link-item:hover {
  background: #F3F4F6;
  border-color: #D1D5DB;
}

.link-icon {
  width: 28px;
  height: 28px;
  color: #374151;
  flex-shrink: 0;
}

.link-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.link-title {
  font-weight: 600;
  color: #111827;
  font-size: 0.9375rem;
}

.link-url {
  color: #6B7280;
  font-size: 0.8125rem;
  margin-top: 0.125rem;
}

.link-arrow {
  width: 20px;
  height: 20px;
  color: #9CA3AF;
}

.github-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.github-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #24292e;
  color: white;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s;
}

.github-btn:hover {
  background: #1b1f23;
}

.github-btn svg {
  width: 16px;
  height: 16px;
}

.license-content {
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 12px;
  padding: 1.25rem;
}

.license-content p {
  color: #4B5563;
  font-size: 0.875rem;
  line-height: 1.7;
  margin: 0;
}

.about-footer {
  text-align: center;
  padding: 1rem 0 0.5rem;
}

.about-footer p {
  color: #9CA3AF;
  font-size: 0.8125rem;
  margin: 0;
}
</style>
