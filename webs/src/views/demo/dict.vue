<!-- 字典组件示例 -->
<script setup lang="ts">
const stringValue = ref("1"); // 性别(值为String)
const numberValue = ref(1); // 性别(值为Number)
</script>

<template>
  <div class="app-container">
    <el-link
      href="https://gitee.com/youlaiorg/vue3-element-admin/blob/master/src/views/demo/dict.vue"
      type="primary"
      target="_blank"
      class="mb-[20px]"
      >{{ $t('demo.sourceLink') }}</el-link
    >
    <el-form>
      <el-form-item :label="$t('demo.gender')">
        <dictionary v-model="stringValue" type-code="gender" />
        <el-link :underline="false" type="primary" class="ml-5"
          >{{ $t('demo.stringValue') }}
        </el-link>
      </el-form-item>

      <el-form-item :label="$t('demo.gender')">
        <dictionary v-model="numberValue" type-code="gender" />
        <el-link :underline="false" type="success" class="ml-5"
          >{{ $t('demo.numberValue') }}
        </el-link>
      </el-form-item>
    </el-form>
  </div>
</template>
