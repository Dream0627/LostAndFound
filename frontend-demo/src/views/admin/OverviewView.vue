<!--
  数据概览（仅 mainadmin）：后台核心指标统计一览。
-->
<template>
  <div>
    <h2>数据概览</h2>
    <p class="muted text-sm">平台核心数据统计（实时）。</p>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <div v-else class="grid grid-stats">
      <div class="card stat">
        <div class="stat-num">{{ count.user_count }}</div>
        <div class="stat-label">注册用户</div>
      </div>
      <div class="card stat">
        <div class="stat-num">{{ count.post_count }}</div>
        <div class="stat-label">帖子总数</div>
      </div>
      <div class="card stat">
        <div class="stat-num">{{ count.pending_post_count }}</div>
        <div class="stat-label">待审核帖子</div>
      </div>
      <div class="card stat">
        <div class="stat-num">{{ count.pending_appeal_count }}</div>
        <div class="stat-label">待处理申诉</div>
      </div>
      <div class="card stat">
        <div class="stat-num">{{ count.today_post_count }}</div>
        <div class="stat-label">今日新增帖子</div>
      </div>
      <div class="card stat">
        <div class="stat-num">{{ count.today_comment_count }}</div>
        <div class="stat-label">今日新增评论</div>
      </div>
    </div>

    <button class="btn btn-ghost btn-sm mt-16" @click="refresh">刷新</button>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { getCount } from "@/api/admin";
import { useToastStore } from "@/stores/toast";

const toast = useToastStore();
const count = ref({});
const loading = ref(false);

async function refresh() {
  loading.value = true;
  try {
    count.value = await getCount();
  } catch (e) {
    toast.error(e?.msg || "加载失败");
  } finally {
    loading.value = false;
  }
}

onMounted(refresh);
</script>

<style scoped>
.grid-stats {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 14px;
}
.stat {
  text-align: center;
  padding: 22px 16px;
}
.stat-num {
  font-size: 30px;
  font-weight: 700;
  color: var(--color-primary-dark);
}
.stat-label {
  margin-top: 6px;
  font-size: 13px;
  color: var(--color-text-muted);
}
</style>