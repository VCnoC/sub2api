<template>
  <AppLayout>
    <div
      data-testid="profile-shell"
      class="mx-auto max-w-[950px] space-y-6"
    >
      <ProfileInfoCard
        class="animate-slide-up [animation-fill-mode:backwards]"
        style="animation-delay: 0ms"
        :user="user"
        :linuxdo-enabled="linuxdoOAuthEnabled"
        :dingtalk-enabled="dingtalkOAuthEnabled"
        :oidc-enabled="oidcOAuthEnabled"
        :oidc-provider-name="oidcOAuthProviderName"
        :wechat-enabled="wechatOAuthEnabled"
        :wechat-open-enabled="wechatOAuthOpenEnabled"
        :wechat-mp-enabled="wechatOAuthMPEnabled"
      />

      <div
        v-if="contactInfo"
        class="group card card-hover relative overflow-hidden border-primary-200 bg-primary-50 p-6 dark:bg-primary-900/20 animate-slide-up [animation-fill-mode:backwards]"
        style="animation-delay: 50ms"
      >
        <div class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-sky-500/10 blur-2xl transition-colors duration-300 group-hover:bg-sky-500/20"></div>
        <div class="relative flex items-center gap-4">
          <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-sky-400 to-blue-600 text-white shadow-lg shadow-sky-500/25 transition-transform duration-300 group-hover:scale-110">
            <Icon name="chat" size="lg" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="h-4 w-1 rounded-full bg-gradient-to-b from-primary-400 to-primary-600"></span>
              <h3 class="font-semibold text-primary-800 dark:text-primary-200">
                {{ t('common.contactSupport') }}
              </h3>
            </div>
            <p class="text-sm font-medium">{{ contactInfo }}</p>
          </div>
        </div>
      </div>

      <ProfilePasswordForm
        class="animate-slide-up [animation-fill-mode:backwards]"
        style="animation-delay: 100ms"
      />

      <ProfileBalanceNotifyCard
        v-if="user && balanceLowNotifyEnabled"
        class="animate-slide-up [animation-fill-mode:backwards]"
        style="animation-delay: 150ms"
        :enabled="user.balance_notify_enabled ?? true"
        :threshold="user.balance_notify_threshold"
        :extra-emails="user.balance_notify_extra_emails ?? []"
        :system-default-threshold="systemDefaultThreshold"
        :user-email="user.email"
      />

      <ProfileTotpCard
        class="animate-slide-up [animation-fill-mode:backwards]"
        style="animation-delay: 200ms"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const contactInfo = ref('')
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const dingtalkOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || ''
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>
