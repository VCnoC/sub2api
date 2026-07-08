<template>
  <div class="min-h-screen bg-[#fafafa] dark:bg-[#020617] text-gray-900 dark:text-gray-100">
    <!-- Background Decoration: Subtle Mesh + Grid -->
    <div class="pointer-events-none fixed inset-0 overflow-hidden">
      <!-- Ambient Glows -->
      <div class="absolute -top-[20%] -left-[10%] w-[50%] h-[50%] rounded-full bg-primary-400/10 dark:bg-primary-500/5 blur-[120px] mix-blend-multiply dark:mix-blend-screen animate-pulse-slow"></div>
      <div class="absolute top-[20%] -right-[10%] w-[40%] h-[60%] rounded-full bg-cyan-400/10 dark:bg-cyan-500/5 blur-[120px] mix-blend-multiply dark:mix-blend-screen" style="animation: pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite reverse;"></div>
      
      <!-- Fine Grid -->
      <div class="absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.03)_1px,transparent_1px)] bg-[size:32px_32px] [mask-image:radial-gradient(ellipse_80%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>
    </div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative flex transition-all duration-300"
      :class="[
        sidebarCollapsed ? 'lg:ml-[88px]' : 'lg:ml-[280px]',
        fullHeight ? 'h-screen flex-col' : 'min-h-screen flex-col',
      ]"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main
        :class="
          fullHeight
            ? 'flex min-h-0 flex-1 flex-col p-4 pb-0 md:p-6 md:pb-0 lg:p-8 lg:pb-0'
            : 'flex-1 p-4 md:p-6 lg:p-8'
        "
      >
        <div class="mx-auto w-full max-w-7xl h-full flex flex-col">
          <slot />
        </div>
      </main>

      <!-- Footer -->
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
