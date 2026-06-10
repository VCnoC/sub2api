<template>
  <footer
    class="glass border-t border-gray-200/50 px-4 py-4 dark:border-dark-700/50 md:px-6"
  >
    <p
      class="mx-auto max-w-[1110px] text-center text-sm font-medium leading-snug text-amber-600 dark:text-amber-400"
    >
      {{ t('common.legalDisclaimer') }}
    </p>
  </footer>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
