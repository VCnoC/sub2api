<template>
  <div class="playground-input-wrap">
    <div class="playground-input">
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
      />

      <!-- 工具栏 -->
      <div class="playground-toolbar">
        <!-- 左侧：分组 + 模型选择 -->
        <div class="flex items-center gap-2 min-w-0 flex-1">
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
          :disabled="disabled || !text.trim() || !modelValue"
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

    <!-- 提示文字 -->
    <p class="playground-hint">
      {{ t('playground.input.hint') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ModelOption, GroupOption } from '@/types/playground'

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
  (e: 'submit', text: string): void
  (e: 'stop'): void
  (e: 'modelChange', value: string): void
  (e: 'groupChange', value: string): void
}>()

const { t } = useI18n()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

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

function submit() {
  if (!text.value.trim() || props.disabled) return
  if (!props.modelValue) return
  emit('submit', text.value)
  text.value = ''
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
</style>
