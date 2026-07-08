<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-[#fafafa] dark:bg-[#020617]"
  >
    <!-- Background Decorations: Subtle Mesh + Grid -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Ambient Glows -->
      <div class="absolute -top-[20%] -left-[10%] w-[50%] h-[50%] rounded-full bg-primary-400/20 dark:bg-primary-500/10 blur-[120px] mix-blend-multiply dark:mix-blend-screen animate-pulse-slow"></div>
      <div class="absolute top-[20%] -right-[10%] w-[40%] h-[60%] rounded-full bg-cyan-400/20 dark:bg-cyan-500/10 blur-[120px] mix-blend-multiply dark:mix-blend-screen" style="animation: pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite reverse;"></div>
      
      <!-- Fine Grid -->
      <div class="absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.03)_1px,transparent_1px)] bg-[size:32px_32px] [mask-image:radial-gradient(ellipse_80%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>
    </div>

    <!-- Header -->
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-7xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-2xl bg-white/50 dark:bg-white/5 shadow-sm ring-1 ring-black/5 dark:ring-white/10 backdrop-blur-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-6 w-6 object-contain" />
          </div>
          <span class="text-xl font-bold tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-2 rounded-full bg-white/50 dark:bg-white/5 px-2 py-1.5 shadow-sm ring-1 ring-black/5 dark:ring-white/10 backdrop-blur-md">
          <LocaleSwitcher />
          
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="rounded-full p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-white/10 dark:hover:text-white" :title="t('home.viewDocs')">
            <Icon name="book" size="sm" />
          </a>

          <button @click="toggleTheme" class="rounded-full p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-white/10 dark:hover:text-white" :title="isDark ? t('home.switchToLight') : t('home.switchToDark')">
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>

          <div class="mx-2 h-4 w-px bg-gray-300 dark:bg-white/10"></div>

          <router-link v-if="isAuthenticated" :to="dashboardPath" class="group flex items-center gap-2 rounded-full bg-gray-900 py-1.5 pl-1.5 pr-4 transition-all hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-100">
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-bold text-white shadow-inner">{{ userInitial }}</span>
            <span class="text-sm font-medium text-white dark:text-gray-900">{{ t('home.dashboard') }}</span>
          </router-link>
          
          <router-link v-else to="/login" class="rounded-full bg-gray-900 px-5 py-1.5 text-sm font-medium text-white transition-all hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100">
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 pt-20 pb-16">
      <div class="mx-auto max-w-7xl">
        
        <!-- Hero Section -->
        <div class="mb-24 text-center">
          <div class="inline-flex animate-slide-down items-center gap-2 rounded-full border border-primary-500/20 bg-primary-500/10 px-4 py-1.5 text-sm font-medium text-primary-600 dark:text-primary-400 backdrop-blur-md mb-8">
            <span class="relative flex h-2 w-2">
              <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary-400 opacity-75"></span>
              <span class="relative inline-flex h-2 w-2 rounded-full bg-primary-500"></span>
            </span>
            {{ siteSubtitle }}
          </div>
          
          <h1 class="mx-auto mb-6 max-w-4xl text-5xl font-extrabold tracking-tight text-gray-900 dark:text-white sm:text-6xl lg:text-7xl animate-slide-up [animation-delay:100ms] [animation-fill-mode:backwards]">
            The Ultimate API <br/>
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-primary-500 to-cyan-500">Gateway Platform</span>
          </h1>
          
          <p class="mx-auto mb-10 max-w-2xl text-lg text-gray-600 dark:text-gray-400 animate-slide-up [animation-delay:200ms] [animation-fill-mode:backwards]">
            Seamlessly manage, route, and monetize your AI API calls with enterprise-grade reliability, real-time analytics, and beautiful glassmorphism design.
          </p>

          <div class="flex items-center justify-center gap-4 animate-slide-up [animation-delay:300ms] [animation-fill-mode:backwards]">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="group relative inline-flex items-center justify-center gap-2 overflow-hidden rounded-full bg-gray-900 px-8 py-3.5 text-sm font-medium text-white transition-all hover:scale-105 hover:shadow-[0_0_40px_rgba(0,0,0,0.2)] dark:bg-white dark:text-gray-900 dark:hover:shadow-[0_0_40px_rgba(255,255,255,0.2)]">
              <span class="relative z-10">{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
              <Icon name="arrowRight" size="sm" class="relative z-10 transition-transform group-hover:translate-x-1" />
            </router-link>
            
            <a v-if="docUrl" :href="docUrl" target="_blank" class="inline-flex items-center justify-center gap-2 rounded-full bg-white/50 dark:bg-white/5 px-8 py-3.5 text-sm font-medium text-gray-900 dark:text-white shadow-sm ring-1 ring-black/5 dark:ring-white/10 backdrop-blur-md transition-all hover:bg-white/80 dark:hover:bg-white/10">
              <Icon name="book" size="sm" />
              {{ t('home.docs') }}
            </a>
          </div>
        </div>

        <!-- Bento Box Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-6 auto-rows-[240px] animate-slide-up [animation-delay:400ms] [animation-fill-mode:backwards]">
          
          <!-- Large Feature 1 -->
          <SpotlightCard class="md:col-span-2 lg:col-span-2 row-span-2 p-8 flex flex-col justify-between">
            <div>
              <div class="mb-6 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-cyan-500 text-white shadow-lg shadow-primary-500/30">
                <Icon name="server" size="lg" />
              </div>
              <h3 class="text-2xl font-bold text-gray-900 dark:text-white mb-3">{{ t('home.features.unifiedGateway') }}</h3>
              <p class="text-gray-600 dark:text-gray-400 leading-relaxed max-w-md">
                {{ t('home.features.unifiedGatewayDesc') }}
              </p>
            </div>
            
            <!-- Visual Element -->
            <div class="relative mt-8 h-40 w-full overflow-hidden rounded-xl border border-gray-200/50 dark:border-white/10 bg-gray-50/50 dark:bg-black/20">
              <div class="absolute inset-0 flex items-center justify-center">
                <div class="flex items-center gap-4">
                  <div class="h-12 w-12 rounded-xl bg-white dark:bg-dark-800 shadow-sm flex items-center justify-center text-xl font-bold">👤</div>
                  <div class="h-1 w-16 bg-gradient-to-r from-gray-300 to-primary-500 dark:from-gray-700"></div>
                  <div class="h-16 w-16 rounded-2xl bg-gradient-to-br from-primary-400 to-cyan-500 shadow-lg shadow-primary-500/20 flex items-center justify-center text-white"><Icon name="server" size="lg"/></div>
                  <div class="h-1 w-16 bg-gradient-to-r from-primary-500 to-gray-300 dark:to-gray-700"></div>
                  <div class="flex flex-col gap-2">
                    <div class="h-8 w-8 rounded-lg bg-white dark:bg-dark-800 shadow-sm flex items-center justify-center text-xs font-bold">G</div>
                    <div class="h-8 w-8 rounded-lg bg-white dark:bg-dark-800 shadow-sm flex items-center justify-center text-xs font-bold">C</div>
                  </div>
                </div>
              </div>
            </div>
          </SpotlightCard>

          <!-- Terminal / Code Snippet -->
          <SpotlightCard class="md:col-span-1 lg:col-span-2 row-span-1 p-0 overflow-hidden bg-gray-900 dark:bg-[#0f172a] border-gray-800">
            <div class="flex h-10 items-center gap-2 border-b border-white/10 bg-white/5 px-4">
              <div class="h-3 w-3 rounded-full bg-red-500/80"></div>
              <div class="h-3 w-3 rounded-full bg-yellow-500/80"></div>
              <div class="h-3 w-3 rounded-full bg-green-500/80"></div>
              <span class="ml-2 text-xs font-mono text-gray-400">request.sh</span>
            </div>
            <div class="p-5 font-mono text-sm leading-relaxed text-gray-300">
              <div class="flex"><span class="text-primary-400 mr-2">$</span> <span class="text-blue-400">curl</span>&nbsp;-X POST /v1/chat/completions \</div>
              <div class="ml-4">-H <span class="text-green-400">"Authorization: Bearer sk-..."</span> \</div>
              <div class="ml-4">-d <span class="text-yellow-400">'{"model": "gpt-4", "messages": [...]}'</span></div>
              <div class="mt-4 text-gray-500"># Response: 200 OK</div>
            </div>
          </SpotlightCard>

          <!-- Feature 2 -->
          <SpotlightCard class="md:col-span-1 lg:col-span-1 row-span-1 p-6 flex flex-col justify-between">
            <div class="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
              <Icon name="key" size="md" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-2">{{ t('home.features.keyManagement') }}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('home.features.keyManagementDesc') }}</p>
            </div>
          </SpotlightCard>

          <!-- Feature 3 -->
          <SpotlightCard class="md:col-span-1 lg:col-span-1 row-span-1 p-6 flex flex-col justify-between">
            <div class="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <Icon name="chart" size="md" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-2">{{ t('home.features.realtimeStats') }}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('home.features.realtimeStatsDesc') }}</p>
            </div>
          </SpotlightCard>

          <!-- Feature 4 -->
          <SpotlightCard class="md:col-span-2 lg:col-span-2 row-span-1 p-6 flex items-center gap-6">
            <div class="flex-1">
              <div class="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-orange-500/10 text-orange-600 dark:text-orange-400">
                <Icon name="shield" size="md" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-2">{{ t('home.features.highAvailability') }}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('home.features.highAvailabilityDesc') }}</p>
            </div>
            <div class="hidden sm:flex flex-1 justify-end">
              <div class="relative h-24 w-24">
                <div class="absolute inset-0 rounded-full border-4 border-primary-500/20 border-t-primary-500 animate-spin"></div>
                <div class="absolute inset-2 rounded-full border-4 border-cyan-500/20 border-b-cyan-500 animate-spin" style="animation-direction: reverse; animation-duration: 1.5s;"></div>
                <div class="absolute inset-0 flex items-center justify-center text-2xl font-bold text-gray-900 dark:text-white">99%</div>
              </div>
            </div>
          </SpotlightCard>

        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="relative z-10 mt-auto border-t border-gray-200/50 dark:border-white/10 bg-white/30 dark:bg-black/30 backdrop-blur-md px-6 py-8">
      <div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 sm:flex-row">
        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-6">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="text-sm font-medium text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white">
            {{ t('home.docs') }}
          </a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="text-sm font-medium text-gray-500 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white">
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import SpotlightCard from '@/components/common/SpotlightCard.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
