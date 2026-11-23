<template>
  <div ref="pickerRef" class="theme-color-picker">
    <button
      type="button"
      class="theme-color-trigger"
      :aria-label="t('settings.colorPicker.hex')"
      :aria-expanded="open"
      @click="togglePicker"
    >
      <span class="theme-color-swatch" :style="{ backgroundColor: color }" />
    </button>

    <div v-if="open" class="theme-color-popover" :class="{ flip }" :style="popoverStyle">
      <div class="theme-color-tabs" role="tablist">
        <button
          v-for="item in tabs"
          :key="item.value"
          type="button"
          role="tab"
          :aria-label="t(item.label)"
          :aria-selected="tab === item.value"
          :class="{ active: tab === item.value }"
          @click="tab = item.value"
        >
          <svg v-if="item.value === 'spectrum'" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M4 3v4l7 12v2h2v-2l7-12V3M4 7h16M8 7l4 5 4-5" />
          </svg>
          <svg v-else-if="item.value === 'palette'" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 3a9 9 0 1 0 0 18h1.5a2 2 0 0 0 1.4-3.4 1.6 1.6 0 0 1 1.1-2.8H18a3 3 0 0 0 3-3A9 9 0 0 0 12 3Z" />
            <path d="M7.5 11h.01M10 7.5h.01M15 7.5h.01M17 11h.01" />
          </svg>
          <svg v-else viewBox="0 0 24 24" aria-hidden="true">
            <path d="M4 6h16M4 12h16M4 18h16" />
            <circle cx="9" cy="6" r="2" />
            <circle cx="15" cy="12" r="2" />
            <circle cx="8" cy="18" r="2" />
          </svg>
        </button>
      </div>

      <div class="picker-content" :class="`content-${tab}`">
        <div v-if="tab === 'spectrum'" class="spectrum-row" :style="{ '--picker-hue-color': hueColor }">
          <input
            v-model.number="hsv.v"
            class="vertical-range brightness-range"
            type="range"
            min="0"
            max="100"
            :aria-label="t('settings.colorPicker.brightness')"
            @input="applyHsv"
          />
          <div
            class="color-spectrum"
            role="slider"
            tabindex="0"
            :aria-label="t('settings.colorPicker.saturation')"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-valuenow="Math.round(hsv.s)"
            :style="spectrumStyle"
            @pointerdown="startSpectrumChange"
            @pointermove="continueSpectrumChange"
            @keydown="handleSpectrumKey"
          />
          <input
            v-model.number="hsv.h"
            class="vertical-range hue-range"
            type="range"
            min="0"
            max="359"
            :aria-label="t('settings.colorPicker.hue')"
            @input="applyHsv"
          />
        </div>

        <div v-else-if="tab === 'palette'" class="color-palette" :aria-label="t('settings.colorPicker.palette')">
          <button
            v-for="preset in presets"
            :key="preset"
            type="button"
            :aria-label="preset"
            :aria-pressed="color === preset"
            :style="{ backgroundColor: preset }"
            @click="choose(preset)"
          />
        </div>

        <div v-else class="color-values">
          <div class="values-toolbar">
            <div class="value-mode" role="group" :aria-label="t('settings.colorPicker.values')">
              <button type="button" :class="{ active: valueMode === 'rgb' }" @click="valueMode = 'rgb'">RGB</button>
              <button type="button" :class="{ active: valueMode === 'hsv' }" @click="valueMode = 'hsv'">HSV</button>
            </div>
            <label class="hex-field">
              <span>#</span>
              <input
                v-model="hexDraft"
                :aria-label="t('settings.colorPicker.hex')"
                maxlength="6"
                spellcheck="false"
                @input="handleHexInput"
                @blur="applyHexDraft"
                @keydown.enter.prevent="applyHexDraft"
              />
            </label>
          </div>

          <div v-if="valueMode === 'rgb'" class="channel-list">
            <div v-for="channel in rgbChannels" :key="channel.key" class="channel-row">
              <label class="channel-field">
                <span>{{ t(channel.label) }}</span>
                <input
                  type="number"
                  min="0"
                  max="255"
                  :value="rgb[channel.index]"
                  @input="updateRgb(channel.index, $event)"
                />
              </label>
              <input
                class="channel-range"
                type="range"
                min="0"
                max="255"
                :value="rgb[channel.index]"
                :style="{ '--channel-start': '#000000', '--channel-end': channel.end }"
                :aria-label="t(channel.label)"
                @input="updateRgb(channel.index, $event)"
              />
            </div>
          </div>
          <div v-else class="channel-list">
            <div v-for="channel in hsvChannels" :key="channel.key" class="channel-row">
              <label class="channel-field">
                <span>{{ t(channel.label) }}</span>
                <input
                  type="number"
                  min="0"
                  :max="channel.max"
                  :value="Math.round(hsv[channel.key])"
                  @input="updateHsv(channel.key, $event)"
                />
              </label>
              <input
                class="channel-range"
                type="range"
                min="0"
                :max="channel.max"
                :value="Math.round(hsv[channel.key])"
                :style="{ background: channel.background }"
                :aria-label="t(channel.label)"
                @input="updateHsv(channel.key, $event)"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="shade-strip" :aria-label="t('settings.colorPicker.values')">
        <button type="button" :aria-label="t('settings.colorPicker.darkest')" :style="{ backgroundColor: shades.darkest }" @click="adjustShade(-24)" />
        <button type="button" :aria-label="t('settings.colorPicker.darker')" :style="{ backgroundColor: shades.darker }" @click="adjustShade(-12)" />
        <span class="current-color" :style="{ backgroundColor: color }" />
        <button type="button" :aria-label="t('settings.colorPicker.lighter')" :style="{ backgroundColor: shades.lighter }" @click="adjustShade(12)" />
        <button type="button" :aria-label="t('settings.colorPicker.lightest')" :style="{ backgroundColor: shades.lightest }" @click="adjustShade(24)" />
      </div>

      <button v-if="canPickFromScreen" type="button" class="screen-picker" :disabled="picking" @click="pickFromScreen">
        {{ t('settings.colorPicker.screenPicker') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

type HSV = { h: number; s: number; v: number }
type HSVChannel = 'h' | 's' | 'v'
const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ (event: 'update:modelValue', value: string): void }>()
const { t } = useI18n()
const presets = [
  '#FF434B', '#E5475C', '#087FD7', '#029BB7', '#7E7978', '#767676',
  '#D5333A', '#E90B22', '#0361AB', '#347E99', '#615F5B', '#494949',
  '#ED6654', '#E90050', '#9694D6', '#00B6C0', '#697687', '#68777A',
  '#DF3600', '#C80050', '#6868D1', '#07888E', '#525C69', '#485257',
  '#BF4C13', '#D6008D', '#8864BB', '#00AD8F', '#557A74', '#64816B',
  '#F65F12', '#D000A5', '#8553B2', '#008874', '#49645E', '#ECECEC',
  '#FF9200', '#ED009A', '#A600B3', '#00CF72', '#3D7900', '#141414',
  '#FFBA00', '#980098', '#740080', '#11803C', '#087800', '#000000',
]
const tabs = [
  { value: 'spectrum' as const, label: 'settings.colorPicker.spectrum' },
  { value: 'palette' as const, label: 'settings.colorPicker.palette' },
  { value: 'values' as const, label: 'settings.colorPicker.values' },
]
const rgbChannels = [
  { key: 'red', label: 'settings.colorPicker.red', index: 0, end: '#FF0000' },
  { key: 'green', label: 'settings.colorPicker.green', index: 1, end: '#00FF00' },
  { key: 'blue', label: 'settings.colorPicker.blue', index: 2, end: '#0000FF' },
]
const pickerRef = ref<HTMLElement | null>(null)
const color = ref(normalizeHex(props.modelValue))
const open = ref(false)
const flip = ref(false)
const popoverStyle = ref<Record<string, string>>({})
const picking = ref(false)
const tab = ref<'spectrum' | 'palette' | 'values'>('spectrum')
const valueMode = ref<'rgb' | 'hsv'>('rgb')
const hexDraft = ref(color.value.slice(1))
const hsv = reactive(hexToHsv(color.value))
const rgb = computed(() => hexToRgb(color.value))
const hsvChannels = computed(() => [
  { key: 'h' as const, label: 'settings.colorPicker.hue', max: 359, background: 'linear-gradient(to right,#f00,#ff0,#0f0,#0ff,#00f,#f0f,#f00)' },
  { key: 's' as const, label: 'settings.colorPicker.saturation', max: 100, background: `linear-gradient(to right,#888,${hueColor.value})` },
  { key: 'v' as const, label: 'settings.colorPicker.brightness', max: 100, background: `linear-gradient(to right,#000,${hueColor.value})` },
])
const hueColor = computed(() => `hsl(${hsv.h} 100% 50%)`)
const spectrumStyle = computed(() => ({
  backgroundImage: 'linear-gradient(to bottom,transparent,#fff),linear-gradient(to right,#f00,#ff0,#0f0,#0ff,#00f,#f0f,#f00)',
  '--spectrum-x': `${hsv.h / 359 * 100}%`,
  '--spectrum-y': `${100 - hsv.s}%`,
}))
const shades = computed(() => ({
  darkest: hsvToHex({ ...hsv, v: Math.max(0, hsv.v - 24) }),
  darker: hsvToHex({ ...hsv, v: Math.max(0, hsv.v - 12) }),
  lighter: hsvToHex({ ...hsv, v: Math.min(100, hsv.v + 12) }),
  lightest: hsvToHex({ ...hsv, v: Math.min(100, hsv.v + 24) }),
}))
const canPickFromScreen = computed(() => typeof window !== 'undefined' && Boolean((window as Window & { EyeDropper?: unknown }).EyeDropper))

watch(() => props.modelValue, syncColor)
watch(open, isOpen => {
  if (isOpen) {
    document.addEventListener('pointerdown', dismissOutside)
    document.addEventListener('keydown', closeOnEscape)
    window.addEventListener('resize', positionPopover)
    nextTick(positionPopover)
  } else {
    document.removeEventListener('pointerdown', dismissOutside)
    document.removeEventListener('keydown', closeOnEscape)
    window.removeEventListener('resize', positionPopover)
  }
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', dismissOutside)
  document.removeEventListener('keydown', closeOnEscape)
  window.removeEventListener('resize', positionPopover)
})

function normalizeHex(value: string) {
  return /^#[\da-f]{6}$/i.test(value) ? value.toUpperCase() : '#409EFF'
}
function hexToRgb(hex: string): [number, number, number] {
  return [1, 3, 5].map(index => parseInt(hex.slice(index, index + 2), 16)) as [number, number, number]
}
function rgbToHex(values: number[]) {
  return `#${values.map(value => Math.max(0, Math.min(255, Math.round(value))).toString(16).padStart(2, '0')).join('')}`.toUpperCase()
}
function hexToHsv(hex: string): HSV {
  const [red, green, blue] = hexToRgb(hex).map(value => value / 255)
  const max = Math.max(red, green, blue)
  const min = Math.min(red, green, blue)
  const delta = max - min
  let hue = 0
  if (delta) {
    if (max === red) hue = 60 * (((green - blue) / delta) % 6)
    else if (max === green) hue = 60 * ((blue - red) / delta + 2)
    else hue = 60 * ((red - green) / delta + 4)
  }
  return { h: (hue + 360) % 360, s: max ? delta / max * 100 : 0, v: max * 100 }
}
function hsvToHex(value: HSV) {
  const saturation = value.s / 100
  const brightness = value.v / 100
  const chroma = brightness * saturation
  const x = chroma * (1 - Math.abs((value.h / 60 % 2) - 1))
  const m = brightness - chroma
  const channels = value.h < 60 ? [chroma, x, 0] : value.h < 120 ? [x, chroma, 0] : value.h < 180 ? [0, chroma, x] : value.h < 240 ? [0, x, chroma] : value.h < 300 ? [x, 0, chroma] : [chroma, 0, x]
  return rgbToHex(channels.map(channel => (channel + m) * 255))
}
function syncColor(value: string) {
  color.value = normalizeHex(value)
  hexDraft.value = color.value.slice(1)
  Object.assign(hsv, hexToHsv(color.value))
}
function choose(value: string) {
  syncColor(value)
  emit('update:modelValue', color.value)
}
function applyHsv() {
  hsv.h = Math.max(0, Math.min(359, Number(hsv.h) || 0))
  hsv.s = Math.max(0, Math.min(100, Number(hsv.s) || 0))
  hsv.v = Math.max(0, Math.min(100, Number(hsv.v) || 0))
  choose(hsvToHex(hsv))
}
function updateRgb(index: number, event: Event) {
  const values = [...rgb.value]
  values[index] = Number((event.target as HTMLInputElement).value)
  choose(rgbToHex(values))
}
function updateHsv(key: HSVChannel, event: Event) {
  hsv[key] = Number((event.target as HTMLInputElement).value)
  applyHsv()
}
function handleHexInput() {
  hexDraft.value = hexDraft.value.replace(/[^\da-f]/gi, '').slice(0, 6).toUpperCase()
  if (hexDraft.value.length === 6) choose(`#${hexDraft.value}`)
}
function applyHexDraft() {
  if (/^[\da-f]{6}$/i.test(hexDraft.value)) choose(`#${hexDraft.value}`)
  else hexDraft.value = color.value.slice(1)
}
function applySpectrumPoint(event: PointerEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  hsv.h = Math.max(0, Math.min(359, ((event.clientX - rect.left) / rect.width) * 359))
  hsv.s = Math.max(0, Math.min(100, (1 - (event.clientY - rect.top) / rect.height) * 100))
  applyHsv()
}
function startSpectrumChange(event: PointerEvent) {
  ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
  applySpectrumPoint(event)
}
function continueSpectrumChange(event: PointerEvent) {
  if (event.buttons === 1) applySpectrumPoint(event)
}
function handleSpectrumKey(event: KeyboardEvent) {
  const amount = event.shiftKey ? 10 : 2
  if (event.key === 'ArrowLeft') hsv.h -= amount
  else if (event.key === 'ArrowRight') hsv.h += amount
  else if (event.key === 'ArrowUp') hsv.s += amount
  else if (event.key === 'ArrowDown') hsv.s -= amount
  else return
  event.preventDefault()
  applyHsv()
}
function adjustShade(delta: number) {
  hsv.v = Math.max(0, Math.min(100, hsv.v + delta))
  applyHsv()
}
function togglePicker(event: MouseEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  flip.value = window.innerHeight - rect.bottom < 452
  open.value = !open.value
}
function positionPopover() {
  const root = pickerRef.value
  const trigger = root?.querySelector('.theme-color-trigger')
  if (!root || !trigger) return
  const rootRect = root.getBoundingClientRect()
  const triggerRect = trigger.getBoundingClientRect()
  flip.value = window.innerHeight - triggerRect.bottom < 452
  const width = Math.min(360, window.innerWidth - 32)
  const sidebarRect = document.querySelector('.sidebar-container')?.getBoundingClientRect()
  const minLeft = Math.max(8, sidebarRect && sidebarRect.right > 0 ? sidebarRect.right + 12 : 8)
  let left = Math.max(minLeft, triggerRect.right - width)
  left = Math.min(left, window.innerWidth - width - 8)
  if (left < minLeft) left = minLeft
  popoverStyle.value = { left: `${left - rootRect.left}px`, right: 'auto' }
}
function dismissOutside(event: PointerEvent) {
  if (!pickerRef.value?.contains(event.target as Node)) open.value = false
}
function closeOnEscape(event: KeyboardEvent) {
  if (event.key === 'Escape') open.value = false
}
async function pickFromScreen() {
  const EyeDropper = (window as Window & { EyeDropper?: new () => { open: () => Promise<{ sRGBHex: string }> } }).EyeDropper
  if (!EyeDropper || picking.value) return
  picking.value = true
  try {
    const result = await new EyeDropper().open()
    choose(result.sRGBHex)
  } catch {
    // Keep the current color if screen color picking is cancelled.
  } finally {
    picking.value = false
  }
}
</script>

<style lang="scss" scoped>
.theme-color-picker { position: relative; z-index: 5; display: flex; align-items: center; justify-content: center; }
.theme-color-trigger { display: grid; width: 40px; height: 40px; place-items: center; padding: 0; border: 0; border-radius: 10px; background: transparent; cursor: pointer; }
.theme-color-trigger:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.theme-color-swatch { width: 36px; height: 36px; border: 1px solid var(--el-border-color-light); border-radius: 9px; box-shadow: inset 0 0 0 1px rgb(0 0 0 / 7%); transition: transform 140ms ease, box-shadow 140ms ease; }
.theme-color-trigger:hover .theme-color-swatch { transform: scale(1.04); box-shadow: 0 2px 9px rgb(0 0 0 / 20%); }
.theme-color-popover { position: absolute; z-index: 40; top: calc(100% + 8px); right: 0; display: flex; width: min(360px, calc(100vw - 32px)); box-sizing: border-box; flex-direction: column; padding: 10px; border: 1px solid var(--el-border-color-light); border-radius: 12px; background: var(--el-bg-color-overlay); color: var(--el-text-color-primary); box-shadow: 0 14px 36px rgb(8 13 28 / 28%); }
.theme-color-popover.flip { top: auto; bottom: calc(100% + 8px); }
.theme-color-tabs { display: grid; height: 34px; flex: none; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px; margin-bottom: 12px; padding: 3px; box-sizing: border-box; border-radius: 9px; background: var(--el-fill-color-light); }
.theme-color-tabs button, .value-mode button { display: grid; min-width: 0; place-items: center; padding: 0; border: 0; border-radius: 7px; background: transparent; color: var(--el-text-color-secondary); cursor: pointer; }
.theme-color-tabs button svg { width: 22px; height: 22px; fill: none; stroke: currentColor; stroke-width: 1.5; stroke-linecap: round; stroke-linejoin: round; }
.theme-color-tabs button.active, .value-mode button.active { background: var(--el-color-primary); color: #fff; }
.picker-content { display: flex; height: 278px; min-height: 278px; flex: none; align-items: center; justify-content: center; }
.spectrum-row { display: grid; width: 328px; max-width: 100%; height: 262px; grid-template-columns: 25px minmax(0, 262px) 25px; gap: 8px; align-items: stretch; justify-content: center; }
.color-spectrum { position: relative; width: 100%; height: 262px; min-width: 0; box-sizing: border-box; overflow: hidden; border-radius: 7px; touch-action: none; cursor: crosshair; }
.color-spectrum::after { position: absolute; top: var(--spectrum-y); left: var(--spectrum-x); width: 16px; height: 16px; content: ''; transform: translate(-50%, -50%); border: 2px solid #fff; border-radius: 50%; box-shadow: 0 1px 4px rgb(0 0 0 / 65%); pointer-events: none; }
.color-spectrum:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.vertical-range { width: 25px; height: 262px; margin: 0; appearance: none; -webkit-appearance: none; writing-mode: vertical-lr; direction: rtl; border: 0; border-radius: 8px; touch-action: none; }
.brightness-range { background: linear-gradient(to top, #000, var(--picker-hue-color)); }
.hue-range { background: linear-gradient(to bottom, #f00, #f0f, #00f, #0ff, #0f0, #ff0, #f00); }
.vertical-range::-webkit-slider-runnable-track { width: 25px; border-radius: 8px; background: transparent; }
.vertical-range::-webkit-slider-thumb { width: 20px; height: 20px; margin-left: 2px; appearance: none; -webkit-appearance: none; border: 2px solid #fff; border-radius: 50%; background: transparent; box-shadow: 0 1px 4px rgb(0 0 0 / 65%); }
.vertical-range::-moz-range-track { background: transparent; }
.vertical-range::-moz-range-thumb { width: 16px; height: 16px; border: 2px solid #fff; border-radius: 50%; background: transparent; box-shadow: 0 1px 4px rgb(0 0 0 / 65%); }
.color-palette { display: grid; width: 328px; max-width: 100%; height: 262px; grid-template-columns: repeat(6, minmax(0, 1fr)); grid-template-rows: repeat(8, minmax(0, 1fr)); align-content: stretch; margin-inline: auto; overflow: hidden; border-radius: 7px; }
.color-palette button { z-index: 0; width: 100%; height: 100%; padding: 0; border: 0; border-radius: 0; cursor: pointer; }
.color-palette button[aria-pressed='true'] { z-index: 1; outline: 2px solid #fff; outline-offset: -3px; box-shadow: inset 0 0 0 1px rgb(0 0 0 / 50%); }
.color-values { display: flex; width: 328px; max-width: 100%; height: 100%; flex-direction: column; justify-content: flex-start; gap: 20px; margin-inline: auto; }
.values-toolbar { display: grid; height: 36px; flex: none; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 10px; }
.value-mode { display: grid; height: 36px; grid-template-columns: 1fr 1fr; gap: 3px; padding: 3px; box-sizing: border-box; border-radius: 8px; background: var(--el-fill-color-light); }
.value-mode button { padding: 0 8px; font-size: 11px; }
.hex-field { display: flex; min-width: 0; height: 36px; align-items: center; justify-content: flex-end; gap: 5px; color: var(--el-text-color-secondary); font-size: 12px; }
.hex-field input { width: 112px; min-width: 0; height: 36px; box-sizing: border-box; padding: 5px 8px; border: 1px solid var(--el-border-color); border-radius: 7px; outline: none; background: var(--el-bg-color); color: var(--el-text-color-primary); font: 12px Consolas, monospace; }
.hex-field input:focus { border-color: var(--el-color-primary); }
.channel-list { display: flex; flex-direction: column; gap: 22px; }
.channel-row { display: grid; height: 36px; grid-template-columns: 124px minmax(0, 1fr); gap: 10px; align-items: center; }
.channel-field { display: grid; min-width: 0; height: 36px; grid-template-columns: 44px minmax(0, 1fr); align-items: center; }
.channel-field span { overflow: hidden; color: var(--el-text-color-secondary); font-size: 12px; line-height: 1; text-align: left; text-overflow: ellipsis; white-space: nowrap; }
.channel-field input { width: 100%; min-width: 0; height: 36px; box-sizing: border-box; padding: 4px 5px; border: 1px solid var(--el-border-color); border-radius: 7px; outline: none; background: var(--el-bg-color); color: var(--el-text-color-primary); font-size: 12px; font-variant-numeric: tabular-nums; text-align: center; appearance: textfield; -moz-appearance: textfield; }
.channel-field input:focus { border-color: var(--el-color-primary); }
.channel-field input::-webkit-inner-spin-button, .channel-field input::-webkit-outer-spin-button { display: none; margin: 0; appearance: none; -webkit-appearance: none; }
.channel-range { width: 100%; height: 20px; margin: 0; padding: 0; appearance: none; -webkit-appearance: none; border: 0; border-radius: 10px; background: linear-gradient(to right, var(--channel-start), var(--channel-end)); touch-action: none; }
.channel-range::-webkit-slider-runnable-track { height: 20px; border-radius: 10px; background: transparent; }
.channel-range::-webkit-slider-thumb { width: 20px; height: 20px; box-sizing: border-box; appearance: none; -webkit-appearance: none; border: 2px solid #fff; border-radius: 50%; background: transparent; box-shadow: 0 1px 3px rgb(0 0 0 / 50%); }
.channel-range::-moz-range-track { height: 20px; border-radius: 10px; background: transparent; }
.channel-range::-moz-range-thumb { width: 16px; height: 16px; box-sizing: border-box; border: 2px solid #fff; border-radius: 50%; background: transparent; box-shadow: 0 1px 3px rgb(0 0 0 / 50%); }
.shade-strip { display: grid; height: 28px; flex: none; grid-template-columns: 1fr 1fr 4fr 1fr 1fr; gap: 0; align-items: center; margin: 8px 4px 22px; overflow: visible; border-radius: 4px; }
.shade-strip button { width: 100%; height: 28px; padding: 0; border: 0; border-radius: 0; cursor: pointer; }
.shade-strip .current-color { position: relative; z-index: 2; width: 100%; height: 46px; box-sizing: border-box; border: 0; border-radius: 3px; box-shadow: 0 3px 8px rgb(8 13 28 / 53%); pointer-events: none; }
.screen-picker { min-height: 34px; flex: none; align-self: center; padding: 0 16px; border: 1px solid var(--el-border-color); border-radius: 7px; background: var(--el-fill-color-light); color: var(--el-text-color-regular); font: inherit; font-size: 12px; cursor: pointer; }
.screen-picker:hover { border-color: var(--el-color-primary-light-5); color: var(--el-color-primary); }
.screen-picker:disabled { opacity: .55; cursor: not-allowed; }
@media (max-width: 420px) {
  .theme-color-popover { right: -8px; width: min(360px, calc(100vw - 24px)); }
  .spectrum-row { grid-template-columns: 22px minmax(0, 1fr) 22px; gap: 7px; }
  .vertical-range { width: 22px; }
  .channel-row { grid-template-columns: 118px minmax(0, 1fr); gap: 7px; }
  .channel-field { grid-template-columns: 40px minmax(0, 1fr); }
}
</style>
