<template>
  <span class="sidebar-menu-icon">
    <el-icon v-if="icon && icon.startsWith('el-icon')" class="sub-el-icon">
      <component :is="icon.replace('el-icon-', '')" />
    </el-icon>
    <svg-icon v-else-if="icon" :icon-class="icon" size="18px" />
    <svg-icon v-else icon-class="menu" size="18px" />
  </span>
  <span v-if="title" class="sidebar-menu-label">{{ translateRouteTitle(title) }}</span>
</template>

<script setup lang="ts">
import { translateRouteTitle } from "@/utils/i18n";

defineProps({
  icon: {
    type: String,
    default: "",
  },
  title: {
    type: String,
    default: "",
  },
});
</script>

<style lang="scss" scoped>
.sub-el-icon {
  width: 18px !important;
  height: 18px;
  margin-right: 0 !important;
  font-size: 18px !important;
  color: currentcolor;
}

.sidebar-menu-icon {
  display: inline-flex;
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  align-items: center;
  justify-content: center;
}

.sidebar-menu-label {
  display: inline-block;
  max-width: 180px;
  overflow: hidden;
  margin-left: 35px;
  opacity: 1;
  white-space: nowrap;
  transition:
    max-width 360ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 300ms ease-in-out;
}

</style>
