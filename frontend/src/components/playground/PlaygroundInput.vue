<template>
  <div class="playground-input-wrap">
    <div class="playground-input">
      <div v-if="attachments.length > 0" class="playground-attachments">
        <div
          v-for="item in attachments"
          :key="item.id"
          class="playground-attachment"
        >
          <img
            v-if="item.kind === 'image' && item.dataUrl"
            :src="item.dataUrl"
            :alt="item.name"
            class="playground-attachment-thumb"
          />
          <span v-else class="playground-attachment-icon">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="8" y1="13" x2="16" y2="13" />
              <line x1="8" y1="17" x2="14" y2="17" />
            </svg>
          </span>
          <span class="playground-attachment-meta">
            <span class="playground-attachment-name">{{ item.name }}</span>
            <span class="playground-attachment-size">{{ formatFileSize(item.size) }}</span>
          </span>
          <button
            type="button"
            class="playground-attachment-remove"
            :title="t('playground.input.removeAttachment')"
            @click="removeAttachment(item.id)"
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
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
      </div>

      <!-- 主输入文本区 -->
      <textarea
        ref="textareaRef"
        v-model="text"
        :placeholder="placeholder || t('playground.input.placeholder')"
        :disabled="disabled"
        rows="1"
        class="playground-textarea"
        @keydown="handleKeydown"
        @input="autoResize"
        @paste="handlePaste"
      />

      <!-- 工具栏 -->
      <div class="playground-toolbar">
        <!-- 左侧：分组 + 模型选择 -->
        <div class="flex items-center gap-2 min-w-0 flex-1">
          <button
            type="button"
            class="playground-tool-btn"
            :disabled="disabled || isProcessingAttachment"
            :title="t('playground.input.attachImage')"
            @click="imageInputRef?.click()"
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
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <polyline points="21 15 16 10 5 21" />
            </svg>
          </button>

          <button
            type="button"
            class="playground-tool-btn"
            :disabled="disabled || isProcessingAttachment"
            :title="t('playground.input.attachDocument')"
            @click="documentInputRef?.click()"
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
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="12" y1="18" x2="12" y2="12" />
              <line x1="9" y1="15" x2="12" y2="12" />
              <line x1="15" y1="15" x2="12" y2="12" />
            </svg>
          </button>

          <!-- 分组下拉 -->
          <select
            :value="groupValue"
            :disabled="disabled || groups.length === 0"
            class="playground-select"
            @change="emitGroupChange(($event.target as HTMLSelectElement).value)"
          >
            <option v-if="groups.length === 0" value="">
              {{ t('playground.input.noGroup') }}
            </option>
            <option
              v-for="g in groups"
              :key="g.value"
              :value="g.value"
              :title="g.desc || g.label"
            >
              {{ g.label }} ({{ formatRatio(g.ratio) }})
            </option>
          </select>

          <!-- 模型下拉 -->
          <select
            :value="modelValue"
            :disabled="disabled || isModelLoading || models.length === 0"
            class="playground-select playground-select-model"
            @change="emitModelChange(($event.target as HTMLSelectElement).value)"
          >
            <option v-if="isModelLoading" value="">
              {{ t('playground.input.loadingModels') }}
            </option>
            <option v-else-if="models.length === 0" value="">
              {{ t('playground.input.noModel') }}
            </option>
            <option v-for="m in models" :key="m.value" :value="m.value">
              {{ m.label }}
            </option>
          </select>
        </div>

        <!-- 右侧：发送/停止按钮 -->
        <button
          v-if="isGenerating"
          type="button"
          class="playground-btn playground-btn-stop"
          :title="t('playground.input.stop')"
          @click="$emit('stop')"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="currentColor"
            class="h-4 w-4"
          >
            <rect x="6" y="6" width="12" height="12" rx="2" />
          </svg>
          <span class="hidden sm:inline">{{ t('playground.input.stop') }}</span>
        </button>
        <button
          v-else
          type="button"
          class="playground-btn playground-btn-send"
          :disabled="disabled || isProcessingAttachment || (!text.trim() && attachments.length === 0) || !modelValue"
          :title="t('playground.input.send')"
          @click="submit"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="h-4 w-4"
          >
            <path d="M22 2L11 13" />
            <path d="M22 2l-7 20-4-9-9-4 20-7z" />
          </svg>
          <span class="hidden sm:inline">{{ t('playground.input.send') }}</span>
        </button>
      </div>
    </div>

    <!-- 提示文字（fullHeight 模式下全局 Footer 被隐藏，法律声明并入此行） -->
    <p class="playground-hint">
      <template v-if="isProcessingAttachment">
        {{ t('playground.input.processingAttachment') }}
      </template>
      <template v-else>
        {{ t('playground.input.hint') }}
        <span class="playground-hint-legal">· {{ t('common.legalDisclaimer') }}</span>
      </template>
    </p>

    <input
      ref="imageInputRef"
      type="file"
      accept="image/png,image/jpeg,image/webp,image/gif"
      multiple
      class="hidden"
      @change="onImageInput"
    />
    <input
      ref="documentInputRef"
      type="file"
      :accept="DOCUMENT_ACCEPT"
      multiple
      class="hidden"
      @change="onDocumentInput"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import type {
  ModelOption,
  GroupOption,
  PlaygroundAttachment,
} from '@/types/playground'

