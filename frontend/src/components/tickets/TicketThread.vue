<!-- 工单不可变消息线程，正文始终按纯文本渲染。 -->
<template>
  <div class="ticket_thread space-y-3">
    <article
      v-for="message in messages"
      :key="message.id"
      class="rounded-md border px-4 py-3"
      :class="messageClass(message)"
    >
      <template v-if="message.kind === 'system'">
        <div class="flex items-center justify-center gap-2 text-xs font-medium text-gray-500 dark:text-dark-400">
          <Icon name="infoCircle" size="sm" />
          <span>{{ systemMessage(message) }}</span>
          <time :datetime="message.created_at">{{ formatDate(message.created_at) }}</time>
        </div>
      </template>

      <template v-else>
        <header class="mb-2 flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ authorLabel(message) }}</span>
            <span v-if="message.kind === 'internal'" class="badge badge-warning">{{ t('tickets.internal') }}</span>
          </div>
          <time class="text-xs text-gray-500 dark:text-dark-400" :datetime="message.created_at">
            {{ formatDate(message.created_at) }}
          </time>
        </header>

        <p v-if="message.body" class="whitespace-pre-wrap break-words text-sm leading-6 text-gray-700 dark:text-gray-200">
          {{ message.body }}
        </p>

        <ul v-if="message.attachments.length" class="mt-3 grid gap-2 sm:grid-cols-2">
          <li
            v-for="attachment in message.attachments"
            :key="attachment.id"
            class="flex min-w-0 items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-800"
          >
            <Icon name="document" size="sm" class="flex-shrink-0 text-gray-400" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-xs font-medium text-gray-700 dark:text-gray-200">{{ attachment.original_name }}</p>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                <template v-if="attachment.deleted_at">{{ t('tickets.attachments.deleted') }}</template>
                <template v-else>{{ formatBytes(attachment.size_bytes) }}</template>
              </p>
            </div>
            <button
              v-if="!attachment.deleted_at"
              type="button"
              class="btn-icon h-7 w-7 flex-shrink-0"
              :title="t('tickets.attachments.open')"
              @click="emit('openAttachment', attachment)"
            >
              <Icon :name="attachment.media_type.startsWith('image/') ? 'eye' : 'download'" size="sm" />
            </button>
            <slot v-if="!attachment.deleted_at" name="attachment-actions" :attachment="attachment" />
          </li>
        </ul>
      </template>
    </article>

    <div v-if="messages.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
      {{ t('tickets.emptyMessages') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDate, formatBytes } from '@/utils/format'
import type { TicketAttachment, TicketMessage } from '@/types/ticket'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ messages: TicketMessage[] }>()
const emit = defineEmits<{ openAttachment: [attachment: TicketAttachment] }>()
const { t, te } = useI18n()

function messageClass(message: TicketMessage): string {
  if (message.kind === 'system') return 'border-dashed border-gray-200 bg-transparent dark:border-dark-700'
  if (message.kind === 'internal') return 'border-amber-200 bg-amber-50/70 dark:border-amber-900/50 dark:bg-amber-950/20'
  return 'border-gray-200 bg-white/70 dark:border-dark-700 dark:bg-dark-900/70'
}

function authorLabel(message: TicketMessage): string {
  if (message.author_role === 'admin') return t('tickets.supportAgent')
  return message.author?.username || message.author?.email || t('tickets.unknownAuthor')
}

function systemMessage(message: TicketMessage): string {
  const event = typeof message.metadata?.event === 'string' ? message.metadata.event : 'updated'
  const key = `tickets.events.${event}`
  return te(key) ? t(key) : t('tickets.events.updated')
}
</script>
