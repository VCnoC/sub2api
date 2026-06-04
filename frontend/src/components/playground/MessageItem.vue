<template>
  <div
    class="group message-row"
    :class="{
      'message-row-user': message.from === 'user',
      'message-row-assistant': message.from === 'assistant',
    }"
  >
    <!-- 头像（仅 AI 显示） -->
    <div v-if="message.from === 'assistant'" class="message-avatar">
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="h-4 w-4 text-primary-600 dark:text-primary-400"
      >
        <path d="M12 8V4H8" />
        <rect width="16" height="12" x="4" y="8" rx="2" />
        <path d="M2 14h2" />
        <path d="M20 14h2" />
        <path d="M15 13v2" />
        <path d="M9 13v2" />
      </svg>
    </div>

    <div class="message-body">
      <!-- 编辑模式 -->
      <div v-if="isEditing" class="message-edit-box">
        <textarea
          v-model="editText"
          class="message-edit-textarea"
          rows="6"
        />
        <div class="message-edit-actions">
          <button
            v-if="message.from === 'user'"
            type="button"
            class="msg-btn msg-btn-primary"
            :disabled="!editChanged || !editText.trim()"
            @click="onSaveAndSubmit"
          >
            {{ t('playground.message.saveAndSubmit') }}
          </button>
          <button
            type="button"
            class="msg-btn msg-btn-primary"
            :disabled="!editChanged || !editText.trim()"
            @click="onSave"
          >
            {{ t('playground.message.save') }}
          </button>
          <button type="button" class="msg-btn msg-btn-outline" @click="onCancel">
            {{ t('playground.message.cancel') }}
          </button>
        </div>
      </div>

      <!-- 正常模式 -->
      <template v-else>
        <!-- 推理（思考链）折叠区 -->
        <div
          v-if="hasReasoning"
          class="message-reasoning"
          :class="{ 'message-reasoning-streaming': message.isReasoningStreaming }"
        >
          <button
            type="button"
            class="message-reasoning-trigger"
            @click="reasoningOpen = !reasoningOpen"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-3.5 w-3.5"
            >
              <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <span>{{
              message.isReasoningStreaming
                ? t('playground.message.thinking')
                : t('playground.message.thoughtFor')
            }}</span>
            <svg
              :class="['h-3.5 w-3.5 transition-transform', reasoningOpen ? 'rotate-180' : '']"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </button>
          <div v-show="reasoningOpen" class="message-reasoning-content">
            {{ message.reasoning?.content }}
          </div>
        </div>

        <!-- 加载中（无内容时的脉冲） -->
        <div v-if="showLoader" class="message-loader">
          <div class="loader-dot"></div>
          <div class="loader-dot"></div>
          <div class="loader-dot"></div>
        </div>

        <!-- 错误消息 -->
        <div v-else-if="message.status === 'error'" class="message-error">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="h-4 w-4 mt-0.5 flex-shrink-0"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          <span>{{ currentContent || t('playground.error.unknown') }}</span>
        </div>

        <!-- 正常内容（Markdown 渲染） -->
        <div
          v-else-if="currentContent"
          class="message-content"
          :class="contentClass"
        >
          <div
            v-if="message.from === 'assistant'"
            class="markdown-body"
            @click="onMarkdownClick"
            v-html="renderedHtml"
          />
          <div v-else class="user-content">{{ currentContent }}</div>
        </div>

        <!-- 多版本切换 -->
        <div v-if="versionCount > 1" class="message-version-switcher">
          <button
            type="button"
            class="msg-btn-icon"
            :disabled="activeVersionIndex === 0"
            :title="t('playground.message.prevVersion')"
            @click="$emit('switchVersion', message.key, activeVersionIndex - 1)"
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
              <polyline points="15 18 9 12 15 6" />
            </svg>
          </button>
          <span class="message-version-page">
            {{ activeVersionIndex + 1 }} / {{ versionCount }}
          </span>
          <button
            type="button"
            class="msg-btn-icon"
            :disabled="activeVersionIndex === versionCount - 1"
            :title="t('playground.message.nextVersion')"
            @click="$emit('switchVersion', message.key, activeVersionIndex + 1)"
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
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </button>
        </div>

        <!-- 消息操作（悬停或最后 AI 消息常驻） -->
        <div
          class="message-actions"
          :class="{
            'message-actions-visible': isLastAssistant,
          }"
        >
          <button
            type="button"
            class="msg-btn-icon"
            :title="t('playground.message.copy')"
            @click="onCopy"
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
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
            </svg>
          </button>

          <button
            v-if="message.from === 'assistant' && !isGenerating"
            type="button"
            class="msg-btn-icon"
            :title="t('playground.message.regenerate')"
            @click="$emit('regenerate', message)"
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
              <polyline points="23 4 23 10 17 10" />
              <polyline points="1 20 1 14 7 14" />
              <path
                d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"
              />
            </svg>
          </button>

          <button
            type="button"
            class="msg-btn-icon"
            :title="t('playground.message.edit')"
            @click="$emit('edit', message)"
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
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
          </button>

          <button
            type="button"
            class="msg-btn-icon msg-btn-icon-danger"
            :title="t('playground.message.delete')"
            @click="$emit('remove', message)"
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
              <path
                d="M19 6l-2 14a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L5 6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              />
            </svg>
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Marked } from 'marked'
import type { Tokens } from 'marked'
import DOMPurify from 'dompurify'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import type { Message } from '@/types/playground'

