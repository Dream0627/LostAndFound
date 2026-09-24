<!--
  申诉审核（仅 mainadmin）：将 pending 申诉审核为 approved / rejected。
  审核通过会自动级联恢复对应账号。
  这里以“待处理申诉”为工作台，也可通过 GET /admin/reviews?type=appeal 获取。
-->
<template>
  <div>
    <h2>申诉审核</h2>
    <p class="muted text-sm">审核通过（approved）时，系统会自动级联恢复该账号及其同批删除的内容。</p>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="appeals.length === 0" text="没有待处理申诉" icon="✅" />
      <div v-else class="card">
        <ul class="appeal-list">
          <li v-for="a in appeals" :key="a.id" class="appeal-item">
            <div class="review-info">
              <div class="row">
                <span class="tag tag-pending">申诉 #{{ a.id }}</span>
                <span class="review-title">{{ reasonLabel(a.reason) }}</span>
              </div>
              <p class="appeal-content">{{ a.content || "（无补充说明）" }}</p>
              <p class="muted text-sm">用户 ID：{{ a.user_id }} · {{ formatTime(a.created_at) }}</p>
            </div>
            <div class="row">
              <button class="btn btn-sm" @click="handleReview(a.id, 'approved')">通过并恢复</button>
              <button class="btn btn-danger btn-sm" @click="handleReview(a.id, 'rejected')">驳回</button>
            </div>
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listReviews, reviewAppeal } from "@/api/admin";
import { useToastStore } from "@/stores/toast";
import EmptyState from "@/components/EmptyState.vue";

const toast = useToastStore();
const appeals = ref([]);
const loading = ref(false);

function reasonLabel(r) {
  return { self_regret: "自行注销反悔", wrongful_ban: "被误封请求撤回", other: "其他" }[r] || r;
}
function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchAppeals() {
  loading.value = true;
  try {
    const data = await listReviews({ type: "appeal" });
    appeals.value = data.appeals || [];
  } catch (e) {
    toast.error(e?.msg || "加载失败");
  } finally {
    loading.value = false;
  }
}

async function handleReview(appealId, status) {
  try {
    await reviewAppeal(appealId, status);
    toast.success(status === "approved" ? "已通过，账号已恢复" : "已驳回");
    fetchAppeals();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

onMounted(fetchAppeals);
</script>

<style scoped>
.appeal-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.appeal-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid var(--color-border);
}
.appeal-item:last-child {
  border-bottom: none;
}
.review-info {
  flex: 1;
}
.review-title {
  font-weight: 600;
}
.appeal-content {
  margin: 8px 0 6px;
  white-space: pre-wrap;
}
</style>
