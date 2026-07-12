<!-- 工单状态徽章，确保列表和详情使用一致语义。 -->
<template>
  <span class="badge" :class="statusClass(status)">{{ statusLabel(status) }}</span>
</template>

<script setup lang="ts">
import { useTicketFormat } from '@/composables/useTicketFormat'
import type { TicketStatus } from '@/types/ticket'

defineProps<{ status: TicketStatus }>()

const { statusLabel, statusClass } = useTicketFormat()
</script>
