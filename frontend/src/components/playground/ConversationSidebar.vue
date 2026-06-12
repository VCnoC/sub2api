<template>
  <aside class="conversation-sidebar">
    <!-- 头部：标题 + 新建按钮 -->
    <div class="conversation-sidebar-header">
      <span class="conversation-sidebar-title">{{ t('playground.conversations.title') }}</span>
      <button
        type="button"
        class="conversation-new-btn"
        :disabled="loading"
        @click="emit('create')"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="h-3.5 w-3.5"
        >
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        {{ t('playground.conversations.new') }}
      </button>
    </div>

    <!-- 保留期提示 -->
    <p class="conversation-retention-hint">
      {{ t('playground.conversations.retentionHint') }}
    </p>

    <!-- 列表 -->
    <div class="conversation-list">
      <!-- 加载骨架 -->
      <template v-if="loading">
        <div v-for="i in 3" :key="i" class="conversation-skeleton"></div>
      </template>

      <!-- 空态 -->
      <p v-else-if="conversations.length === 0" class="conversation-empty">
        {{ t('playground.conversations.empty') }}
      </p>

      <!-- 会话卡片（与骨架/空态互斥） -->
      <template v-else>
      <button
        v-for="item in conversations"
        :key="item.id"
        type="button"
        class="conversation-card"
        :class="{ 'conversation-card-active': item.id === activeId }"
        @click="emit('select', item.id)"
      >
        <span class="conversation-card-title">{{ item.title || t('playground.conversations.untitled') }}</span>
        <span class="conversation-card-time">{{ formatRelativeTime(item.last_activity_at) }}</span>

        <!-- 悬浮删除按钮 -->
        <span
          class="conversation-card-delete"
          :title="t('playground.conversations.delete')"
          @click.stop="emit('remove', item.id)"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="h-3.5 w-3.5"
          >
            <polyline points="3 6 5 6 21 6" />
            <path d="M19 6l-2 14a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L5 6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          </svg>
        </span>
      </button>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatRelativeTime } from '@/utils/format'
import type { ConversationSummary } from '@/types/playground'

defineProps<{
  conversations: ConversationSummary[]
  activeId: number | null
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [id: number]
  create: []
  remove: [id: number]
}>()

const { t } = useI18n()
</script>

<style scoped>
.conversation-sidebar {
  /* 高度由父级 flex 的 align-items: stretch 提供（外边距可正常生效），不用 h-full */
  @apply flex w-60 flex-shrink-0 flex-col rounded-2xl border border-gray-200 bg-white;
  @apply dark:border-dark-600 dark:bg-dark-800;
}

.conversation-sidebar-header {
  @apply flex items-center justify-between gap-2 px-3 pt-3;
}

.conversation-sidebar-title {
  @apply text-sm font-semibold text-gray-900 dark:text-white;
}

.conversation-new-btn {
  @apply inline-flex h-7 items-center gap-1 rounded-lg border border-gray-200 bg-white px-2 text-xs font-medium text-gray-700 transition-colors;
  @apply hover:border-primary-400 hover:text-primary-600;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-500 dark:hover:text-primary-300;
  @apply disabled:cursor-not-allowed disabled:opacity-50;
}

.conversation-retention-hint {
  @apply px-3 pb-1 pt-1.5 text-[11px] text-gray-400 dark:text-gray-500;
}

.conversation-list {
  @apply flex min-h-0 flex-1 flex-col gap-1.5 overflow-y-auto px-2 pb-2;
}

.conversation-skeleton {
  @apply h-14 flex-shrink-0 animate-pulse rounded-xl bg-gray-100 dark:bg-dark-700;
}

.conversation-empty {
  @apply px-2 pt-4 text-center text-xs text-gray-400 dark:text-gray-500;
}

.conversation-card {
  @apply relative flex flex-shrink-0 flex-col items-start gap-0.5 rounded-xl border border-transparent px-3 py-2.5 text-left transition-colors;
  @apply hover:bg-gray-50 dark:hover:bg-dark-700;
}

.conversation-card-active {
  @apply border-primary-400 bg-primary-50/60;
  @apply dark:border-primary-500 dark:bg-primary-900/20;
}

.conversation-card-title {
  @apply w-full truncate pr-6 text-sm font-medium text-gray-800 dark:text-gray-100;
}

.conversation-card-time {
  @apply text-xs text-gray-400 dark:text-gray-500;
}

.conversation-card-delete {
  @apply absolute right-2 top-2.5 hidden rounded-md p-1 text-gray-400 transition-colors;
  @apply hover:bg-rose-50 hover:text-rose-500 dark:hover:bg-rose-900/20 dark:hover:text-rose-300;
}

.conversation-card:hover .conversation-card-delete {
  @apply block;
}
</style>