interface Props {
  message: Message
  /** 当前显示的版本索引（默认 0） */
  activeVersionIndex?: number
  /** 是否在编辑模式 */
  isEditing?: boolean
  /** 是否是最后一条 AI 消息（操作按钮常驻） */
  isLastAssistant?: boolean
  /** 是否在生成中（禁用 regenerate） */
  isGenerating?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  activeVersionIndex: 0,
  isEditing: false,
  isLastAssistant: false,
  isGenerating: false,
})

const emit = defineEmits<{
  (e: 'regenerate', message: Message): void
  (e: 'edit', message: Message): void
  (e: 'remove', message: Message): void
  (e: 'save', key: string, content: string): void
  (e: 'saveAndSubmit', key: string, content: string): void
  (e: 'cancelEdit'): void
  (e: 'switchVersion', key: string, index: number): void
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const appStore = useAppStore()

// 独立的 Marked 实例 — 不污染全局 marked 配置（公告 / 自定义页面共用全局实例）
// 自定义 code renderer：用 .pg-code-block 包装，加上语言标签与复制按钮
const markedInstance = new Marked({ breaks: true, gfm: true })
markedInstance.use({
  renderer: {
    code({ text, lang }: Tokens.Code): string {
      const language = (lang || '').trim().toLowerCase() || 'plaintext'
      // 转义 HTML 特殊字符（marked 默认不对 code 内容做 escape）
      const escapedText = String(text)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
      // 复制按钮的 SVG（保留 24x24 视框，CSS 控制大小）
      const copyIcon =
        '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>'
      const checkIcon =
        '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="20 6 9 17 4 12"/></svg>'
      return (
        '<div class="pg-code-block">' +
        '<div class="pg-code-block-header">' +
        `<span class="pg-code-block-lang">${language}</span>` +
        '<button type="button" class="pg-code-block-copy" aria-label="copy">' +
        `<span class="pg-code-copy-icon">${copyIcon}</span>` +
        `<span class="pg-code-check-icon">${checkIcon}</span>` +
        '</button>' +
        '</div>' +
        `<pre class="pg-code-block-content"><code class="language-${language}">${escapedText}</code></pre>` +
        '</div>'
      )
    },
  },
})

// ==================== 计算属性 ====================

const versionCount = computed(() => props.message.versions?.length ?? 1)

const currentContent = computed(
  () =>
    props.message.versions?.[Math.min(props.activeVersionIndex, versionCount.value - 1)]
      ?.content || ''
)

const hasReasoning = computed(
  () =>
    props.message.from === 'assistant' &&
    !!props.message.reasoning?.content
)

const reasoningOpen = ref(true)

const showLoader = computed(
  () =>
    props.message.from === 'assistant' &&
    !props.message.isReasoningStreaming &&
    (props.message.status === 'loading' ||
      (props.message.status === 'streaming' && !currentContent.value))
)

const contentClass = computed(() => ({
  'content-user': props.message.from === 'user',
  'content-assistant': props.message.from === 'assistant',
}))

const renderedHtml = computed(() => {
  if (!currentContent.value) return ''
  const raw = markedInstance.parse(currentContent.value) as string
  return DOMPurify.sanitize(raw)
})

// 事件委托：捕获代码块内的复制按钮点击 → 取出 code 文本 → 写剪贴板 → 触发反馈动画
async function onMarkdownClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null
  const btn = target?.closest('.pg-code-block-copy') as HTMLElement | null
  if (!btn) return
  e.preventDefault()
  const codeEl = btn
    .closest('.pg-code-block')
    ?.querySelector<HTMLElement>('pre code')
  const text = codeEl?.textContent ?? ''
  if (!text) {
    appStore.showInfo(t('playground.message.noContent'))
    return
  }
  const ok = await copyToClipboard(text, t('playground.message.codeCopied'))
  if (ok) {
    btn.classList.add('pg-copied')
    window.setTimeout(() => btn.classList.remove('pg-copied'), 1500)
  }
}

// ==================== 编辑 ====================

