/**
 * 对话广场状态管理 composable
 *
 * 职责：
 *   - config / parameterEnabled / messages 三套独立 localStorage 持久化
 *   - models / groups 内存态（从 API 加载，不持久化）
 *   - 配额溢出降级（5MB 边界）
 *   - 历史消息自动截断（保留最近 MAX_HISTORY_MESSAGES 条）
 *
 * 移植自 new-api/web/default/src/features/playground/hooks/use-playground-state.ts
 * 适配 Vue 3 Composition API
 */

import { ref, watch } from 'vue'
import {
  DEFAULT_CONFIG,
  DEFAULT_PARAMETER_ENABLED,
  STORAGE_KEYS,
  MAX_HISTORY_MESSAGES,
} from '@/constants/playground'
import type {
  Message,
  PlaygroundConfig,
  ParameterEnabled,
  ModelOption,
  GroupOption,
} from '@/types/playground'

// ==================== localStorage 工具 ====================

function loadFromStorage<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return fallback
    const parsed = JSON.parse(raw)
    return parsed as T
  } catch {
    return fallback
  }
}

function saveToStorage(key: string, value: unknown): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value))
    return true
  } catch (e) {
    // 配额溢出（QuotaExceededError）
    // eslint-disable-next-line no-console
    console.warn(`[playground] localStorage save failed for ${key}`, e)
    return false
  }
}

// ==================== 主 composable ====================

export function usePlaygroundState() {
  // 初始化：localStorage merge defaults
  const config = ref<PlaygroundConfig>({
    ...DEFAULT_CONFIG,
    ...loadFromStorage<Partial<PlaygroundConfig>>(STORAGE_KEYS.CONFIG, {}),
  })

  const parameterEnabled = ref<ParameterEnabled>({
    ...DEFAULT_PARAMETER_ENABLED,
    ...loadFromStorage<Partial<ParameterEnabled>>(
      STORAGE_KEYS.PARAMETER_ENABLED,
      {}
    ),
  })

  const messages = ref<Message[]>(
    loadFromStorage<Message[]>(STORAGE_KEYS.MESSAGES, [])
  )

  // 模型/分组列表（不持久化，每次进入页面重新拉取）
  const models = ref<ModelOption[]>([])
  const groups = ref<GroupOption[]>([])

  // ==================== 持久化监听 ====================

  watch(
    config,
    (val) => {
      saveToStorage(STORAGE_KEYS.CONFIG, val)
    },
    { deep: true }
  )

  watch(
    parameterEnabled,
    (val) => {
      saveToStorage(STORAGE_KEYS.PARAMETER_ENABLED, val)
    },
    { deep: true }
  )

  watch(
    messages,
    (val) => {
      // 配额防护：超过 MAX_HISTORY_MESSAGES 时自动截断头部
      let toSave = val
      if (val.length > MAX_HISTORY_MESSAGES) {
        toSave = val.slice(val.length - MAX_HISTORY_MESSAGES)
      }
      const ok = saveToStorage(STORAGE_KEYS.MESSAGES, toSave)
      // 配额溢出时进一步截断重试
      if (!ok && toSave.length > 20) {
        const truncated = toSave.slice(-20)
        saveToStorage(STORAGE_KEYS.MESSAGES, truncated)
      }
    },
    { deep: true }
  )

  // ==================== 操作 API ====================

  /** 更新单个配置项 */
  function updateConfig<K extends keyof PlaygroundConfig>(
    key: K,
    value: PlaygroundConfig[K]
  ) {
    config.value = { ...config.value, [key]: value }
  }

  /** 更新单个参数启用开关 */
  function updateParameterEnabled(
    key: keyof ParameterEnabled,
    value: boolean
  ) {
    parameterEnabled.value = { ...parameterEnabled.value, [key]: value }
  }

  /** 整体覆盖消息列表 */
  function setMessages(next: Message[]) {
    messages.value = next
  }

  /** 函数式更新消息列表 */
  function updateMessages(updater: (prev: Message[]) => Message[]) {
    messages.value = updater(messages.value)
  }

  /** 清空对话历史 */
  function clearMessages() {
    messages.value = []
  }

  /** 重置配置为默认值 */
  function resetConfig() {
    config.value = { ...DEFAULT_CONFIG }
    parameterEnabled.value = { ...DEFAULT_PARAMETER_ENABLED }
  }

  /** 导出当前配置为 JSON */
  function exportConfig(): string {
    return JSON.stringify(
      {
        config: config.value,
        parameterEnabled: parameterEnabled.value,
        exportedAt: new Date().toISOString(),
        version: 1,
      },
      null,
      2
    )
  }

  /** 导入 JSON 配置（容错：仅吃已知字段） */
  function importConfig(json: string): boolean {
    try {
      const data = JSON.parse(json) as {
        config?: Partial<PlaygroundConfig>
        parameterEnabled?: Partial<ParameterEnabled>
      }
      if (data.config) {
        config.value = { ...DEFAULT_CONFIG, ...data.config }
      }
      if (data.parameterEnabled) {
        parameterEnabled.value = {
          ...DEFAULT_PARAMETER_ENABLED,
          ...data.parameterEnabled,
        }
      }
      return true
    } catch {
      return false
    }
  }

  return {
    // State
    config,
    parameterEnabled,
    messages,
    models,
    groups,
    // Actions
    updateConfig,
    updateParameterEnabled,
    setMessages,
    updateMessages,
    clearMessages,
    resetConfig,
    exportConfig,
    importConfig,
  }
}
