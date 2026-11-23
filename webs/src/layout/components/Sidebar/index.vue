<template>
  <div :class="{ 'has-logo': sidebarLogo, 'sidebar-left-layout': layout === 'left' }">
    <div v-if="layout === 'left'" class="sidebar-left-content">
      <el-scrollbar class="sidebar-menu-scroll">
        <SidebarMenu :menu-list="flatMenuRoutes" base-path="" />
      </el-scrollbar>
      <div class="sidebar-brand-footer">
        <NavbarRight placement="sidebar" />
        <SidebarLogo v-if="sidebarLogo" :collapse="!appStore.sidebar.opened" />
        <span v-if="sidebarLogo && appStore.sidebar.opened" class="sidebar-copyright">© 2026 SubLinkX</span>
      </div>
    </div>
    <!-- 混合布局 -->
    <div class="flex w-full" v-else-if="layout == 'mix'">
      <SidebarLogo v-if="sidebarLogo" :collapse="!appStore.sidebar.opened" />
      <SidebarMixTopMenu class="flex-1" />
      <NavbarRight />
    </div>
    <!-- 顶部布局 -->
    <template v-else>
      <SidebarLogo v-if="sidebarLogo" :collapse="!appStore.sidebar.opened" />
      <el-scrollbar>
        <SidebarMenu :menu-list="permissionStore.routes" base-path="" />
      </el-scrollbar>
      <NavbarRight v-if="layout === 'top'" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore, usePermissionStore, useAppStore } from "@/store";
import path from "path-browserify";

const appStore = useAppStore();
const settingsStore = useSettingsStore();
const permissionStore = usePermissionStore();

const sidebarLogo = computed(() => settingsStore.sidebarLogo);
const layout = computed(() => settingsStore.layout);

const flatMenuRoutes = computed(() => {
	function flatten(routes: any[], parentPath = "", inheritedMeta?: any): any[] {
		return routes.flatMap((route) => {
			if (route.meta?.hidden) return [];
			if (route.meta?.title === "userset") return [];
			const fullPath = path.resolve(parentPath, route.path || "");
			const meta = route.meta || inheritedMeta;
			const visibleChildren = (route.children || []).filter((child: any) => !child.meta?.hidden);
			if (visibleChildren.length) {
				return flatten(visibleChildren, fullPath, meta);
			}
			return [{ ...route, path: fullPath, meta, children: undefined }];
		});
	}

	return flatten(permissionStore.routes);
});
</script>

<style lang="scss" scoped>
.has-logo:not(.sidebar-left-layout) {
  .el-scrollbar {
    height: calc(100vh - $navbar-height);
  }
}

.sidebar-left-content {
	display: flex;
	height: 100%;
	flex-direction: column;
	box-sizing: border-box;
	padding: 8px;

	.sidebar-menu-scroll {
		min-height: 0;
		flex: 1;
		height: auto;
	}

	.sidebar-brand-footer {
		flex: 0 0 auto;
		padding: 10px 10px 12px;
		border-top: 1px solid var(--el-border-color-light);
	}

	:deep(.sidebar-menu-scroll > .el-scrollbar__wrap > .el-scrollbar__view) {
		min-height: 100%;
	}

	.sidebar-copyright {
		display: block;
		padding: 2px 0 0 32px;
		color: #99a2ad;
		font-size: 10px;
		white-space: nowrap;
	}
}
</style>
