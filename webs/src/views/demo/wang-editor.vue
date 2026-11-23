<!-- wangEditor富文本编辑器示例 -->
<script setup lang="ts">
import Editor from "@/components/WangEditor/index.vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const value = ref(t('demo.editorInitial'));
</script>

<template>
  <div class="app-container">
    <el-link
      href="https://gitee.com/youlaiorg/vue3-element-admin/blob/master/src/views/demo/wang-editor.vue"
      type="primary"
      target="_blank"
      class="mb-[20px]"
      >{{ $t('demo.sourceLink') }}</el-link
    >
    <editor v-model="value" style="height: calc(100vh - 180px)" />
  </div>
</template>
