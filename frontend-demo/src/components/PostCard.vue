<!--
  帖子卡片：列表页展示。含类型/状态徽章、标题、内容摘要、地点、时间与操作入口。
  点击卡片主体进入详情；管理员可在此直接审核（emit 事件由父页面处理）。
-->
<template>
  <article class="post-card" @click="$router.push(`/posts/${post.id}`)">
    <div class="pc-head">
      <StatusBadge kind="type" :value="post.type" />
      <StatusBadge kind="finished" :value="post.is_finished" />
      <StatusBadge v-if="showStatus" kind="status" :value="post.status" />
    </div>

    <h3 class="pc-title">{{ post.title }}</h3>
    <p class="pc-content">{{ excerpt }}</p>

    <img v-if="post.image_url" :src="post.image_url" class="pc-image" alt="帖子图片" />

    <div class="pc-meta">
      <span v-if="post.location_name" class="pc-loc">📍 {{ post.location_name }}</span>
      <span class="pc-time muted">{{ formatTime(post.created_at) }}</span>
    </div>
  </article>
</template>

<script setup>
import { computed } from "vue";
import StatusBadge from "./StatusBadge.vue";

const props = defineProps({
  post: { type: Object, required: true },
  showStatus: { type: Boolean, default: false },
});

const excerpt = computed(() => {
  const text = props.post.content || "";
  return text.length > 90 ? `${text.slice(0, 90)}…` : text;
});

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString("zh-CN", { hour12: false });
}
</script>

<style scoped>
.post-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 16px;
  cursor: pointer;
  transition: transform 0.12s ease, box-shadow 0.12s ease, border-color 0.12s ease;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.post-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow);
  border-color: var(--color-primary);
}
.pc-head {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.pc-title {
  margin: 0;
  font-size: 16px;
  line-height: 1.4;
}
.pc-content {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 14px;
  min-height: 42px;
}
.pc-image {
  width: 100%;
  height: 150px;
  object-fit: cover;
  border-radius: var(--radius-sm);
}
.pc-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  margin-top: auto;
}
.pc-loc {
  color: var(--color-primary-dark);
}
</style>
