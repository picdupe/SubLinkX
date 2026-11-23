<template>
  <div style="padding: 30px">
    <el-link
      href="https://gitee.com/youlaiorg/vue3-element-admin/blob/master/src/views/demo/multi-level/level1.vue"
      type="primary"
      target="_blank"
      class="mb-10"
      >{{ $t('demo.sourceLink') }}</el-link
    >

    <el-alert :closable="false" :title="$t('route.level1')">
      <router-view />
    </el-alert>
  </div>
</template>
