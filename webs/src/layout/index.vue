<template>
  <div class="wh-full" :class="classObj">
    <!-- 遮罩层 -->
    <div
      v-if="classObj.mobile && classObj.openSidebar"
      class="wh-full fixed-lt z-999 bg-black bg-opacity-30"
      @click="handleOutsideClick"
    ></div>

    <!-- 公用侧边栏 -->
    <Sidebar class="sidebar-container" />

    <!-- 混合布局 -->
    <div v-if="layout === 'mix'" class="mix-container">
      <div class="mix-container__left">
        <SidebarMenu :menu-list="mixLeftMenus" :base-path="activeTopMenuPath" />
        <div class="sidebar-toggle">
          <hamburger
            :is-active="appStore.sidebar.opened"
            @toggle-click="toggleSidebar"
          />
        </div>
      </div>

      <div class="main-container">
        <AppMain />
      </div>
    </div>

    <!-- 左侧和顶部布局 -->
    <div v-else class="main-container">
      <AppMain />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAppStore, useSettingsStore, usePermissionStore } from "@/store";
import { DeviceEnum } from "@/enums/DeviceEnum";

const appStore = useAppStore();
const settingsStore = useSettingsStore();
const permissionStore = usePermissionStore();

const fixedHeader = computed(() => settingsStore.fixedHeader); // 是否固定header
const layout = computed(() => settingsStore.layout); // 布局模式 left top mix
const activeTopMenuPath = computed(() => appStore.activeTopMenuPath); // 顶部菜单激活path
const mixLeftMenus = computed(() => permissionStore.mixLeftMenus); // 混合布局左侧菜单

watch(
  () => activeTopMenuPath.value,
  (newVal) => {
    permissionStore.setMixLeftMenus(newVal);
  },
  {
    deep: true,
    immediate: true,
  }
);

const classObj = computed(() => ({
  hideSidebar: !appStore.sidebar.opened,
  openSidebar: appStore.sidebar.opened,
  mobile: appStore.device === DeviceEnum.MOBILE,
  "layout-left": layout.value === "left",
  "layout-top": layout.value === "top",
  "layout-mix": layout.value === "mix",
}));

const width = useWindowSize().width;
const WIDTH = 992; // 响应式布局容器固定宽度  大屏（>=1200px） 中屏（>=992px） 小屏（>=768px）

watchEffect(() => {
  if (width.value < WIDTH) {
    appStore.toggleDevice(DeviceEnum.MOBILE);
    appStore.closeSideBar();
  } else {
    appStore.toggleDevice(DeviceEnum.DESKTOP);

    if (width.value >= 1200) {
      appStore.openSideBar();
    } else {
      appStore.closeSideBar();
    }
  }
});

function handleOutsideClick() {
  appStore.closeSideBar();
}

function toggleSidebar() {
  appStore.toggleSidebar();
}
</script>

<style lang="scss" scoped>
.fixed-header {
  position: fixed;
  top: 0;
  right: 0;
  z-index: 9;
  width: calc(100% - $sidebar-width);
  transition: width 360ms cubic-bezier(0.22, 1, 0.36, 1);
}

.sidebar-container {
  position: fixed;
  top: 14px;
  bottom: 14px;
  left: 14px;
  z-index: 999;
  width: $sidebar-width;
  height: auto;
  overflow: hidden;
  background-color: $menu-background;
	border: 1px solid var(--el-border-color-light);
	border-radius: 22px;
	box-shadow: 0 8px 28px rgb(15 23 42 / 7%);
  transition: width 0.28s;

  :deep(.el-menu) {
    border: none;
		padding: 10px 9px;
		background: transparent;
		--el-menu-bg-color: transparent;
		--el-menu-text-color: var(--menu-text);
		--el-menu-hover-bg-color: var(--menu-hover);
		--el-menu-active-color: #409eff;
		--el-menu-item-height: 36px;
		border-radius: var(--app-radius);
	}

	html.dark & :deep(.el-menu) {
		--el-menu-text-color: #a8b3c5;
		--el-menu-hover-bg-color: rgb(51 65 85 / 55%);
		--el-menu-active-color: #93c5fd;
	}

	:deep(.el-menu-item),
	:deep(.el-sub-menu__title) {
		margin: 3px 0;
		border-radius: var(--app-radius);
		font-weight: 550;
		transition: color 160ms ease, background 160ms ease;
	}

	:deep(.el-menu-item.is-active) {
		background: var(--menu-active-background);
		color: var(--menu-active-text);
		font-weight: 650;
	}

	html.dark & :deep(.el-menu-item.is-active) {
		background: rgb(64 158 255 / 20%);
		color: #93c5fd;
	}

	:deep(.el-menu-item:hover),
	:deep(.el-sub-menu__title:hover) {
		background: var(--menu-hover);
  }

	html.dark & :deep(.el-menu-item:hover),
	html.dark & :deep(.el-sub-menu__title:hover) {
		background: rgb(51 65 85 / 55%);
	}
}