const MAX_ATTACHMENTS = 6
const MAX_IMAGE_BYTES = 10 * 1024 * 1024
const MAX_DOCUMENT_BYTES = 20 * 1024 * 1024
const DOCUMENT_ACCEPT = [
  '.txt',
  '.md',
  '.markdown',
  '.json',
  '.csv',
  '.tsv',
  '.xml',
  '.yaml',
  '.yml',
  '.log',
  '.js',
  '.ts',
  '.tsx',
  '.jsx',
  '.vue',
  '.go',
  '.py',
  '.java',
  '.c',
  '.cpp',
  '.h',
  '.hpp',
  '.cs',
  '.rs',
  '.php',
  '.rb',
  '.sql',
  '.html',
  '.css',
].join(',')
const DOCUMENT_EXTENSIONS = new Set(
  DOCUMENT_ACCEPT.split(',').map((item) => item.slice(1))
)

interface Props {
  modelValue: string
  models: ModelOption[]
  isModelLoading?: boolean
  groupValue: string
  groups: GroupOption[]
  disabled?: boolean
  isGenerating?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  isModelLoading: false,
  disabled: false,
  isGenerating: false,
  placeholder: '',
})

const emit = defineEmits<{
  (e: 'submit', text: string, attachments: PlaygroundAttachment[]): void
  (e: 'stop'): void
  (e: 'modelChange', value: string): void
  (e: 'groupChange', value: string): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const imageInputRef = ref<HTMLInputElement | null>(null)
const documentInputRef = ref<HTMLInputElement | null>(null)
const attachments = ref<PlaygroundAttachment[]>([])
const isProcessingAttachment = ref(false)

function emitModelChange(v: string) {
  emit('modelChange', v)
}
function emitGroupChange(v: string) {
  emit('groupChange', v)
}

function formatRatio(r: number): string {
  if (r === 1) return '1x'
  if (r < 1) return `${r.toFixed(2)}x`
  return `${r}x`
}

function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function removeAttachment(id: string) {
  attachments.value = attachments.value.filter((item) => item.id !== id)
}

function submit() {
  if ((!text.value.trim() && attachments.value.length === 0) || props.disabled) return
  if (!props.modelValue) return
  emit('submit', text.value, attachments.value)
  text.value = ''
  attachments.value = []
  nextTick(autoResize)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    submit()
  }
}

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  const max = 240
  el.style.height = Math.min(el.scrollHeight, max) + 'px'
}

function canAppend(file: File, maxBytes: number): boolean {
  if (attachments.value.length >= MAX_ATTACHMENTS) {
    appStore.showError(t('playground.input.tooManyAttachments', { count: MAX_ATTACHMENTS }))
    return false
  }
  if (file.size > maxBytes) {
    appStore.showError(
      t('playground.input.fileTooLarge', {
        name: file.name,
        size: formatFileSize(maxBytes),
      })
    )
    return false
  }
  return true
}

function isSupportedDocument(file: File): boolean {
  const ext = file.name.split('.').pop()?.toLowerCase() || ''
  if (DOCUMENT_EXTENSIONS.has(ext)) return true
  return file.type.startsWith('text/')
}

function readAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

function readAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsText(file)
  })
}

/** 追加单个图片附件（文件选择与粘贴共用） */
async function appendImageFile(file: File) {
  if (!file.type.startsWith('image/')) {
    appStore.showError(t('playground.input.unsupportedImage', { name: file.name }))
    return
  }
  if (!canAppend(file, MAX_IMAGE_BYTES)) return
  const dataUrl = await readAsDataURL(file)
  attachments.value = [
    ...attachments.value,
    {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      kind: 'image',
      name: file.name,
      type: file.type || 'image/*',
      size: file.size,
      dataUrl,
    },
  ]
}

/** 追加单个文档附件（文件选择与粘贴共用） */
async function appendDocumentFile(file: File) {
  if (!isSupportedDocument(file)) {
    appStore.showError(t('playground.input.unsupportedDocument', { name: file.name }))
    return
  }
  if (!canAppend(file, MAX_DOCUMENT_BYTES)) return
  const content = await readAsText(file)
  if (!content.trim()) {
    appStore.showError(t('playground.input.emptyDocument', { name: file.name }))
    return
  }
  attachments.value = [
    ...attachments.value,
    {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      kind: 'document',
      name: file.name,
      type: file.type || 'text/plain',
      size: file.size,
      text: content,
    },
  ]
}

