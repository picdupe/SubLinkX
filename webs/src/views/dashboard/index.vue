<template>
  <div class="dashboard-container">
    <section class="welcome-panel">
      <div class="welcome-copy">
        <h1>{{ greetings }}</h1>
      </div>
      <div class="welcome-mark" aria-hidden="true">
        <span class="mark-orbit orbit-one"></span>
        <span class="mark-orbit orbit-two"></span>
        <img
          class="mark-core"
          :src="avatarSrc"
          :alt="userStore.user.nickname || userStore.user.username"
          @error="handleAvatarError"
        />
      </div>
    </section>

    <section class="stats-grid" :aria-label="t('dashboard.resourceStatistics')">
      <el-card v-for="item in statisticData" :key="item.key" class="stat-card" shadow="never">
        <div class="stat-topline">
          <span class="stat-icon"><svg-icon :icon-class="item.iconClass" size="21px" /></span>
          <span class="stat-label">{{ item.title }}</span>
        </div>
        <div class="stat-value">{{ item.value }}</div>
      </el-card>
      <el-card class="stat-card status-card" shadow="never">
        <div class="stat-topline">
          <span class="status-dot"></span>
          <span class="stat-label">{{ t('dashboard.serviceStatus') }}</span>
        </div>
        <div class="status-value">{{ t('dashboard.running') }}</div>
      </el-card>
    </section>

    <section class="overview-panel">
      <div class="section-heading">
        <div>
          <h2>{{ t('dashboard.workspace') }}</h2>
        </div>
      </div>
      <div class="feature-row">
        <router-link to="/subcription/subs" class="feature-item"><span class="feature-index">01</span><strong>{{ t('dashboard.subscriptionManagement') }}</strong></router-link>
        <router-link to="/subcription/nodes" class="feature-item"><span class="feature-index">02</span><strong>{{ t('dashboard.nodeManagement') }}</strong></router-link>
        <router-link to="/subcription/template" class="feature-item"><span class="feature-index">03</span><strong>{{ t('dashboard.templateManagement') }}</strong></router-link>
        <router-link to="/relay/manager" class="feature-item"><span class="feature-index">04</span><strong>{{ t('dashboard.relayManagement') }}</strong></router-link>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: "Dashboard",
  inheritAttrs: false,
});

import { useI18n } from "vue-i18n";
import { useUserStore } from "@/store/modules/user";
import { getSubTotal,getNodeTotal } from "@/api/total";
const userStore = useUserStore();
const { t } = useI18n();
const displayName = computed(() => {
  const nickname = userStore.user.nickname?.trim();
  if (nickname && !/[\u3400-\u9fff]/.test(nickname)) return nickname;
  const username = userStore.user.username?.trim();
  if (username && !/[\u3400-\u9fff]/.test(username)) return username;
  return t("common.admin");
});
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
function handleAvatarError(event: Event) {
  const image = event.currentTarget as HTMLImageElement;
  if (!image.src.endsWith("/avatar.gif")) image.src = fallbackAvatar;
}
const date: Date = new Date();
const subTotal = ref(0);
const nodeTotal = ref(0);
// 右上角数量
const statisticData = computed(() => [
  {
    value: subTotal.value,
    iconClass: "link",
    title: t('dashboard.subscription'),
    key: "message",
  },
  {
    value: nodeTotal.value,
    iconClass: "publish",
    title: t('dashboard.nodes'),
    key: "upcoming",
  },
]);
const getsubtotal = async () => {
  const { data } = await getSubTotal();
  subTotal.value = data;
};
const getnodetotal = async () => {
  const { data } = await getNodeTotal();
  nodeTotal.value = data;
};
onMounted(() => {
  getsubtotal();
  getnodetotal();
});
const greetings = computed(() => {
  const hours = date.getHours();
  if (hours >= 6 && hours < 8) {
    return t('dashboard.greeting.earlyMorning');
  } else if (hours >= 8 && hours < 12) {
    return t('dashboard.greeting.morning', { name: displayName.value });
  } else if (hours >= 12 && hours < 18) {
    return t('dashboard.greeting.afternoon', { name: displayName.value });
  } else if (hours >= 18 && hours < 24) {
    return t('dashboard.greeting.evening', { name: displayName.value });
  } else {
    return t('dashboard.greeting.night');
  }
});




</script>

