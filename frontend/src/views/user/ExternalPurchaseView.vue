<template>
  <AppLayout>
    <div class="external-purchase-layout">
      <div class="card flex-1 min-h-0 overflow-hidden">
        <div class="external-purchase-shell">
          <a
            :href="purchaseUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm external-purchase-open"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
            {{ t('customPage.openInNewTab') }}
          </a>
          <iframe
            :src="purchaseUrl"
            class="external-purchase-frame"
            allow="payment *"
            allowfullscreen
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const defaultPurchaseUrl = 'https://pay.ldxp.cn/shop/A3B6EPW1'
const purchaseUrl = computed(() => appStore.cachedPublicSettings?.purchase_subscription_url?.trim() || defaultPurchaseUrl)
</script>

<style scoped>
.external-purchase-layout {
  @apply flex flex-col;
  height: calc(100vh - 64px - 4rem);
}

.external-purchase-shell {
  @apply relative h-full w-full overflow-hidden rounded-2xl bg-white dark:bg-dark-950;
}

.external-purchase-open {
  @apply absolute right-3 top-3 z-10;
  @apply shadow-sm backdrop-blur supports-[backdrop-filter]:bg-white/80 dark:supports-[backdrop-filter]:bg-dark-800/80;
}

.external-purchase-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
</style>
