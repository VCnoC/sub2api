<template>
  <AppLayout full-height>
    <div class="playground-page">
      <!-- 顶部工具条 -->
      <div class="playground-header animate-slide-up [animation-fill-mode:backwards]" style="animation-delay: 0ms">
        <div class="playground-header-left">
          <div class="flex items-center gap-2">
            <span class="h-4 w-1 rounded-full bg-gradient-to-b from-primary-400 to-primary-600"></span>
            <h1 class="playground-title">
              {{ t('playground.title') }}
            </h1>
          </div>
          <span v-if="config.group" class="playground-subtitle">
            {{ t('playground.subtitle', { group: config.group }) }}
          </span>
        </div>

        <div class="playground-header-actions">
          <!-- 系统提示词 -->
          <button
            type="button"
            class="header-btn"
            :class="{ 'header-btn-active': config.systemPrompt }"
            :title="t('playground.actions.systemPrompt')"
            @click="systemPromptDialogOpen = true"
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
              <path d="M14 9V5a3 3 0 0 0-6 0v4" />
              <rect x="2" y="9" width="20" height="13" rx="2" />
            </svg>
            <span class="hidden md:inline">{{ t('playground.actions.systemPrompt') }}</span>
          </button>

          <!-- 参数面板 -->
          <button
            type="button"
            class="header-btn"
            :title="t('playground.actions.params')"
            @click="paramDialogOpen = true"
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
              <line x1="4" y1="21" x2="4" y2="14" />
              <line x1="4" y1="10" x2="4" y2="3" />
              <line x1="12" y1="21" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12" y2="3" />
              <line x1="20" y1="21" x2="20" y2="16" />
              <line x1="20" y1="12" x2="20" y2="3" />
              <line x1="1" y1="14" x2="7" y2="14" />
              <line x1="9" y1="8" x2="15" y2="8" />
              <line x1="17" y1="16" x2="23" y2="16" />
            </svg>
            <span class="hidden md:inline">{{ t('playground.actions.params') }}</span>
          </button>

          <!-- 导入 -->
          <button
            type="button"
            class="header-btn-icon"
            :title="t('playground.actions.import')"
            @click="onImportClick"
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
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" y1="3" x2="12" y2="15" />
            </svg>
          </button>

          <!-- 导出 -->
          <button
            type="button"
            class="header-btn-icon"
            :title="t('playground.actions.export')"
            @click="onExport"
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
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
          </button>

          <!-- 清空 -->
          <button
            type="button"
            class="header-btn-icon header-btn-danger"
            :disabled="messages.length === 0"
            :title="t('playground.actions.clear')"
            @click="confirmClear"
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
              <polyline points="3 6 5 6 21 6" />
              <path
                d="M19 6l-2 14a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L5 6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              />
            </svg>
          </button>
        </div>
      </div>

      <!-- 主体：会话侧栏 + （消息区 + 输入区） -->
      <div class="playground-body animate-slide-up [animation-fill-mode:backwards]" style="animation-delay: 50ms">
        <!-- 左侧会话列表（移动端隐藏）；mb 与右侧输入区的底部留白对齐 -->
        <ConversationSidebar
          class="mb-3 hidden md:mb-4 md:flex"
          :conversations="conversations"
          :active-id="activeConversationId"
          :loading="isLoadingList"
          @select="handleSelectConversation"
          @create="handleNewConversation"
          @remove="handleRemoveConversation"
        />

        <div class="playground-main">
        <PlaygroundChat
          :messages="messages"
          :is-generating="isGenerating"
          :editing-key="editingKey"
          :system-prompt="config.systemPrompt"
          :version-index-map="versionIndexMap"
          @regenerate="handleRegenerate"
          @edit="handleEdit"
          @remove="handleRemove"
          @save-edit="handleSaveEdit"
          @save-edit-and-submit="handleSaveEditAndSubmit"
          @cancel-edit="handleCancelEdit"
          @switch-version="handleSwitchVersion"
        />

        <div class="mx-auto w-full max-w-4xl">
          <PlaygroundInput
            :model-value="config.model"
            :models="modelOptions"
            :is-model-loading="isLoadingModels"
            :group-value="config.group"
            :groups="groupOptions"
            :disabled="isGenerating"
            :is-generating="isGenerating"
            @submit="handleSend"
            @stop="stopGeneration"
            @model-change="onModelChange"
            @group-change="onGroupChange"
          />
        </div>
        </div>
      </div>

      <!-- 系统提示词对话框 -->
      <div v-if="systemPromptDialogOpen" class="dialog-overlay" @click.self="systemPromptDialogOpen = false">
        <div class="dialog-panel animate-slide-up">
          <div class="flex items-center gap-2">
            <span class="h-4 w-1 rounded-full bg-gradient-to-b from-primary-400 to-primary-600"></span>
            <h3 class="dialog-title">{{ t('playground.systemPrompt.title') }}</h3>
          </div>
          <p class="dialog-desc">{{ t('playground.systemPrompt.desc') }}</p>
          <textarea
            v-model="systemPromptDraft"
            rows="8"
            class="dialog-textarea"
            :placeholder="t('playground.systemPrompt.placeholder')"
          />
          <div class="dialog-actions">
            <button class="msg-btn msg-btn-outline" @click="systemPromptDialogOpen = false">
              {{ t('playground.message.cancel') }}
            </button>
            <button class="msg-btn msg-btn-primary bg-gradient-to-r from-primary-500 to-primary-600 shadow-lg shadow-primary-500/25 hover:from-primary-600 hover:to-primary-700" @click="onSaveSystemPrompt">
              {{ t('playground.message.save') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 参数面板对话框 -->
      <div v-if="paramDialogOpen" class="dialog-overlay" @click.self="paramDialogOpen = false">
        <div class="dialog-panel dialog-panel-wide animate-slide-up">
          <div class="flex items-center gap-2">
            <span class="h-4 w-1 rounded-full bg-gradient-to-b from-primary-400 to-primary-600"></span>
            <h3 class="dialog-title">{{ t('playground.params.title') }}</h3>
          </div>
          <p class="dialog-desc">{{ t('playground.params.desc') }}</p>

          <div class="dialog-params">
            <ParamSlider
              v-model:value="config.temperature"
              v-model:enabled="parameterEnabled.temperature"
              :label="t('playground.params.temperature')"
              :min="0"
              :max="2"
              :step="0.01"
            />
            <ParamSlider
              v-model:value="config.top_p"
              v-model:enabled="parameterEnabled.top_p"
              :label="t('playground.params.topP')"
              :min="0"
              :max="1"
              :step="0.01"
            />
            <ParamSlider
              v-model:value="config.max_tokens"
              v-model:enabled="parameterEnabled.max_tokens"
              :label="t('playground.params.maxTokens')"
              :min="1"
              :max="32768"
              :step="1"
              integer
            />
            <ParamSlider
              v-model:value="config.frequency_penalty"
              v-model:enabled="parameterEnabled.frequency_penalty"
              :label="t('playground.params.frequencyPenalty')"
              :min="-2"
              :max="2"
              :step="0.01"
            />
            <ParamSlider
              v-model:value="config.presence_penalty"
              v-model:enabled="parameterEnabled.presence_penalty"
              :label="t('playground.params.presencePenalty')"
              :min="-2"
              :max="2"
              :step="0.01"
            />

            <!-- Stream toggle -->
            <div class="dialog-toggle-row">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-200">
                {{ t('playground.params.stream') }}
              </label>
              <button
                type="button"
                class="toggle-switch"
                :class="{ 'toggle-switch-on': config.stream }"
                @click="config.stream = !config.stream"
              >
                <span class="toggle-knob" />
              </button>
            </div>
          </div>

          <div class="dialog-actions">
            <button class="msg-btn msg-btn-outline" @click="paramDialogOpen = false">
              {{ t('playground.message.cancel') }}
            </button>
            <button class="msg-btn msg-btn-outline" @click="onResetParams">
              {{ t('playground.params.reset') }}
            </button>
            <button class="msg-btn msg-btn-primary bg-gradient-to-r from-primary-500 to-primary-600 shadow-lg shadow-primary-500/25 hover:from-primary-600 hover:to-primary-700" @click="paramDialogOpen = false">
              {{ t('playground.message.save') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 隐藏的文件输入（用于导入） -->
      <input
        ref="fileInputRef"
        type="file"
        accept="application/json"
        class="hidden"
        @change="onImportFile"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import PlaygroundChat from '@/components/playground/PlaygroundChat.vue'
import PlaygroundInput from '@/components/playground/PlaygroundInput.vue'
import ConversationSidebar from '@/components/playground/ConversationSidebar.vue'
import ParamSlider from '@/components/playground/ParamSlider.vue'
import { usePlaygroundState } from '@/composables/playground/usePlaygroundState'
import { useConversations } from '@/composables/playground/useConversations'
import {
  useChatHandler,
  createUserMessage,
  createLoadingAssistantMessage,
} from '@/composables/playground/useChatHandler'
import { getAvailableModels } from '@/api/playground'
import { getAvailable as fetchGroups } from '@/api/groups'
import { useAppStore } from '@/stores/app'
import type {
  Message,
  ModelOption,
  GroupOption,
  PlaygroundAttachment,
} from '@/types/playground'

const { t } = useI18n()
const appStore = useAppStore()

const {
  config,
  parameterEnabled,
  messages,
  models,
  groups,
  updateMessages,
  setMessages,
  clearMessages,
  resetConfig,
  exportConfig,
  importConfig,
} = usePlaygroundState()

const { sendChat, stopGeneration, isGenerating } = useChatHandler({
  config,
  parameterEnabled,
  messages,
  updateMessages,
})

// ==================== 多会话管理 ====================

const {
  conversations,
  activeConversationId,
  isLoadingList,
  loadConversations,
  selectConversation,
  newConversation,
  removeConversation,
  scheduleSave,
} = useConversations({
  getMessages: () => messages.value,
  setMessages,
  getModel: () => config.value.model,
  getGroupName: () => config.value.group,
  defaultTitle: () => t('playground.conversations.defaultTitle'),
})

// 流式回复结束（true → false）后防抖保存当前会话
watch(isGenerating, (generating, wasGenerating) => {
  if (wasGenerating && !generating) {
    scheduleSave()
  }
})

/** 生成中切换/新建前的收尾：停止本轮生成（部分回复定稿）并排入保存队列 */
function stopAndQueueSave() {
  if (!isGenerating.value) return
  stopGeneration()
  // 设置待保存任务：后续 selectConversation/newConversation 内部 flush 时
  // 会在切换前把部分回复落库到原会话
  scheduleSave()
}

async function handleSelectConversation(id: number) {
  stopAndQueueSave()
  try {
    await selectConversation(id)
  } catch {
    appStore.showError(t('playground.conversations.loadFailed'))
  }
}

async function handleNewConversation() {
  stopAndQueueSave()
  await newConversation()
}

async function handleRemoveConversation(id: number) {
  if (!window.confirm(t('playground.conversations.confirmDelete'))) return
  // 删除的是正在生成的当前会话 → 先停止生成
  if (isGenerating.value && id === activeConversationId.value) {
    stopGeneration()
  }
  try {
    await removeConversation(id)
  } catch {
    appStore.showError(t('playground.conversations.deleteFailed'))
  }
}

// ==================== 模型 / 分组加载 ====================

const isLoadingModels = ref(false)
const groupOptions = computed<GroupOption[]>(() => groups.value)
const modelOptions = computed<ModelOption[]>(() => models.value)

async function loadGroups() {
  try {
    const list = await fetchGroups()
    groups.value = list.map((g): GroupOption => ({
      label: g.name,
      value: g.name,
      ratio: g.rate_multiplier ?? 1,
      platform: g.platform,
      desc: g.description ?? undefined,
    }))
    // 兜底：当前 group 不在可用列表时选第一个
    if (!groups.value.some((g) => g.value === config.value.group) && groups.value.length > 0) {
      config.value.group = groups.value[0].value
    }
  } catch (e) {
    appStore.showError(t('playground.error.loadGroupsFailed'))
  }
}

async function loadModels(group: string) {
  if (!group) return
  isLoadingModels.value = true
  try {
    const list = await getAvailableModels(group)
    models.value = list.map((m): ModelOption => ({
      label: m.id,
      value: m.id,
      platform: m.platform,
    }))
    if (!models.value.some((m) => m.value === config.value.model) && models.value.length > 0) {
      config.value.model = models.value[0].value
    }
  } catch {
    appStore.showError(t('playground.error.loadModelsFailed'))
    models.value = []
  } finally {
    isLoadingModels.value = false
  }
}

onMounted(async () => {
  await loadGroups()
  if (config.value.group) {
    await loadModels(config.value.group)
  }
  // 会话列表加载放最后：旧 localStorage 迁移需要 model/group 已就绪
  try {
    await loadConversations()
  } catch {
    appStore.showError(t('playground.conversations.loadFailed'))
  }
})

// 分组变化时刷新模型
function onGroupChange(value: string) {
  config.value.group = value
  config.value.model = ''
  loadModels(value)
}
function onModelChange(value: string) {
  config.value.model = value
}

// ==================== 消息操作 ====================

const editingKey = ref<string | null>(null)
const versionIndexMap = ref<Record<string, number>>({})

function handleSend(text: string, attachments: PlaygroundAttachment[] = []) {
  if (!text.trim() && attachments.length === 0) return
  const userMsg = createUserMessage(text, attachments)
  const placeholder = createLoadingAssistantMessage()
  const next = [...messages.value, userMsg, placeholder]
  setMessages(next)
  sendChat(next)
}

function handleRegenerate(target: Message) {
  // 找到该 AI 消息的位置；取它之前所有的消息（含用户最后一条），追加新的占位
  const idx = messages.value.findIndex((m) => m.key === target.key)
  if (idx === -1) return
  const upstream = messages.value.slice(0, idx)
  const placeholder = createLoadingAssistantMessage()
  // 把新占位的 key 替换为目标 key（保留位置 + 多版本累计）
  placeholder.key = target.key
  placeholder.versions = [
    { id: `v${target.versions.length}`, content: '' },
    ...target.versions,
  ]
  const next = [...upstream, placeholder]
  setMessages(next)
  // 把多版本切换索引指向新版本（0 = 最新）
  versionIndexMap.value = { ...versionIndexMap.value, [target.key]: 0 }
  sendChat(next)
}

function handleEdit(target: Message) {
  if (isGenerating.value) {
    appStore.showInfo(t('playground.message.waitGeneration'))
    return
  }
  editingKey.value = target.key
}

function handleCancelEdit() {
  editingKey.value = null
}

function handleSaveEdit(key: string, content: string) {
  updateMessages((prev) =>
    prev.map((m) =>
      m.key === key
        ? { ...m, versions: [{ ...m.versions[0], content }, ...m.versions.slice(1)] }
        : m
    )
  )
  editingKey.value = null
  scheduleSave()
}

function handleSaveEditAndSubmit(key: string, content: string) {
  const idx = messages.value.findIndex((m) => m.key === key)
  if (idx === -1) return
  // 替换内容并截断后续，再追加新占位
  const updated = messages.value.slice(0, idx + 1).map((m, i) =>
    i === idx
      ? { ...m, versions: [{ ...m.versions[0], content }, ...m.versions.slice(1)] }
      : m
  )
  const placeholder = createLoadingAssistantMessage()
  const next = [...updated, placeholder]
  setMessages(next)
  editingKey.value = null
  sendChat(next)
}

function handleRemove(target: Message) {
  if (!window.confirm(t('playground.message.confirmDelete'))) return
  updateMessages((prev) => prev.filter((m) => m.key !== target.key))
  scheduleSave()
}

function handleSwitchVersion(key: string, index: number) {
  const msg = messages.value.find((m) => m.key === key)
  if (!msg) return
  if (index < 0 || index >= msg.versions.length) return
  versionIndexMap.value = { ...versionIndexMap.value, [key]: index }
  // 把对应版本提到 versions[0]（用于下次发送时取 versions[0].content）
  updateMessages((prev) =>
    prev.map((m) => {
      if (m.key !== key) return m
      const next = [...m.versions]
      const [picked] = next.splice(index, 1)
      next.unshift(picked)
      return { ...m, versions: next }
    })
  )
  versionIndexMap.value[key] = 0
}

// ==================== 顶部工具栏操作 ====================

const systemPromptDialogOpen = ref(false)
const systemPromptDraft = ref('')

watch(systemPromptDialogOpen, (open) => {
  if (open) systemPromptDraft.value = config.value.systemPrompt
})

function onSaveSystemPrompt() {
  config.value.systemPrompt = systemPromptDraft.value
  systemPromptDialogOpen.value = false
  appStore.showSuccess(t('playground.systemPrompt.saved'))
}

const paramDialogOpen = ref(false)

function onResetParams() {
  resetConfig()
  appStore.showSuccess(t('playground.params.resetDone'))
}

function confirmClear() {
  if (!window.confirm(t('playground.actions.confirmClear'))) return
  clearMessages()
  // 同步清空当前会话的服务端记录
  scheduleSave()
  appStore.showSuccess(t('playground.actions.cleared'))
}

const fileInputRef = ref<HTMLInputElement | null>(null)
function onImportClick() {
  fileInputRef.value?.click()
}

function onImportFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (ev) => {
    const text = ev.target?.result as string
    if (importConfig(text)) {
      appStore.showSuccess(t('playground.actions.imported'))
    } else {
      appStore.showError(t('playground.actions.importFailed'))
    }
  }
  reader.readAsText(file)
  ;(e.target as HTMLInputElement).value = ''
}

function onExport() {
  const json = exportConfig()
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sub2api-playground-config-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  appStore.showSuccess(t('playground.actions.exported'))
}

// 状态机维护：流式开始 / 结束时同步 versionIndexMap 不需要额外动作
// 因为 handleRegenerate 已把新版本放在 versions[0]，MessageItem 默认显示 versions[0]
</script>

<style scoped>
.playground-page {
  /* 高度由 AppLayout fullHeight 模式的 flex 链路提供，无需魔法数字 */
  @apply flex min-h-0 flex-1 flex-col gap-0;
}

.playground-header {
  @apply flex items-center justify-between gap-2 border-b border-gray-200 px-2 pb-3 dark:border-dark-700;
}

.playground-header-left {
  @apply flex items-center gap-3 min-w-0;
}

.playground-title {
  @apply text-lg font-semibold text-gray-900 dark:text-white;
}

.playground-subtitle {
  @apply hidden truncate text-xs text-gray-500 md:inline dark:text-gray-400;
}

.playground-header-actions {
  @apply flex items-center gap-1.5;
}

.header-btn {
  @apply inline-flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 text-xs font-medium text-gray-700 transition-colors;
  @apply hover:bg-gray-50;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:bg-dark-700;
}

.header-btn-active {
  @apply border-primary-400 bg-primary-50 text-primary-700;
  @apply dark:border-primary-500 dark:bg-primary-900/30 dark:text-primary-300;
}

.header-btn-icon {
  @apply inline-flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 bg-white text-gray-600 transition-colors;
  @apply hover:bg-gray-50;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700;
  @apply disabled:cursor-not-allowed disabled:opacity-50;
}

.header-btn-danger {
  @apply hover:border-rose-300 hover:bg-rose-50 hover:text-rose-600;
  @apply dark:hover:border-rose-700 dark:hover:bg-rose-900/20 dark:hover:text-rose-300;
}

.playground-body {
  /* 横向布局：左侧会话栏 + 右侧聊天主区 */
  @apply flex min-h-0 flex-1 flex-row gap-3 pt-3;
}

.playground-main {
  @apply flex min-h-0 min-w-0 flex-1 flex-col;
}

/* Dialog */
.dialog-overlay {
  @apply fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4;
}

.dialog-panel {
  @apply w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl;
  @apply dark:bg-dark-800;
}

.dialog-panel-wide {
  @apply max-w-lg;
}

.dialog-title {
  @apply text-lg font-semibold text-gray-900 dark:text-white;
}

.dialog-desc {
  @apply mt-1 text-sm text-gray-500 dark:text-gray-400;
}

.dialog-textarea {
  @apply mt-4 w-full resize-none rounded-lg border border-gray-200 bg-white p-3 text-sm leading-6 focus:border-primary-400 focus:outline-none focus:ring-1 focus:ring-primary-400;
  @apply dark:border-dark-600 dark:bg-dark-700 dark:text-gray-100;
}

.dialog-params {
  @apply mt-4 max-h-96 space-y-4 overflow-y-auto pr-2;
}

.dialog-actions {
  @apply mt-6 flex justify-end gap-2;
}

.dialog-toggle-row {
  @apply flex items-center justify-between;
}

.msg-btn {
  @apply inline-flex h-9 items-center justify-center rounded-lg px-4 text-sm font-medium transition-colors;
}
.msg-btn-primary {
  @apply bg-primary-600 text-white hover:bg-primary-700;
}
.msg-btn-outline {
  @apply border border-gray-200 text-gray-700 hover:bg-gray-50;
  @apply dark:border-dark-600 dark:text-gray-200 dark:hover:bg-dark-700;
}

.toggle-switch {
  @apply relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full bg-gray-300 transition-colors;
  @apply dark:bg-dark-600;
}
.toggle-switch-on {
  @apply bg-primary-600;
}
.toggle-knob {
  @apply pointer-events-none inline-block h-4 w-4 translate-x-0.5 translate-y-0.5 transform rounded-full bg-white shadow ring-0 transition;
}
.toggle-switch-on .toggle-knob {
  @apply translate-x-4;
}
</style>
