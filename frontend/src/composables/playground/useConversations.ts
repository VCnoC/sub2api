/**
 * 对话广场多会话管理 composable
 *
 * 职责：
 *   - 会话列表加载 / 新建 / 切换（懒加载详情）/ 删除 / 改名
 *   - 消息保存：流式回复结束后由调用方触发 scheduleSave（防抖 1s）
 *   - beforeunload 时同步 flush，降低最后一轮消息丢失风险
 *   - 旧版 localStorage 单会话数据的一次性迁移导入
 *
 * 设计约束：
 *   - 消息列表（messages ref）仍由 usePlaygroundState 持有，本 composable
 *     通过 options 注入的 getter/setter 读写，避免双状态源
 *   - ⚠️ 后端 Update 语义：model/group_name 缺省 = 清空！所有 PUT 请求
 *     必须带上 model 与 group_name（见 UpdateConversationRequest 注释）
 */

import { ref, onMounted, onBeforeUnmount } from 'vue'
import { playgroundAPI } from '@/api/playground'
import {
  CONVERSATION_SAVE_DEBOUNCE_MS,
  CONVERSATION_TITLE_MAX_CHARS,
  STORAGE_KEYS,
} from '@/constants/playground'
import type { ConversationSummary, Message } from '@/types/playground'

interface UseConversationsOptions {
  /** 读取当前消息列表（来自 usePlaygroundState） */
  getMessages: () => Message[]
  /** 覆盖当前消息列表（切换会话时调用） */
  setMessages: (messages: Message[]) => void
  /** 当前选中的模型 ID（保存时随消息一并写入） */
  getModel: () => string
  /** 当前选中的分组名（保存时随消息一并写入） */
  getGroupName: () => string
  /** 默认会话标题（i18n 注入，如「新对话」） */
  defaultTitle: () => string
}

/** 从首条用户消息提取标题（前 N 个字符） */
function deriveTitle(messages: Message[]): string | null {
  const firstUser = messages.find((m) => m.from === 'user')
  const content = firstUser?.versions?.[0]?.content?.trim()
  if (!content) return null
  const chars = Array.from(content)
  return chars.slice(0, CONVERSATION_TITLE_MAX_CHARS).join('')
}

