<template>
  <!-- 顶部导航栏 -->
  <div class="navbar-container">
    <!-- 导航栏左侧 -->
    <NavbarLeft />
  </div>
</template>

<style lang="scss" scoped>
.navbar-container {
  @apply flex-x-between;

	box-sizing: border-box;
	width: 100%;
  height: $navbar-height;
	padding: 0 clamp(12px, 1.4vw, 22px);
	border-bottom: 1px solid #e5e8eb;
	background: rgb(255 255 255 / 96%);
	backdrop-filter: blur(18px) saturate(145%);
	-webkit-backdrop-filter: blur(18px) saturate(145%);
}

html.dark .navbar-container {
	border-bottom-color: rgb(148 163 184 / 14%);
	background: rgb(15 23 42 / 74%);
}
</style>