const editText = ref('')
const editOriginal = ref('')
const editChanged = computed(() => editText.value !== editOriginal.value)

watch(
  () => [props.isEditing, props.message.key],
  () => {
    if (props.isEditing) {
      const v = currentContent.value
      editText.value = v
      editOriginal.value = v
    }
  },
  { immediate: true }
)

function onCancel() {
  emit('cancelEdit')
}
function onSave() {
  emit('save', props.message.key, editText.value)
}
function onSaveAndSubmit() {
  emit('saveAndSubmit', props.message.key, editText.value)
}

// ==================== 复制 ====================

async function onCopy() {
  if (!currentContent.value) {
    appStore.showInfo(t('playground.message.noContent'))
    return
  }
  await copyToClipboard(currentContent.value, t('playground.message.copied'))
}
</script>

<style scoped>
.message-row {
  @apply relative flex gap-3 py-4;
}
.message-row-user {
  @apply justify-end;
}
.message-row-assistant {
  @apply justify-start;
}

.message-avatar {
  @apply flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 mt-1;
  @apply dark:bg-primary-900/30;
}

.message-body {
  @apply min-w-0 max-w-full;
}

.message-row-user .message-body {
  @apply max-w-[85%] sm:max-w-[62ch] md:max-w-[68ch] lg:max-w-[72ch];
}
.message-row-assistant .message-body {
  @apply w-full max-w-none flex-1;
}

/* Content */
.message-content {
  @apply break-words text-base leading-relaxed sm:leading-7;
}

.content-user {
  @apply rounded-3xl bg-gray-100 px-4 py-2.5 text-gray-900;
  @apply dark:bg-dark-700 dark:text-white;
}

.content-assistant {
  @apply bg-transparent p-0 text-gray-900;
  @apply dark:text-gray-100;
}

.user-content {
  @apply whitespace-pre-wrap break-words;
}

/* Reasoning */
.message-reasoning {
  @apply mb-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2;
  @apply dark:border-dark-600 dark:bg-dark-700/50;
}

.message-reasoning-streaming {
  @apply border-primary-300 dark:border-primary-700;
}

.message-reasoning-trigger {
  @apply flex w-full items-center gap-1.5 text-xs font-medium text-gray-600;
  @apply dark:text-gray-300;
}

.message-reasoning-content {
  @apply mt-2 max-h-72 overflow-y-auto whitespace-pre-wrap break-words border-t border-gray-200 pt-2 text-xs leading-relaxed text-gray-500;
  @apply dark:border-dark-600 dark:text-gray-400;
}

/* Loader */
.message-loader {
  @apply flex items-center gap-1 py-2;
}
.loader-dot {
  @apply h-2 w-2 animate-pulse rounded-full bg-gray-400;
  @apply dark:bg-gray-500;
}
.loader-dot:nth-child(2) { animation-delay: 0.2s; }
.loader-dot:nth-child(3) { animation-delay: 0.4s; }

/* Error */
.message-error {
  @apply flex items-start gap-2 rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700;
  @apply dark:border-rose-900/50 dark:bg-rose-900/20 dark:text-rose-300;
}

/* Version switcher */
.message-version-switcher {
  @apply mt-1.5 flex items-center gap-2 text-xs text-gray-500;
  @apply dark:text-gray-400;
}
.message-version-page {
  @apply tabular-nums;
}

/* Actions */
.message-actions {
  @apply mt-1.5 flex items-center gap-0.5 opacity-0 transition-opacity duration-150;
}
.message-row:hover .message-actions {
  @apply opacity-100;
}
.message-actions-visible {
  @apply opacity-100;
}

.msg-btn-icon {
  @apply inline-flex h-7 w-7 items-center justify-center rounded-md text-gray-500 transition-colors;
  @apply hover:bg-gray-100 hover:text-gray-900;
  @apply dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white;
  @apply disabled:cursor-not-allowed disabled:opacity-40;
}
.msg-btn-icon-danger {
  @apply hover:text-rose-600 dark:hover:text-rose-400;
}

/* Edit */
.message-edit-box {
  @apply space-y-2;
}
.message-edit-textarea {
  @apply w-full rounded-lg border border-gray-200 bg-white p-3 font-mono text-sm leading-6 focus:border-primary-400 focus:outline-none focus:ring-1 focus:ring-primary-400;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100;
}
.message-edit-actions {
  @apply flex flex-wrap gap-2;
}

.msg-btn {
  @apply inline-flex h-8 items-center justify-center rounded-lg px-3 text-xs font-medium transition-colors;
}
.msg-btn-primary {
  @apply bg-primary-600 text-white hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-50;
}
.msg-btn-outline {
  @apply border border-gray-200 text-gray-700 hover:bg-gray-50;
  @apply dark:border-dark-600 dark:text-gray-200 dark:hover:bg-dark-700;
}

