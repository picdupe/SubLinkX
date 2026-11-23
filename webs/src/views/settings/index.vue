<template>
  <div class="settings-page">
    <section class="settings-section">
      <h2>{{ $t('settings.theme') }}</h2>
      <div class="settings-row">
        <span>{{ $t('settings.darkMode') }}</span>
        <div class="settings-control-slot"><el-switch v-model="isDark" /></div>
      </div>
      <div class="settings-row">
        <span>{{ $t('settings.themeColor') }}</span>
        <div class="settings-control-slot"><ThemeColorPicker v-model="settingsStore.themeColor" /></div>
      </div>
      <div class="settings-row">
        <span>{{ $t('settings.language') }}</span>
        <div class="settings-control-slot">
          <div class="settings-language-control"><LangSelect /></div>
        </div>
      </div>
    </section>

    <section class="settings-section">
      <h2>{{ $t('settings.interface') }}</h2>
      <div class="settings-row">
        <span>{{ $t('settings.sidebarLogo') }}</span>
        <div class="settings-control-slot"><el-switch v-model="settingsStore.sidebarLogo" /></div>
      </div>
      <div class="settings-row">
        <span>{{ $t('settings.watermark') }}</span>
        <div class="settings-control-slot"><el-switch v-model="settingsStore.watermarkEnabled" /></div>
      </div>
    </section>

    <section class="settings-section">
      <h2>{{ $t("userset.title") }}</h2>
      <PasswordSettings />
    </section>
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from "@/store";
import { ThemeEnum } from "@/enums/ThemeEnum";
import PasswordSettings from "@/views/settings/PasswordSettings.vue";

const settingsStore = useSettingsStore();
const isDark = computed({
  get: () => settingsStore.theme === ThemeEnum.DARK,
  set: (enabled: boolean) =>
    settingsStore.changeTheme(enabled ? ThemeEnum.DARK : ThemeEnum.LIGHT),
});
</script>

<style lang="scss" scoped>
.settings-page {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
  gap: 18px;
  padding: 0 0 12px;
}

.settings-section {
  min-width: 0;
  padding: 22px 24px;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--app-radius);
  background: var(--el-bg-color-overlay);

  h2 {
    margin: 0 0 18px;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 650;
  }
}

.settings-row {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: var(--el-text-color-regular);

}

.settings-control-slot {
  display: flex;
  width: 52px;
  min-width: 52px;
  min-height: 40px;
  align-items: center;
  justify-content: center;
}

.settings-language-control {
  box-sizing: border-box;
  display: flex;
  width: 44px;
  height: 36px;
  align-items: center;
  justify-content: center;
  padding: 0 10px;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--app-radius);
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
  }
}
</style>
