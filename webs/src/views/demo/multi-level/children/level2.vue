<template>
  <div style="padding: 30px">
    <el-alert :closable="false" :title="$t('route.level2')" type="success">
      <router-view />
    </el-alert>
  </div>
</template>
