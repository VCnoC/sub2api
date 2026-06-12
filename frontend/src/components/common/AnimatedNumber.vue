<template>
  <span class="tabular-nums">{{ display }}</span>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'

/**
 * 数字滚动动画组件：值变化时从旧值平滑滚动到新值（ease-out cubic）。
 * 尊重系统 prefers-reduced-motion 设置，减少动效时直接显示目标值。
 */
const props = withDefaults(
  defineProps<{
    value: number
    /** 自定义格式化函数（如 formatTokens / toFixed），默认千分位整数 */
    format?: (n: number) => string
    /** 动画时长（毫秒） */
    duration?: number
  }>(),
  { duration: 800 }
)

const formatValue = (n: number): string =>
  props.format ? props.format(n) : Math.round(n).toLocaleString()

const display = ref(formatValue(0))
let raf = 0
let from = 0

function animate(to: number) {
  cancelAnimationFrame(raf)
  if (
    typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  ) {
    from = to
    display.value = formatValue(to)
    return
  }
  const start = performance.now()
  const startVal = from
  const tick = (now: number) => {
    const p = Math.min(1, (now - start) / props.duration)
    const eased = 1 - Math.pow(1 - p, 3)
    display.value = formatValue(p >= 1 ? to : startVal + (to - startVal) * eased)
    if (p < 1) {
      raf = requestAnimationFrame(tick)
    } else {
      from = to
    }
  }
  raf = requestAnimationFrame(tick)
}

watch(
  () => props.value,
  (v) => animate(Number.isFinite(v) ? v : 0),
  { immediate: true }
)

onBeforeUnmount(() => cancelAnimationFrame(raf))
</script>
