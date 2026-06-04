<template>
  <div class="param-row">
    <div class="param-header">
      <label class="param-label">{{ label }}</label>
      <div class="param-controls">
        <input
          v-if="enabled"
          type="number"
          :value="value"
          :min="min"
          :max="max"
          :step="step"
          class="param-number-input"
          @input="onNumberInput"
        />
        <button
          type="button"
          class="param-toggle"
          :class="{ 'param-toggle-on': enabled }"
          @click="$emit('update:enabled', !enabled)"
        >
          <span class="param-toggle-knob" />
        </button>
      </div>
    </div>
    <input
      v-if="enabled"
      type="range"
      :value="value"
      :min="min"
      :max="max"
      :step="step"
      class="param-slider"
      @input="onSliderInput"
    />
  </div>
</template>

<script setup lang="ts">
interface Props {
  label: string
  value: number
  enabled: boolean
  min: number
  max: number
  step: number
  integer?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  integer: false,
})

const emit = defineEmits<{
  (e: 'update:value', v: number): void
  (e: 'update:enabled', v: boolean): void
}>()

function sanitize(v: string): number {
  const n = props.integer ? parseInt(v, 10) : parseFloat(v)
  if (Number.isNaN(n)) return props.min
  return Math.max(props.min, Math.min(props.max, n))
}

function onSliderInput(e: Event) {
  emit('update:value', sanitize((e.target as HTMLInputElement).value))
}
function onNumberInput(e: Event) {
  emit('update:value', sanitize((e.target as HTMLInputElement).value))
}
</script>

<style scoped>
.param-row {
  @apply space-y-1.5;
}
.param-header {
  @apply flex items-center justify-between gap-3;
}
.param-label {
  @apply text-sm font-medium text-gray-700 dark:text-gray-200;
}
.param-controls {
  @apply flex items-center gap-2;
}
.param-number-input {
  @apply h-7 w-20 rounded-md border border-gray-200 bg-white px-2 text-right text-xs tabular-nums focus:border-primary-400 focus:outline-none focus:ring-1 focus:ring-primary-400;
  @apply dark:border-dark-600 dark:bg-dark-700 dark:text-gray-100;
}
.param-slider {
  @apply h-1.5 w-full appearance-none rounded-full bg-gray-200 accent-primary-600;
  @apply dark:bg-dark-600;
}
.param-toggle {
  @apply relative inline-flex h-4 w-7 flex-shrink-0 cursor-pointer rounded-full bg-gray-300 transition-colors;
  @apply dark:bg-dark-600;
}
.param-toggle-on {
  @apply bg-primary-600;
}
.param-toggle-knob {
  @apply pointer-events-none inline-block h-3 w-3 translate-x-0.5 translate-y-0.5 transform rounded-full bg-white shadow ring-0 transition;
}
.param-toggle-on .param-toggle-knob {
  @apply translate-x-3.5;
}
</style>
