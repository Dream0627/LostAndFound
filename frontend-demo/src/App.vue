<!--
  根组件：顶部导航 + 内容区 + 轻提示容器。
  导航栏仅在有登录态时展示“发布/消息/我的”等入口（见 AppNavbar）。
-->
<template>
  <AppNavbar />
  <main class="app-main">
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </main>
  <footer class="app-footer">
    <span>LAF · 校园失物招领 · 前端示例（Vue3 全家桶）</span>
  </footer>
  <ToastHost />
</template>

<script setup>
import AppNavbar from "./components/AppNavbar.vue";
import ToastHost from "./components/ToastHost.vue";
</script>

<style scoped>
.app-main {
  flex: 1;
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  padding: 24px 16px 48px;
  box-sizing: border-box;
}
.app-footer {
  padding: 20px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}
/* 路由切换的淡入淡出，提升观感。 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.16s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
