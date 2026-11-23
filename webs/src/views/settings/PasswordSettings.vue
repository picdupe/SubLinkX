<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { updateUserPassword } from "@/api/user";
import { useUserStore } from "@/store";

const { t } = useI18n();
const userStore = useUserStore();
const userinfo = ref<{ username?: string; avatar?: string }>();
const username = ref("");
const password = ref("");
const passwordInput = ref();
const confirmationOpen = ref(false);
const isSubmitting = ref(false);
const hasLoggedOut = ref(false);

function focusPassword() {
  passwordInput.value?.focus();
}

onMounted(async () => {
  userinfo.value = await userStore.getUserInfo();
});

async function resetPassword() {
  if (confirmationOpen.value || isSubmitting.value || hasLoggedOut.value) return;
  if (!username.value.trim() || !password.value.trim()) {
    ElMessage.error(t("userset.message.xx1"));
    return;
  }
  const passwordBytes = new TextEncoder().encode(password.value.trim()).length;
  if (passwordBytes < 8 || passwordBytes > 72) {
    ElMessage.error(t("userset.message.xx2"));
    return;
  }

  confirmationOpen.value = true;
  try {
    await ElMessageBox.confirm(t("userset.message.xx3"), t("userset.message.title"), {
      confirmButtonText: t("common.confirm"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    });
  } catch {
    confirmationOpen.value = false;
    return;
  }

  confirmationOpen.value = false;
  isSubmitting.value = true;
  try {
    await updateUserPassword({
      username: username.value.trim(),
      password: password.value.trim(),
    });
    hasLoggedOut.value = true;
    ElMessage.success(t("userset.message.xx4"));
    await userStore.resetToken();
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <form class="password-settings" @submit.prevent="resetPassword">
    <div v-if="userinfo?.username" class="current-user">
      <span>{{ $t('userset.currentAccount') }}</span>
      <el-tag round>{{ userinfo.username }}</el-tag>
    </div>
    <el-input
      v-model="username"
      @keydown.enter.prevent.stop="focusPassword"
      :placeholder="$t('userset.newUsername')"
      autocomplete="username"
    />
    <el-input
      ref="passwordInput"
      v-model="password"
      type="password"
      show-password
      @keydown.enter.prevent.stop="resetPassword"
      autocomplete="new-password"
      :placeholder="$t('userset.newPassword')"
    />
    <el-button type="primary" native-type="submit" :loading="isSubmitting" :disabled="confirmationOpen || hasLoggedOut">{{ $t('userset.updateAction') }}</el-button>
  </form>
</template>

<style lang="scss" scoped>
.password-settings {
  display: grid;
  gap: 12px;
}

.current-user {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.password-settings .el-button {
  justify-self: start;
}
</style>
