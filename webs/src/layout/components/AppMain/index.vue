<template>
  <section class="app-main">
    <router-view>
      <template #default="{ Component, route }">
        <transition
          enter-active-class="animate__animated animate__fadeIn"
          mode="out-in"
        >
          <keep-alive :include="cachedViews">
            <component :is="Component" :key="route.path" />
          </keep-alive>
        </transition>
      </template>
    </router-view>
  </section>
</template>

<script setup lang="ts">
import { useTagsViewStore } from "@/store";

const cachedViews = computed(() => useTagsViewStore().cachedViews); // 缓存页面集合
</script>

<style lang="scss" scoped>
.app-main {
  position: relative;
  width: 100%;
  box-sizing: border-box;
  min-height: calc(100vh - 28px);
  overflow: visible;
  border-radius: 22px;
  background: transparent;
  box-shadow: none;
}

.hasTagsView .app-main {
  min-height: calc(100vh - 28px);
}

.fixed-header + .app-main {
  min-height: calc(100vh - 28px);
  padding-top: 0;
}

.hasTagsView .fixed-header + .app-main {
  min-height: calc(100vh - 28px);
  padding-top: 0;
}

.layout-mix,
.layout-top {
  .fixed-header + .app-main {
    padding-top: 0;
  }
}

.layout-mix {
  .app-main {
    height: calc(100vh - 28px);
    min-height: calc(100vh - 28px);
    padding-top: 0;
    overflow-y: auto;
  }

  .hasTagsView .app-main {
    height: calc(100vh - 28px);
    min-height: calc(100vh - 28px);
  }

  .fixed-header + .app-main {
    min-height: calc(100vh - 28px);
  }

  .hasTagsView .fixed-header + .app-main {
    height: calc(100vh - 28px);
    min-height: calc(100vh - 28px);
    padding-top: 0;
  }
}

.layout-top {
  .hasTagsView .fixed-header + .app-main {
    padding-top: $tags-view-height;
  }
}
</style>
