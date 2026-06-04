<template>
  <div ref="scrollRef" class="playground-chat" @scroll="onScroll">
    <!-- 空态 -->
    <div v-if="messages.length === 0" class="playground-empty">
      <div class="playground-empty-icon">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="h-10 w-10"
        >
          <path
            d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
          />
        </svg>
      </div>
      <h3 class="playground-empty-title">
        {{ t('playground.chat.emptyTitle') }}
      </h3>
      <p class="playground-empty-desc">
        {{ t('playground.chat.emptyDesc') }}
      </p>

      <!-- 系统提示词卡片 -->
      <div v-if="systemPrompt" class="playground-empty-system">
        <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('playground.chat.systemPromptLabel') }}
        </div>
        <div class="mt-1 text-sm text-gray-700 dark:text-gray-200">
          {{ systemPrompt }}
        </div>
      </div>
    </div>

    <!-- 消息列表 -->
    <div v-else class="playground-messages mx-auto w-full max-w-4xl px-4 pb-6">
      <MessageItem
        v-for="(msg, idx) in messages"
        :key="msg.key"
        :message="msg"
        :active-version-index="versionIndexMap[msg.key] ?? 0"
        :is-editing="editingKey === msg.key"
        :is-last-assistant="
          idx === messages.length - 1 && msg.from === 'assistant'
        "
        :is-generating="isGenerating"
        @regenerate="$emit('regenerate', $event)"
        @edit="$emit('edit', $event)"
        @remove="$emit('remove', $event)"
        @save="(key, content) => emit('saveEdit', key, content)"
        @save-and-submit="
          (key, content) => emit('saveEditAndSubmit', key, content)
        "
        @cancel-edit="emit('cancelEdit')"
        @switch-version="
          (key, index) => emit('switchVersion', key, index)
        "
      />
    </div>

    <!-- 滚到底部按钮 -->
    <button
      v-if="!isAtBottom && messages.length > 0"
      type="button"
      class="scroll-to-bottom"
      :title="t('playground.chat.scrollToBottom')"
      @click="() => scrollToBottom(true)"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="h-4 w-4"
      >
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import MessageItem from './MessageItem.vue'
import type { Message } from '@/types/playground'

interface Props {
  messages: Message[]
  isGenerating?: boolean
  editingKey?: string | null
  systemPrompt?: string
  /** 每条消息的当前显示版本索引（{ [messageKey]: number }） */
  versionIndexMap?: Record<string, number>
}

const props = withDefaults(defineProps<Props>(), {
  isGenerating: false,
  editingKey: null,
  systemPrompt: '',
  versionIndexMap: () => ({}),
})

const emit = defineEmits<{
  (e: 'regenerate', message: Message): void
  (e: 'edit', message: Message): void
  (e: 'remove', message: Message): void
  (e: 'saveEdit', key: string, content: string): void
  (e: 'saveEditAndSubmit', key: string, content: string): void
  (e: 'cancelEdit'): void
  (e: 'switchVersion', key: string, versionIndex: number): void
}>()

const { t } = useI18n()

const scrollRef = ref<HTMLElement | null>(null)
const isAtBottom = ref(true)

function onScroll() {
  const el = scrollRef.value
  if (!el) return
  const dist = el.scrollHeight - el.scrollTop - el.clientHeight
  isAtBottom.value = dist < 64
}

function scrollToBottom(smooth = true) {
  const el = scrollRef.value
  if (!el) return
  el.scrollTo({
    top: el.scrollHeight,
    behavior: smooth ? 'smooth' : 'auto',
  })
}

// 自动滚动跟随
watch(
  () => props.messages,
  async () => {
    if (isAtBottom.value) {
      await nextTick()
      scrollToBottom(false)
    }
  },
  { deep: true }
)
</script>

<style scoped>
.playground-chat {
  @apply relative h-full overflow-y-auto;
}

/* Empty State */
.playground-empty {
  @apply flex h-full flex-col items-center justify-center px-4 py-12 text-center;
}

.playground-empty-icon {
  @apply mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary-600;
  @apply dark:bg-primary-900/30 dark:text-primary-400;
}

.playground-empty-title {
  @apply text-xl font-semibold text-gray-900 dark:text-white;
}

.playground-empty-desc {
  @apply mt-2 max-w-md text-sm text-gray-500 dark:text-gray-400;
}

.playground-empty-system {
  @apply mt-6 w-full max-w-md rounded-xl border border-gray-200 bg-gray-50 p-4 text-left;
  @apply dark:border-dark-600 dark:bg-dark-800;
}

.scroll-to-bottom {
  @apply absolute bottom-4 right-4 inline-flex h-9 w-9 items-center justify-center rounded-full border border-gray-200 bg-white text-gray-700 shadow-md transition-all hover:scale-110;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300;
}
</style>
