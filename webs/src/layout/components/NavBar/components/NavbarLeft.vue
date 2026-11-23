<template>
  <div class="navbar-left">
    <hamburger
      :is-active="appStore.sidebar.opened"
      @toggle-click="toggleSideBar"
    />
    <breadcrumb />
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from "@/store";

const appStore = useAppStore();

function toggleSideBar() {
  appStore.toggleSidebar();
}
</script>

<style scoped>
.navbar-left {
	min-width: 0;
	display: flex;
	align-items: center;
	gap: 4px;
}
</style>
