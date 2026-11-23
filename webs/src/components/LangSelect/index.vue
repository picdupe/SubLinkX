<template>
  <el-dropdown
    trigger="click"
    popper-class="language-dropdown-popper"
    @command="handleLanguageChange"
  >
    <div>
      <svg-icon icon-class="language" :size="size" />
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in langOptions"
          :key="item.value"
          :disabled="appStore.language === item.value"
          :command="item.value"
        >
          {{ item.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/store/modules/app";
import { LanguageEnum } from "@/enums/LanguageEnum";

defineProps({
  size: {
    type: String,
    required: false,
  },
});

const appStore = useAppStore();
const { locale, t } = useI18n();
const langOptions = computed(() => [
  { label: t("langSelect.zhCn"), value: LanguageEnum.ZH_CN },
  { label: t("langSelect.en"), value: LanguageEnum.EN },
]);

function handleLanguageChange(lang: string) {
  locale.value = lang;
  appStore.changeLanguage(lang);

  ElMessage.success(t("langSelect.message.success"));
}
</script>

<style lang="scss" scoped>
:global(.language-dropdown-popper) {
  border-radius: 14px !important;
}

:global(.language-dropdown-popper .el-dropdown-menu) {
  padding: 5px;
}

:global(.language-dropdown-popper .el-dropdown-menu__item) {
	 margin: 2px 4px;
  border-radius: 9px;
}
</style>
