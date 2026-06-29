<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import Icon from '@/components/icons/Icon.vue'

export interface AnnouncementBarProps {
  /** 通知唯一标识，用于 localStorage 记忆关闭状态 */
  id: string
  /** 通知文本，支持 HTML */
  content: string
  /** 背景色类名 */
  bgClass?: string
  /** 文字色类名 */
  textClass?: string
  /** 左侧徽章文字 */
  badge?: string
  /** 左侧徽章背景色 */
  badgeClass?: string
  /** 是否允许关闭 */
  dismissible?: boolean
}

const props = withDefaults(defineProps<AnnouncementBarProps>(), {
  bgClass: 'bg-[#1a1f2e] dark:bg-[#0f1219]',
  textClass: 'text-slate-300 dark:text-slate-300',
  badge: '通知',
  badgeClass: 'bg-indigo-600 text-white',
  dismissible: true,
})

const emit = defineEmits<{
  close: []
}>()

const visible = ref(true)
const storageKey = computed(() => `announcement-dismissed-${props.id}`)

onMounted(() => {
  if (props.dismissible) {
    const dismissed = localStorage.getItem(storageKey.value)
    if (dismissed === '1') {
      visible.value = false
    }
  }
})

function handleClose() {
  visible.value = false
  if (props.dismissible) {
    localStorage.setItem(storageKey.value, '1')
  }
  emit('close')
}
</script>

<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="-translate-y-full opacity-0"
    enter-to-class="translate-y-0 opacity-100"
    leave-active-class="transition-all duration-200 ease-in"
    leave-from-class="translate-y-0 opacity-100"
    leave-to-class="-translate-y-full opacity-0"
  >
    <div
      v-if="visible"
      :class="[
        'relative z-[60] flex items-center justify-center gap-3',
        'px-4 py-2 text-xs sm:text-sm',
        'border-b border-white/5',
        bgClass,
        textClass,
      ]"
    >
      <!-- 左侧徽章（绝对定位，不占据居中空间） -->
      <span
        v-if="badge"
        :class="[
          'absolute left-4 rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider',
          badgeClass,
        ]"
      >
        {{ badge }}
      </span>

      <!-- 内容区：纯展示、居中 -->
      <div
        class="truncate text-center"
        :title="content.replace(/<[^>]+>/g, '')"
        v-html="content"
      />

      <!-- 关闭按钮 -->
      <button
        v-if="dismissible"
        type="button"
        class="absolute right-2 rounded p-1 text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
        aria-label="关闭通知"
        @click="handleClose"
      >
        <Icon name="x" size="sm" />
      </button>
    </div>
  </Transition>
</template>
