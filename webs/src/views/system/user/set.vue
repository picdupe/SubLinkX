<script setup lang="ts">
import PasswordSettings from "@/views/settings/PasswordSettings.vue";
</script>

<template>
  <div class="password-page">
    <section class="password-panel">
      <h2>{{ $t("userset.title") }}</h2>
      <PasswordSettings />
    </section>
  </div>
</template>

<style lang="scss" scoped>
.password-page {
  padding: 20px;
}

.password-panel {
  max-width: 520px;
  padding: 24px;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--app-radius);
  background: var(--el-bg-color-overlay);

  h2 {
    margin: 0 0 18px;
    font-size: 16px;
  }
}
</style>
