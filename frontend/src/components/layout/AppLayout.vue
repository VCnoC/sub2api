<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Background Decoration: mesh 渐变 + 漂移极光 + 细网格 -->
    <div class="pointer-events-none fixed inset-0 overflow-hidden">
      <div class="absolute inset-0 bg-mesh-gradient"></div>
      <div class="aurora-blob aurora-blob-1"></div>
      <div class="aurora-blob aurora-blob-2"></div>
      <div class="aurora-blob aurora-blob-3"></div>
      <div class="absolute inset-0 bg-grid-pattern"></div>
    </div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative flex transition-all duration-300"
      :class="[
        sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64',
        fullHeight ? 'h-screen flex-col' : 'min-h-screen flex-col',
      ]"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content：吃满剩余空间，把 Footer 推到底部 -->
      <main
        :class="
          fullHeight
            ? 'flex min-h-0 flex-1 flex-col p-4 pb-0 md:p-6 md:pb-0 lg:p-8 lg:pb-0'
            : 'flex-1 p-4 md:p-6 lg:p-8'
        "
      >
        <slot />
      </main>

      <!-- Footer：fullHeight 页面隐藏（法律声明由页面自行承载） -->
      <AppFooter v-if="!fullHeight" />
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import AppFooter from './AppFooter.vue'

withDefaults(
  defineProps<{
    /** 全屏模式：内容区锁定视口高度（聊天类页面），隐藏全局 Footer */
    fullHeight?: boolean
  }>(),
  { fullHeight: false }
)

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