.layout-left .sidebar-container {
	:deep(.sidebar-menu-scroll .el-menu) {
		padding-right: 0;
		padding-left: 0;
	}

	:deep(.sidebar-menu-scroll .el-menu > a) {
		display: flex;
		width: 100%;
	}

	:deep(.sidebar-menu-scroll .el-menu-item) {
		padding-left: 0;
	}

	:deep(.sidebar-menu-scroll .el-menu .el-menu-item) {
		position: relative;
	}

	:deep(.sidebar-menu-scroll .el-menu .el-menu-item .sidebar-menu-icon) {
		position: absolute;
		left: 9px;
		top: 50%;
		transform: translateY(-50%);
	}
}

.layout-left:not(.hideSidebar) .sidebar-container {
	:deep(.sidebar-menu-label) {
		transition-delay: 80ms;
	}
}

.main-container {
  position: relative;
  min-height: 100%;
  box-sizing: border-box;
  min-width: 0;
  padding: 14px 12px;
  margin-left: calc($sidebar-width + 14px);
  transition: margin-left 360ms cubic-bezier(0.22, 1, 0.36, 1);
}

.layout-top {
  .fixed-header {
    top: $navbar-height;
    width: 100%;
  }

  .sidebar-container {
    z-index: 999;
    display: flex;
    width: 100% !important;
    height: $navbar-height;

    :deep(.el-scrollbar) {
      flex: 1;
      height: $navbar-height;
    }

    :deep(.el-menu-item),
    :deep(.el-sub-menu__title),
    :deep(.el-menu--horizontal) {
      height: $navbar-height;
      line-height: $navbar-height;
    }

    :deep(.el-menu--collapse) {
      width: 100%;
    }
  }

  .main-container {
    min-height: calc(100vh - $navbar-height);
    padding-top: $navbar-height;
    margin-left: 0;
  }
}

.layout-mix {
  .sidebar-container {
    width: 100% !important;
    height: $navbar-height;

    :deep(.el-scrollbar) {
      flex: 1;
      height: $navbar-height;
    }

    :deep(.el-menu-item),
    :deep(.el-sub-menu__title),
    :deep(.el-menu--horizontal) {
      height: $navbar-height;
      line-height: $navbar-height;
    }

    :deep(.el-menu--horizontal.el-menu) {
      border: none;
    }
  }

  .mix-container {
    display: flex;
    height: 100%;
    padding-top: $navbar-height;

    .mix-container__left {
      position: relative;
      width: $sidebar-width;
      height: 100%;

      :deep(.el-menu) {
        height: 100%;
        border: none;
      }

      .sidebar-toggle {
        position: absolute;
        bottom: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        height: 50px;
        line-height: 50px;
        box-shadow: 0 0 6px -2px var(--el-color-primary);

        div:hover {
          background-color: var(--menu-background);
        }

        :deep(svg) {
          color: var(--el-color-primary) !important;
        }
      }
    }

    .main-container {
      flex: 1;
      min-width: 0;
    padding: 14px;
      margin-left: 0;

      .fixed-header {
        top: $navbar-height;
      }
    }
  }
}