/* === Markdown 代码块（v-html 渲染产物，需用 :deep() 突破 scoped） === */
.markdown-body :deep(.pg-code-block) {
  @apply my-3 overflow-hidden rounded-xl border border-gray-200 bg-gray-50;
  @apply dark:border-dark-600 dark:bg-dark-900/60;
}
.markdown-body :deep(.pg-code-block-header) {
  @apply flex items-center justify-between border-b border-gray-200 bg-gray-100 px-3 py-1.5;
  @apply dark:border-dark-600 dark:bg-dark-800;
}
.markdown-body :deep(.pg-code-block-lang) {
  @apply font-mono text-xs uppercase tracking-wide text-gray-500;
  @apply dark:text-gray-400;
}
.markdown-body :deep(.pg-code-block-copy) {
  @apply relative inline-flex h-7 w-7 cursor-pointer items-center justify-center rounded-md text-gray-500 transition-colors;
  @apply hover:bg-gray-200 hover:text-gray-900;
  @apply dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white;
}
.markdown-body :deep(.pg-code-block-copy svg) {
  width: 14px;
  height: 14px;
}
.markdown-body :deep(.pg-code-copy-icon),
.markdown-body :deep(.pg-code-check-icon) {
  @apply inline-flex transition-opacity duration-150;
}
.markdown-body :deep(.pg-code-check-icon) {
  @apply absolute opacity-0;
}
.markdown-body :deep(.pg-code-block-copy.pg-copied) {
  @apply text-green-600 dark:text-green-400;
}
.markdown-body :deep(.pg-code-block-copy.pg-copied .pg-code-copy-icon) {
  @apply opacity-0;
}
.markdown-body :deep(.pg-code-block-copy.pg-copied .pg-code-check-icon) {
  @apply opacity-100;
}
/* 强力重置 <pre>：去掉所有可能来自全局/UA 的边框、圆角、阴影、背景，
   让外层 .pg-code-block 成为唯一的视觉容器 */
.markdown-body :deep(.pg-code-block pre),
.markdown-body :deep(pre.pg-code-block-content) {
  margin: 0 !important;
  padding: 14px 18px !important;
  background: transparent !important;
  border: 0 !important;
  border-radius: 0 !important;
  box-shadow: none !important;
  outline: none !important;
  overflow-x: auto;
  font-size: 0.9rem;
  line-height: 1.65;
  color: inherit;
  display: block;
  width: 100%;
}
.markdown-body :deep(.pg-code-block pre code),
.markdown-body :deep(.pg-code-block-content code) {
  background: transparent !important;
  border: 0 !important;
  border-radius: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
  color: inherit;
  display: block;
  white-space: pre;
  font-family: 'JetBrains Mono', 'Fira Code', Menlo, Monaco, 'Courier New', monospace;
  font-size: inherit;
  line-height: inherit;
}

/* 行内 code（非围栏块） */
.markdown-body :deep(:not(pre) > code) {
  @apply rounded-md bg-gray-100 px-1.5 py-0.5 font-mono text-[0.875em] text-rose-600;
  @apply dark:bg-dark-700 dark:text-rose-300;
}

/* 段落基础排版 */
.markdown-body :deep(p) {
  @apply my-2 leading-relaxed first:mt-0 last:mb-0;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  @apply my-2 ml-6 space-y-1;
}
.markdown-body :deep(ul) { list-style: disc; }
.markdown-body :deep(ol) { list-style: decimal; }
.markdown-body :deep(blockquote) {
  @apply my-3 border-l-4 border-gray-300 pl-3 text-gray-600;
  @apply dark:border-dark-600 dark:text-gray-300;
}
.markdown-body :deep(h1) { @apply mt-4 mb-2 text-2xl font-bold; }
.markdown-body :deep(h2) { @apply mt-4 mb-2 text-xl font-bold; }
.markdown-body :deep(h3) { @apply mt-3 mb-1.5 text-lg font-semibold; }
.markdown-body :deep(h4) { @apply mt-3 mb-1 text-base font-semibold; }
.markdown-body :deep(a) {
  @apply text-primary-600 underline hover:text-primary-700;
  @apply dark:text-primary-400 dark:hover:text-primary-300;
}
.markdown-body :deep(table) {
  @apply my-3 w-full border-collapse text-sm;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  @apply border border-gray-200 px-2.5 py-1.5;
  @apply dark:border-dark-600;
}
.markdown-body :deep(th) {
  @apply bg-gray-100 font-semibold dark:bg-dark-700;
}
.markdown-body :deep(hr) {
  @apply my-4 border-gray-200 dark:border-dark-600;
}
</style>
