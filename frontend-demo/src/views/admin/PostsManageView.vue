<!--
  帖子管理（postadmin/mainadmin）：集中管理帖子状态，直观在列表中直接操作。
  - 支持关键词搜索 + 审核状态 / 完成状态筛选；
  - 待审核帖：通过 / 驳回（走审核接口）；
  - 已通过帖：下架（改 rejected）；已驳回帖：重新通过（改 approved）；均不可改回待审核；
  - 任意状态：标记完成 / 取消完成（直接改 is_finished）。
-->
<template>
  <div>
    <div class="page-head">
      <div>
        <h2>帖子管理</h2>
        <p class="muted text-sm">一站式管理帖子：搜索、筛选，并直接在列表中修改审核状态与完成状态。</p>
      </div>
    </div>

    <div class="card filters">
      <div class="search-bar">
        <input
          v-model.trim="keyword"
          class="input"
          placeholder="按标题搜索物品名称…"
          @keydown.enter="applyFilters"
        />
        <button class="btn btn-sm" @click="applyFilters">搜索</button>
        <button v-if="keyword" class="btn btn-ghost btn-sm" @click="clearKeyword">清空</button>
      </div>

      <div class="filter-group">
        <span class="filter-label">审核状态</span>
        <label v-for="opt in statusOptions" :key="opt.value" class="chip" :class="{ active: status === opt.value }">
          <input type="radio" :checked="status === opt.value" @change="setStatus(opt.value)" /> {{ opt.label }}
        </label>
      </div>

      <div class="filter-group">
        <span class="filter-label">完成状态</span>
        <label v-for="opt in finishedOptions" :key="opt.value" class="chip" :class="{ active: finished === opt.value }">
          <input type="radio" :checked="finished === opt.value" @change="setFinished(opt.value)" /> {{ opt.label }}
        </label>
      </div>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="posts.length === 0" text="没有符合条件的帖子" icon="🔍" />

      <div v-else class="card">
        <ul class="pm-list">
          <li v-for="p in posts" :key="p.id" class="pm-item">
            <div class="pm-info">
              <div class="row">
                <StatusBadge kind="type" :value="p.type" />
                <StatusBadge kind="status" :value="p.status" />
                <StatusBadge kind="finished" :value="p.is_finished" />
                <router-link :to="`/posts/${p.id}`" class="pm-title">{{ p.title }}</router-link>
              </div>
              <p class="muted text-sm">
                帖子 #{{ p.id }} · 作者：{{ p.author_name || ('用户 ' + p.user_id) }} · {{ formatTime(p.created_at) }}
                <template v-if="p.location_name"> · 📍 {{ p.location_name }}</template>
              </p>
            </div>

            <div class="pm-actions">
              <!-- 待审核：走审核流 -->
              <template v-if="p.status === 'pending'">
                <button class="btn btn-sm" @click="handleReview(p, 'approved')">通过</button>
                <button class="btn btn-danger btn-sm" @click="handleReview(p, 'rejected')">驳回</button>
              </template>
              <!-- 已通过：可下架；已驳回：可重新通过 -->
              <button v-else-if="p.status === 'approved'" class="btn btn-ghost btn-sm" @click="handleStatus(p, 'rejected')">
                下架
              </button>
              <button v-else-if="p.status === 'rejected'" class="btn btn-sm" @click="handleStatus(p, 'approved')">
                重新通过
              </button>

              <!-- 完成状态切换 -->
              <button class="btn btn-ghost btn-sm" @click="handleToggleFinished(p)">
                {{ p.is_finished ? "取消完成" : "标记完成" }}
              </button>
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
      </div>
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listPosts, reviewPost } from "@/api/post";
import { updatePostStatus, updatePostFinished } from "@/api/admin";
import { useToastStore } from "@/stores/toast";
import StatusBadge from "@/components/StatusBadge.vue";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const toast = useToastStore();

const posts = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);

const keyword = ref("");
const status = ref(""); // '' | pending | approved | rejected
const finished = ref(""); // '' | 'true' | 'false'

const statusOptions = [
  { value: "", label: "全部" },
  { value: "pending", label: "待审核" },
  { value: "approved", label: "已通过" },
  { value: "rejected", label: "已驳回" },
];
const finishedOptions = [
  { value: "", label: "全部" },
  { value: "false", label: "进行中" },
  { value: "true", label: "已完成" },
];

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

function buildParams() {
  const params = { page: page.value, page_size: pageSize.value };
  if (keyword.value) params.keyword = keyword.value;
  if (status.value) params.status = [status.value];
  if (finished.value) params.finished = finished.value;
  return params;
}

async function fetchList() {
  loading.value = true;
  try {
    const data = await listPosts(buildParams());
    posts.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    posts.value = [];
    total.value = 0;
    toast.error(e?.msg || "加载失败");
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  fetchList();
}
function clearKeyword() {
  keyword.value = "";
  applyFilters();
}
function setStatus(v) {
  status.value = v;
  applyFilters();
}
function setFinished(v) {
  finished.value = v;
  applyFilters();
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

async function handleReview(p, target) {
  const tip = target === "approved" ? `确定通过帖子「${p.title}」？` : `确定驳回帖子「${p.title}」？驳回后仅作者与管理员可见。`;
  if (!window.confirm(tip)) return;
  try {
    await reviewPost(p.id, target);
    toast.success(target === "approved" ? "已通过审核" : "已驳回");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

async function handleStatus(p, target) {
  const tip = target === "rejected" ? `确定下架帖子「${p.title}」？` : `确定重新通过帖子「${p.title}」？`;
  if (!window.confirm(tip)) return;
  try {
    await updatePostStatus(p.id, target);
    toast.success(target === "rejected" ? "已下架" : "已重新通过");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

async function handleToggleFinished(p) {
  const target = !p.is_finished;
  const tip = target ? `确定将帖子「${p.title}」标记为已完成？` : `确定取消帖子「${p.title}」的已完成状态？`;
  if (!window.confirm(tip)) return;
  try {
    await updatePostFinished(p.id, target);
    toast.success(target ? "已标记完成" : "已取消完成");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

onMounted(fetchList);
</script>

<style scoped>
.page-head {
  margin-bottom: 16px;
}
.filters {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.search-bar {
  display: flex;
  gap: 8px;
}
.search-bar .input {
  flex: 1;
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
  min-width: 72px;
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
.pm-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.pm-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 0;
  border-top: 1px solid var(--color-border);
}
.pm-item:first-child {
  border-top: none;
}
.pm-info {
  flex: 1;
  min-width: 0;
}
.pm-info p {
  margin: 6px 0 0;
}
.pm-title {
  font-weight: 600;
}
.pm-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  justify-content: flex-end;
}
@media (max-width: 720px) {
  .pm-item {
    flex-direction: column;
  }
  .pm-actions {
    justify-content: flex-start;
  }
}
</style>