<style lang="scss" scoped>
.dashboard-container {
	width: 100%;
	padding: 0 0 30px;
	margin: 0;
	color: #172033;

	.welcome-panel {
		position: relative;
		display: flex;
		min-height: 220px;
		align-items: center;
		justify-content: space-between;
		overflow: hidden;
		padding: 38px 46px;
		border: 1px solid rgb(255 255 255 / 82%);
		border-radius: var(--app-radius);
		background: linear-gradient(112deg, rgb(255 255 255 / 90%), rgb(247 249 251 / 96%));
		box-shadow: 0 12px 30px rgb(37 49 63 / 7%), inset 0 1px 0 rgb(255 255 255 / 90%);
		backdrop-filter: blur(22px) saturate(145%);
		-webkit-backdrop-filter: blur(22px) saturate(145%);
	}

	h1 { margin: 0; font-size: clamp(25px, 3vw, 34px); font-weight: 700; letter-spacing: -.04em; }
	.welcome-mark { position: relative; display: grid; width: 150px; height: 150px; place-items: center; margin-right: 8%; }
	.mark-orbit { position: absolute; width: 132px; height: 132px; border: 1px solid rgb(64 158 255 / 24%); border-radius: 50%; }
	.orbit-two { width: 100px; height: 100px; border-color: rgb(64 158 255 / 36%); }
	.mark-core { display: block; width: 62px; height: 62px; border: 3px solid rgb(255 255 255 / 88%); border-radius: 50%; background: #5bd4d0; box-shadow: 0 12px 28px rgb(64 158 255 / 22%); object-fit: cover; }
	.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 15rem), 1fr)); gap: 18px; margin: 22px 0; }
	.stat-card { border: 1px solid #e8edf5; }
	:deep(.stat-card .el-card__body) { padding: 22px 24px; }
	.stat-topline { display: flex; align-items: center; gap: 10px; }
	.stat-icon { display: grid; width: 38px; height: 38px; place-items: center; border-radius: var(--app-radius); background: #eaf4ff; color: #2488ee; }
	.stat-label { color: #64748b; font-size: 14px; font-weight: 600; }
	.stat-value { margin-top: 18px; color: #172033; font-size: 34px; font-weight: 720; line-height: 1; letter-spacing: -.04em; }
	.status-dot { width: 9px; height: 9px; border-radius: 50%; background: #22c55e; box-shadow: 0 0 0 4px #dcfce7; }
	.status-value { margin-top: 18px; color: #16a34a; font-size: 25px; font-weight: 700; }
	.overview-panel { padding: 28px; border: 1px solid rgb(255 255 255 / 80%); border-radius: var(--app-radius); background: rgb(255 255 255 / 62%); box-shadow: inset 0 1px 0 rgb(255 255 255 / 90%), 0 12px 32px rgb(15 23 42 / 4%); backdrop-filter: blur(18px) saturate(140%); -webkit-backdrop-filter: blur(18px) saturate(140%); }
	.section-heading { display: flex; align-items: center; justify-content: space-between; }
	h2 { margin: 0; font-size: 20px; letter-spacing: -.02em; }
	.feature-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 14rem), 1fr)); gap: 14px; margin-top: 22px; }
	.feature-item { display: flex; align-items: center; gap: 13px; padding: 15px; border: 1px solid #e1e5e9; border-radius: var(--app-radius); background: #f7f8fa; color: inherit; text-decoration: none; transition: border-color 160ms ease, background 160ms ease, transform 160ms ease; }
	.feature-item:hover { transform: translateY(-2px); border-color: #b9dcff; background: #f1f8ff; }
	.feature-index { color: #2488ee; font-size: 12px; font-weight: 750; }
	.feature-item strong, .feature-item small { display: block; }
	.feature-item strong { font-size: 13px; }
	.svg-icon { fill: currentcolor !important; }
}

html.dark .dashboard-container { color: var(--el-text-color-primary); }
html.dark .welcome-panel { border-color: rgb(148 163 184 / 16%); background: linear-gradient(112deg, rgb(30 41 59 / 78%), rgb(23 37 58 / 62%)); }
html.dark .stat-value { color: var(--el-text-color-primary); }
html.dark .feature-item { border-color: #2a3444; background: #18202d; }
html.dark .feature-item:hover { border-color: rgb(96 165 250 / 35%); background: #1b2a40; }
html.dark .overview-panel { border-color: rgb(148 163 184 / 16%); background: rgb(15 23 42 / 58%); }

@media (max-width: 760px) {
	.dashboard-container { padding: 0 0 16px; }
	.dashboard-container .welcome-panel { min-height: 180px; padding: 26px; }
	.dashboard-container .welcome-mark { width: 100px; height: 100px; margin-right: 0; transform: scale(.75); }
	.dashboard-container .stats-grid, .dashboard-container .feature-row { grid-template-columns: minmax(0, 1fr); }
}
</style>