export function useConversations(options: UseConversationsOptions) {
  /** 会话摘要列表（按最后活动时间倒序） */
  const conversations = ref<ConversationSummary[]>([])
  /** 当前激活的会话 ID；null 表示尚未落库的「草稿」状态 */
  const activeConversationId = ref<number | null>(null)
  const isLoadingList = ref(false)
  const isLoadingDetail = ref(false)
  const isSaving = ref(false)

  let saveTimer: ReturnType<typeof setTimeout> | null = null
  /** 保存链：串行化并发的 saveNow 调用，避免乱序覆盖 */
  let saveChain: Promise<void> = Promise.resolve()

  // ==================== 列表 ====================

  /** 加载会话列表（含旧 localStorage 数据迁移） */
  async function loadConversations() {
    isLoadingList.value = true
    try {
      await migrateLegacyMessages()
      conversations.value = await playgroundAPI.listConversations()
    } finally {
      isLoadingList.value = false
    }
  }

  /** 把指定会话提到列表顶部并刷新活动时间（本地乐观更新） */
  function touchSummary(id: number, patch?: Partial<ConversationSummary>) {
    const idx = conversations.value.findIndex((c) => c.id === id)
    if (idx < 0) return
    const updated = {
      ...conversations.value[idx],
      ...patch,
      last_activity_at: new Date().toISOString(),
    }
    conversations.value.splice(idx, 1)
    conversations.value.unshift(updated)
  }

  // ==================== 切换 / 新建 / 删除 / 改名 ====================

  /** 切换会话：先 flush 未保存内容，再懒加载目标会话详情 */
  async function selectConversation(id: number) {
    if (id === activeConversationId.value) return
    await flushSave()
    isLoadingDetail.value = true
    try {
      const detail = await playgroundAPI.getConversation(id)
      activeConversationId.value = id
      options.setMessages(detail.messages ?? [])
    } finally {
      isLoadingDetail.value = false
    }
  }

  /** 新建会话：flush 当前内容后进入空白草稿态（首次保存时才落库） */
  async function newConversation() {
    await flushSave()
    activeConversationId.value = null
    options.setMessages([])
  }

  /** 删除会话；若删除的是当前会话则回到草稿态 */
  async function removeConversation(id: number) {
    await playgroundAPI.deleteConversation(id)
    conversations.value = conversations.value.filter((c) => c.id !== id)
    if (activeConversationId.value === id) {
      activeConversationId.value = null
      options.setMessages([])
    }
  }

  /** 重命名会话（同时带上 model/group_name，防止被后端清空语义误伤） */
  async function renameConversation(id: number, title: string) {
    const summary = conversations.value.find((c) => c.id === id)
    await playgroundAPI.updateConversation(id, {
      title,
      model: summary?.model ?? null,
      group_name: summary?.group_name ?? null,
    })
    const idx = conversations.value.findIndex((c) => c.id === id)
    if (idx >= 0) conversations.value[idx] = { ...conversations.value[idx], title }
  }

  // ==================== 保存 ====================

  /** 防抖调度保存（流式回复结束后由调用方触发） */
  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => {
      saveTimer = null
      saveChain = saveChain.then(() => saveNow()).catch(() => {})
    }, CONVERSATION_SAVE_DEBOUNCE_MS)
  }

  /** 立即保存（切换/新建/卸载前调用），等待保存链完成 */
  async function flushSave() {
    if (saveTimer) {
      clearTimeout(saveTimer)
      saveTimer = null
      saveChain = saveChain.then(() => saveNow()).catch(() => {})
    }
    await saveChain
  }

  /** 执行一次保存：草稿态 → 创建会话；已有会话 → 全量 PUT */
  async function saveNow() {
    const messages = options.getMessages()
    isSaving.value = true
    try {
      if (activeConversationId.value === null) {
        // 草稿态：无消息不落库
        if (messages.length === 0) return
        const title = deriveTitle(messages) ?? options.defaultTitle()
        const created = await playgroundAPI.createConversation({
          title,
          model: options.getModel() || null,
          group_name: options.getGroupName() || null,
          messages,
        })
        activeConversationId.value = created.id
        conversations.value.unshift({
          id: created.id,
          title: created.title,
          model: created.model,
          group_name: created.group_name,
          last_activity_at: created.last_activity_at,
          created_at: created.created_at,
        })
        return
      }

      const id = activeConversationId.value
      const summary = conversations.value.find((c) => c.id === id)
      // 标题仍为默认值且已有用户消息时，自动补一次标题
      const autoTitle =
        summary && summary.title === options.defaultTitle()
          ? (deriveTitle(messages) ?? undefined)
          : undefined
      await playgroundAPI.updateConversation(id, {
        ...(autoTitle ? { title: autoTitle } : {}),
        model: options.getModel() || null,
        group_name: options.getGroupName() || null,
        messages,
      })
      touchSummary(id, autoTitle ? { title: autoTitle } : undefined)
    } finally {
      isSaving.value = false
    }
  }

  // ==================== 旧数据迁移 ====================

  /**
   * 旧版单会话 localStorage 数据一次性迁移：
   * 存在非空消息记录 → 创建一个会话导入，成功后删除旧 key。
   */
  async function migrateLegacyMessages() {
    let legacy: Message[] | null = null
    try {
      const raw = localStorage.getItem(STORAGE_KEYS.MESSAGES)
      if (raw) {
        const parsed = JSON.parse(raw) as Message[]
        if (Array.isArray(parsed) && parsed.length > 0) legacy = parsed
      }
    } catch {
      legacy = null
    }
    if (!legacy) {
      // 空数组/损坏数据也清掉，避免每次进页面重复解析
      localStorage.removeItem(STORAGE_KEYS.MESSAGES)
      return
    }
    try {
      await playgroundAPI.createConversation({
        title: deriveTitle(legacy) ?? options.defaultTitle(),
        model: options.getModel() || null,
        group_name: options.getGroupName() || null,
        messages: legacy,
      })
      localStorage.removeItem(STORAGE_KEYS.MESSAGES)
    } catch {
      // 迁移失败保留旧数据，下次进入页面重试
    }
  }

  // ==================== 生命周期 ====================

  /** beforeunload 时尽力同步保存（异步请求不保证送达，仅降低丢失概率） */
  function onBeforeUnload() {
    if (saveTimer) {
      clearTimeout(saveTimer)
      saveTimer = null
      void saveNow()
    }
  }

  onMounted(() => {
    window.addEventListener('beforeunload', onBeforeUnload)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('beforeunload', onBeforeUnload)
    // 组件卸载（路由跳转）时 flush 未保存内容
    void flushSave()
  })

  return {
    // State
    conversations,
    activeConversationId,
    isLoadingList,
    isLoadingDetail,
    isSaving,
    // Actions
    loadConversations,
    selectConversation,
    newConversation,
    removeConversation,
    renameConversation,
    scheduleSave,
    flushSave,
  }
}
