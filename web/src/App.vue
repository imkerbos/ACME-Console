<template>
  <router-view v-if="isLoginRoute" />
  <MainLayout v-else>
    <router-view />
  </MainLayout>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import MainLayout from './layouts/MainLayout.vue'

const route = useRoute()

const isLoginRoute = computed(() => {
  return route.path === '/login'
})
</script>
