<template>
  <div :class="['navbar-right', { 'sidebar-navbar-right': placement === 'sidebar' }]">
    <!-- 用户头像 -->
    <el-dropdown
      class="setting-item sidebar-account"
      trigger="click"
      popper-class="account-dropdown-popper"
    >
      <div class="account-trigger">
        <img
          :src="avatarSrc"
          class="avatar-image"
          alt=""
          @error="handleAvatarError"
        />
        <span class="navbar-username">{{ userStore.user.username }}</span>
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="logout">
            {{ $t("navbar.logout") }}
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

  </div>
</template>
<script setup lang="ts">
import {
  useTagsViewStore,
  useUserStore,
} from "@/store";
import { useI18n } from "vue-i18n";

withDefaults(defineProps<{ placement?: "header" | "sidebar" }>(), {
	placement: "header",
});

const tagsViewStore = useTagsViewStore();
const userStore = useUserStore();
const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const fallbackAvatar = `${import.meta.env.BASE_URL}avatar.gif`;
const avatarSrc = computed(() => {
  const avatar = userStore.user.avatar?.trim();
  if (!avatar) return fallbackAvatar;
  if (/^(https?:|data:|blob:|\/\/)/i.test(avatar)) return avatar;

  const relativePath = avatar.replace(/^\/+/, "");
  if (import.meta.env.DEV) {
    const apiOrigin = import.meta.env.VITE_APP_API_URL?.replace(/\/+$/, "");
    return apiOrigin ? `${apiOrigin}/${relativePath}` : fallbackAvatar;
  }
  return `/${relativePath}`;
});

/**
 * 注销
 */
function logout() {
  ElMessageBox.confirm(t("navbar.logoutPrompt"), t("common.deletePrompt"), {
    confirmButtonText: t("common.confirm"),
    cancelButtonText: t("common.cancel"),
    type: "warning",
    lockScroll: false,
  }).then(() => {
    userStore
      .logout()
      .then(() => {
        tagsViewStore.delAllViews();
      })
      .then(() => {
        router.push(`/login?redirect=${route.fullPath}`);
      });
  });
}

function handleAvatarError(event: Event) {
  const image = event.currentTarget as HTMLImageElement;
  if (image.src !== new URL(fallbackAvatar, window.location.href).href) {
    image.src = fallbackAvatar;
  }
}
</script>
<style lang="scss" scoped>
.setting-item {
	display: inline-flex;
	min-width: 40px;
	height: 40px;
	align-items: center;
	justify-content: center;
	line-height: 1;
	margin: 0 2px;
	border-radius: var(--app-radius);
	color: #64748b;
  text-align: center;
  cursor: pointer;
	transition: color 160ms ease, background 160ms ease;

  &:hover {
		background: #eaf4ff;
		color: #2488ee;
  }
}

.navbar-right {
	display: flex;
	align-items: center;
	gap: 4px;
}

.avatar-image {
	width: 24px;
	height: 24px;
	flex: 0 0 24px;
	border-radius: 50%;
	object-fit: cover;
}

.setting-item .account-trigger {
	display: flex;
	width: 100%;
	height: 100%;
	align-items: center;
	justify-content: center;
	padding: 0 10px;
}

.sidebar-navbar-right {
	display: flex;
	flex-direction: column;
	align-items: stretch;
	gap: 3px;

	.setting-item {
		box-sizing: border-box;
		width: 100%;
		min-width: 0;
		height: 36px;
		flex: 0 0 36px;
		justify-content: flex-start;
		padding: 0 10px;
		margin: 0;
		border-radius: 10px;
	}

	.sidebar-account {
		order: 5;
		height: 40px;
		flex-basis: 40px;
		margin-top: 5px;
		background: #f3f5f7;
	}

	.navbar-username {
		margin-left: 10px;
		font-size: 13px;
	}
}

:global(.hideSidebar .sidebar-navbar-right) {
	align-items: center;
}

:global(.hideSidebar .sidebar-navbar-right .setting-item) {
	width: 36px;
	justify-content: center;
	padding: 0;
}

:global(.hideSidebar .sidebar-navbar-right .sidebar-account) {
	background: transparent;
}

:global(.hideSidebar .sidebar-navbar-right .navbar-username) {
	display: none;
}

.layout-top,
.layout-mix {
  .setting-item,
  .el-icon {
    color: var(--el-color-white);
  }
}

.dark .setting-item:hover {
	background: rgb(37 99 235 / 18%);
	color: #93c5fd;
}

:global(.account-dropdown-popper) {
	border-radius: 14px !important;
}

:global(.account-dropdown-popper .el-dropdown-menu) {
	padding: 6px 8px;
}

:global(.account-dropdown-popper .el-dropdown-menu__item) {
	margin: 0;
	border-radius: 7px;
	background: transparent !important;
}

:global(.account-dropdown-popper .el-dropdown-menu__item:hover),
:global(.account-dropdown-popper .el-dropdown-menu__item:focus) {
	background: transparent !important;
	color: var(--el-color-primary);
}
</style>
