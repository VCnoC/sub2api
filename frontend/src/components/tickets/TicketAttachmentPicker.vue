<!-- 工单附件选择器，在上传前执行与服务端一致的数量和大小检查。 -->
<template>
  <div class="ticket_attachment_picker space-y-2">
    <div class="flex flex-wrap items-center gap-2">
      <label
        class="btn btn-secondary btn-sm cursor-pointer"
        :class="{ 'pointer-events-none opacity-60': disabled }"
      >
        <Icon name="upload" size="sm" />
        <span>{{ t('tickets.attachments.choose') }}</span>
        <input
          ref="inputRef"
          class="sr-only"
          type="file"
          multiple
          :disabled="disabled"
          :accept="acceptedTypes"
          @change="handleFiles"
        />
      </label>
      <span class="text-xs text-gray-500 dark:text-dark-400">
        {{ t('tickets.attachments.rules') }}
      </span>
    </div>

    <p v-if="error" role="alert" class="text-xs text-red-600 dark:text-red-400">{{ error }}</p>

    <ul v-if="modelValue.length" class="divide-y divide-gray-100 rounded-md border border-gray-200 dark:divide-dark-700 dark:border-dark-700">
      <li v-for="(file, index) in modelValue" :key="`${file.name}-${file.size}-${index}`" class="flex min-w-0 items-center gap-3 px-3 py-2">
        <Icon name="document" size="sm" class="flex-shrink-0 text-gray-400" />
        <span class="min-w-0 flex-1 truncate text-sm text-gray-700 dark:text-gray-200">{{ file.name }}</span>
        <span class="flex-shrink-0 text-xs text-gray-500 dark:text-dark-400">{{ formatBytes(file.size) }}</span>
        <button
          type="button"
          class="btn-icon h-7 w-7 text-gray-400 hover:text-red-600"
          :disabled="disabled"
          :title="t('tickets.attachments.remove')"
          @click="removeFile(index)"
        >
          <Icon name="x" size="sm" />
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatBytes } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'

const MAX_FILES = 5
const MAX_FILE_BYTES = 5 * 1024 * 1024
const MAX_TOTAL_BYTES = 20 * 1024 * 1024
const ALLOWED_EXTENSIONS = new Set(['jpg', 'jpeg', 'png', 'webp', 'txt', 'log', 'json'])

const props = withDefaults(defineProps<{ modelValue: File[]; disabled?: boolean }>(), {
  disabled: false,
})
const emit = defineEmits<{ 'update:modelValue': [files: File[]] }>()
const { t } = useI18n()
const inputRef = ref<HTMLInputElement | null>(null)
const error = ref('')
const acceptedTypes = '.jpg,.jpeg,.png,.webp,.txt,.log,.json'

function validationError(files: File[]): string {
  if (files.length > MAX_FILES) return t('tickets.errors.tooManyFiles')
  if (files.some((file) => file.size <= 0 || file.size > MAX_FILE_BYTES)) {
    return t('tickets.errors.fileTooLarge')
  }
  if (files.reduce((sum, file) => sum + file.size, 0) > MAX_TOTAL_BYTES) {
    return t('tickets.errors.filesTooLarge')
  }
  if (files.some((file) => !ALLOWED_EXTENSIONS.has(file.name.split('.').pop()?.toLowerCase() ?? ''))) {
    return t('tickets.errors.invalidFileType')
  }
  return ''
}

function handleFiles(event: Event): void {
  const selected = Array.from((event.target as HTMLInputElement).files ?? [])
  const next = [...props.modelValue, ...selected]
  error.value = validationError(next)
  if (!error.value) emit('update:modelValue', next)
  if (inputRef.value) inputRef.value.value = ''
}

function removeFile(index: number): void {
  error.value = ''
  emit('update:modelValue', props.modelValue.filter((_, itemIndex) => itemIndex !== index))
}
</script>