/** 批量处理文件列表（按类型分流到图片/文档管线） */
async function appendFiles(files: File[], imagesOnly = false) {
  isProcessingAttachment.value = true
  try {
    for (const file of files) {
      if (file.type.startsWith('image/')) {
        await appendImageFile(file)
      } else if (!imagesOnly) {
        await appendDocumentFile(file)
      } else {
        appStore.showError(t('playground.input.unsupportedImage', { name: file.name }))
      }
    }
  } catch {
    appStore.showError(t('playground.input.attachmentReadFailed'))
  } finally {
    isProcessingAttachment.value = false
  }
}

async function onImageInput(e: Event) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length === 0) return
  await appendFiles(files, true)
}

async function onDocumentInput(e: Event) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length === 0) return
  await appendFiles(files)
}

/**
 * 粘贴处理：剪贴板中带文件（截图/复制的文件）时拦截默认行为，
 * 走附件管线（图片 → dataUrl，文本类文档 → 读取内容）；纯文本粘贴不受影响。
 */
async function handlePaste(e: ClipboardEvent) {
  const files = Array.from(e.clipboardData?.files || [])
  if (files.length === 0) return // 纯文本 → 浏览器默认粘贴
  e.preventDefault()
  await appendFiles(files)
}

watch(() => text.value, () => nextTick(autoResize))
</script>

<style scoped>
.playground-input-wrap {
  @apply px-2 pb-3 md:pb-4 md:px-0;
}

.playground-input {
  @apply rounded-2xl border border-gray-200 bg-white shadow-sm transition-shadow;
  @apply dark:border-dark-600 dark:bg-dark-800;
}

.playground-input:focus-within {
  @apply border-primary-400 shadow-md;
  @apply dark:border-primary-500;
}

.playground-attachments {
  @apply flex gap-2 overflow-x-auto border-b border-gray-100 px-3 py-2.5;
  @apply dark:border-dark-700;
}

.playground-attachment {
  @apply inline-flex h-12 max-w-[220px] flex-shrink-0 items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-2;
  @apply dark:border-dark-600 dark:bg-dark-700;
}

.playground-attachment-thumb {
  @apply h-8 w-8 flex-shrink-0 rounded-md object-cover;
}

.playground-attachment-icon {
  @apply inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md bg-white text-gray-500;
  @apply dark:bg-dark-800 dark:text-gray-300;
}

.playground-attachment-meta {
  @apply min-w-0 flex-1;
}

.playground-attachment-name {
  @apply block truncate text-xs font-medium text-gray-700;
  @apply dark:text-gray-100;
}

.playground-attachment-size {
  @apply block text-[11px] text-gray-500;
  @apply dark:text-gray-400;
}

.playground-attachment-remove {
  @apply inline-flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md text-gray-400 transition-colors;
  @apply hover:bg-gray-200 hover:text-rose-600;
  @apply dark:hover:bg-dark-600 dark:hover:text-rose-300;
}

.playground-textarea {
  @apply w-full resize-none border-0 bg-transparent px-5 pt-4 pb-2 text-base leading-6;
  @apply text-gray-900 placeholder-gray-400;
  @apply focus:outline-none focus:ring-0;
  @apply dark:text-white dark:placeholder-gray-500;
  min-height: 56px;
  max-height: 240px;
}

.playground-textarea:disabled {
  @apply cursor-not-allowed opacity-60;
}

.playground-toolbar {
  @apply flex items-center justify-between gap-2 border-t border-gray-100 px-3 py-2.5;
  @apply dark:border-dark-700;
}

.playground-select {
  @apply flex h-8 min-w-0 cursor-pointer items-center rounded-lg border border-gray-200 bg-gray-50 px-2.5 text-xs font-medium text-gray-700 transition-colors;
  @apply hover:bg-gray-100;
  @apply dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600;
  max-width: 180px;
  text-overflow: ellipsis;
}

.playground-select-model {
  max-width: 220px;
}

.playground-select:disabled {
  @apply cursor-not-allowed opacity-50;
}

.playground-tool-btn {
  @apply inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-gray-600 transition-colors;
  @apply hover:bg-gray-100 hover:text-gray-900;
  @apply dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600 dark:hover:text-white;
  @apply disabled:cursor-not-allowed disabled:opacity-50;
}

.playground-btn {
  @apply inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium transition-all;
  @apply focus:outline-none focus:ring-2 focus:ring-offset-1;
}

.playground-btn-send {
  @apply bg-primary-600 text-white;
  @apply hover:bg-primary-700;
  @apply focus:ring-primary-500;
  @apply disabled:cursor-not-allowed disabled:opacity-50;
}

.playground-btn-stop {
  @apply bg-gray-900 text-white;
  @apply hover:bg-gray-800;
  @apply focus:ring-gray-600;
  @apply dark:bg-gray-200 dark:text-gray-900 dark:hover:bg-gray-100;
}

.playground-hint {
  @apply mt-1.5 text-center text-xs text-gray-400;
}

.playground-hint-legal {
  @apply text-amber-600/70 dark:text-amber-400/70;
}
</style>
