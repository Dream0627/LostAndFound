<!-- 空状态占位：列表无数据或加载失败时展示，保持界面不空。 -->
<template>
  <div class="empty">
    <div class="empty-icon">{{ icon }}</div>
    <p class="empty-text">{{ text }}</p>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  text: { type: String, default: "暂无数据" },
  icon: { type: String, default: "🗂️" },
});
</script>

<style scoped>
.empty {
  padding: 48px 20px;
  text-align: center;
  color: var(--color-text-muted);
}
.empty-icon {
  font-size: 40px;
  margin-bottom: 8px;
}
.empty-text {
  margin: 0 0 12px;
}
</style>
