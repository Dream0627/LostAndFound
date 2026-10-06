<!--
  反馈审批（仅 mainadmin）：查看用户反馈，按状态过滤，并对待处理反馈采纳/驳回。
  注意：反馈状态语义与帖子不同——pending 待处理、approved 已采纳、rejected 已驳回。
-->
<template>
  <div>
    <h2>反馈审批</h2>
    <p class="muted text-sm">查看用户提交的反馈。只有「待处理」的反馈可审批为已采纳 / 已驳回。</p>

    <div class="card filters">
      <div class="filter-group">
        <span class="filter-label">状态</span>
        <label v-for="opt in statusOptions" :key="opt.value" class="chip" :class="{ active: status === opt.value }">
          <input type="radio" :checked="status === opt.value" @change="setStatus(opt.value)" /> {{ opt.label }}
        </label>
      </div>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="feedbacks.length === 0" text="暂无反馈" icon="💬" />

      <ul v-else class="fb-list">
        <li v-for="f in feedbacks" :key="f.id" class="card">
          <div class="row-between">
            <div class="row">
              <span class="fb-user">{{ f.author_name || ("用户 " + f.user_id) }}</span>
              <span class="tag" :class="statusClass(f.status)">{{ statusLabel(f.status) }}</span>
            </div>
            <span class="muted text-sm">{{ formatTime(f.created_at) }}</span>
          </div>
          <p class="fb-content">{{ f.content }}</p>
          <div v-if="f.status === 'pending'" class="row mt-8">
            <button class="btn btn-sm" @click="handleReview(f, 'approved')">采纳</button>
            <button class="btn btn-ghost btn-sm" @click="handleReview(f, 'rejected')">驳回</button>
          </div>
        </li>
      </ul>

      <Pagination
        :page="page"
        :page-size="pageSize"
        :total="total"
        @change="onPageChange"
        @change-page-size="onPageSizeChange"
      />
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listFeedbacks, reviewFeedback } from "@/api/admin";
import { useToastStore } from "@/stores/toast";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const toast = useToastStore();

const feedbacks = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const status = ref("");

const statusOptions = [
  { value: "", label: "全部" },
  { value: "pending", label: "待处理" },
  { value: "approved", label: "已采纳" },
  { value: "rejected", label: "已驳回" },
];

function statusLabel(s) {
  return { pending: "待处理", approved: "已采纳", rejected: "已驳回" }[s] || s;
}
function statusClass(s) {
  return { pending: "tag-pending", approved: "tag-approved", rejected: "tag-rejected" }[s] || "";
}

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

function setStatus(v) {
  status.value = v;
  page.value = 1;
  fetchList();
}

async function fetchList() {
  loading.value = true;
  try {
    const params = { page: page.value, page_size: pageSize.value };
    if (status.value) params.status = status.value;
    const data = await listFeedbacks(params);
    feedbacks.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    feedbacks.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

function onPageChange(p) {
  page.value = p;
  fetchList();
}

function onPageSizeChange(size) {
  pageSize.value = size;
  page.value = 1;
  fetchList();
}

async function handleReview(f, target) {
  const tip = target === "approved" ? "确定采纳这条反馈？" : "确定驳回这条反馈？";
  if (!window.confirm(tip)) return;
  try {
    await reviewFeedback(f.id, target);
    toast.success(target === "approved" ? "已采纳" : "已驳回");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

onMounted(fetchList);
</script>

<style scoped>
.filters {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.filter-label {
  font-size: 13px;
  color: var(--color-text-muted);
  min-width: 92px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px;
  border: 1px solid var(--color-border);
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
  user-select: none;
  transition: all 0.12s ease;
}
.chip input {
  display: none;
}
.chip.active {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
  color: var(--color-primary-dark);
  font-weight: 600;
}
.fb-list {
  list-style: none;
  padding: 0;
  margin: 16px 0 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.fb-user {
  font-weight: 600;
  font-size: 14px;
}
.fb-content {
  white-space: pre-wrap;
  margin: 10px 0 0;
  font-size: 14px;
}
</style>