.hideSidebar {
  .fixed-header {
    left: $sidebar-width-collapsed;
    width: calc(100% - $sidebar-width-collapsed);
  }

  .main-container {
    margin-left: calc($sidebar-width-collapsed + 14px);
  }

	.sidebar-container {
		:deep(.sidebar-menu-scroll .el-menu) {
			box-sizing: border-box;
			width: 100%;
			padding-right: 0;
			padding-left: 0;
		}

		:deep(.sidebar-menu-scroll .el-menu > a) {
			position: relative;
			display: flex;
			width: 100%;
		}

		:deep(.sidebar-menu-scroll .el-menu .el-menu-item) {
			box-sizing: border-box;
			position: relative;
			width: 36px;
			height: 36px;
			min-height: 36px;
			margin-right: 0;
			margin-left: 0;
			padding-right: 0;
			padding-left: 0;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50% !important;
			line-height: 36px;
		}

		:deep(.sidebar-menu-scroll .el-scrollbar__wrap) {
			overflow-x: hidden;
		}

		:deep(.sidebar-menu-scroll .el-scrollbar__bar.is-horizontal) {
			display: none;
		}

		:deep(.sidebar-menu-scroll .el-menu .el-menu-item .sidebar-menu-icon) {
			position: absolute;
			left: 9px;
			top: 50%;
			transform: translateY(-50%);
		}

		:deep(.sidebar-menu-scroll .el-menu .el-menu-item .svg-icon),
		:deep(.sidebar-menu-scroll .el-menu .el-menu-item .sub-el-icon) {
			margin-right: 0 !important;
			margin-left: 0 !important;
		}

		:deep(.sidebar-menu-scroll .el-menu .el-menu-item .svg-icon) {
			flex: 0 0 18px;
			width: 18px !important;
			height: 18px !important;
		}

		:deep(.sidebar-menu-scroll .el-menu .el-menu-item .sub-el-icon) {
			width: 18px !important;
			font-size: 18px !important;
		}

		:deep(.sidebar-brand-footer) {
			padding-right: 0;
			padding-left: 0;
		}
	}

  &.layout-top {
    .fixed-header {
      left: 0;
      width: 100%;
    }

    .main-container {
      margin-left: 0;
    }
  }

  &.layout-mix {
    .fixed-header {
      left: $sidebar-width-collapsed;
      width: calc(100% - $sidebar-width-collapsed);
    }

    .sidebar-container {
      width: 100% !important;
    }

    .mix-container {
      .mix-container__left {
        width: $sidebar-width-collapsed;
      }
    }
  }
}

.layout-left.hideSidebar {
  .sidebar-container {
    width: $sidebar-width-collapsed !important;
  }

  .main-container {
    width: calc(100% - #{$sidebar-width-collapsed + 14px});
    margin-left: calc($sidebar-width-collapsed + 14px);
  }

  &.mobile {
    .sidebar-container {
      width: $sidebar-width-collapsed !important;
      pointer-events: auto;
      transform: none;
    }

    .main-container {
      width: calc(100% - #{$sidebar-width-collapsed + 10px});
      padding: 10px;
      margin-left: calc($sidebar-width-collapsed + 10px);
    }
  }
}

@media (max-width: 992px) {
	.sidebar-container {
		top: 10px;
		bottom: 10px;
		left: 10px;
		border-radius: 18px;
	}

	.main-container {
		padding: 10px;
	}
}

.mobile {
  .fixed-header {
    left: 0;
    width: 100%;
  }

  .main-container {
    width: 100%;
    padding: 14px;
    margin-left: 0;
  }

  &.layout-top {
    .sidebar-container {
      z-index: 999;
      display: flex;
      width: 100% !important;
      height: $navbar-height;

      :deep(.el-scrollbar) {
        flex: 1;
        min-width: 0;
        height: $navbar-height;
      }
    }

    .main-container {
      padding-top: $navbar-height;
      margin-left: 0;
      overflow: hidden;
    }

    // 顶部模式全局变量修改
    --el-menu-item-height: $navbar-height;
  }
}
</style>
