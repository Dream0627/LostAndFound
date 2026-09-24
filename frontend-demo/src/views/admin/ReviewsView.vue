<!--
  待审批：帖子发布 + 注销申诉 的待办列表（GET /admin/reviews）。
  可按类型筛选（post / appeal）；帖子可直接通过/驳回，申诉引导去申诉审核页。
-->
<template>
  <div>
    <div class="row-between">
      <h2>待审批</h2>
      <div class="row">
        <button
          v-for="opt in typeOptions"
          :key="opt.value"
          class="btn btn-sm"
          :class="type === opt.value ? '' : 'btn-ghost'"
          @click="setType(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <!-- 待审核帖子 -->
      <div v-if="showPosts" class="card">
        <h3>待审核帖子（{{ posts.length }}）</h3>
        <EmptyState v-if="posts.length === 0" text="没有待审核帖子" icon="✅" />
        <ul v-else class="review-list">
          <li v-for="p in posts" :key="p.id" class="review-item">
            <div class="review-info">
              <div class="row">
                <StatusBadge kind="type" :value="p.type" />
                <StatusBadge kind="status" :value="p.status" />
                <span class="review-title">{{ p.title }}</span>
              </div>
              <p class="muted text-sm">{{ excerpt(p.content) }}</p>
              <p class="muted text-sm">
                作者 ID：{{ p.user_id }} · {{ formatTime(p.created_at) }}
                <template v-if="p.location_name"> · 📍 {{ p.location_name }}</template>
              </p>
            </div>
            <div class="row">
              <router-link :to="`/posts/${p.id}`" class="btn btn-ghost btn-sm">查看</router-link>
              <button class="btn btn-sm" @click="handleReviewPost(p.id, 'approved')">通过</button>
              <button class="btn btn-danger btn-sm" @click="handleReviewPost(p.id, 'rejected')">驳回</button>
            </div>
          </li>
        </ul>
      </div>

      <!-- 待处理申诉（仅 mainadmin 可处理） -->
      <div v-if="showAppeals" class="card">
        <h3>待处理申诉（{{ appeals.length }}）</h3>
        <EmptyState v-if="appeals.length === 0" text="没有待处理申诉" icon="✅" />
        <ul v-else class="review-list">
          <li v-for="a in appeals" :key="a.id" class="review-item">
            <div class="review-info">
              <div class="row">
                <span class="tag tag-pending">申诉 #{{ a.id }}</span>
                <span class="review-title">{{ reasonLabel(a.reason) }}</span>
              </div>
              <p class="muted text-sm">{{ a.content || "（无补充说明）" }}</p>
              <p class="muted text-sm">用户 ID：{{ a.user_id }} · {{ formatTime(a.created_at) }}</p>
            </div>
            <div class="row">
              <template v-if="auth.isMainAdmin">
                <button class="btn btn-sm" @click="handleReviewAppeal(a.id, 'approved')">通过并恢复</button>
                <button class="btn btn-danger btn-sm" @click="handleReviewAppeal(a.id, 'rejected')">驳回</button>
              </template>
              <span v-else class="muted text-sm">需超级管理员处理</span>
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
import { reviewPost } from "@/api/post";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";
import StatusBadge from "@/components/StatusBadge.vue";
import EmptyState from "@/components/EmptyState.vue";

const auth = useAuthStore();
const toast = useToastStore();

const type = ref(""); // '' | 'post' | 'appeal'
const posts = ref([]);
const appeals = ref([]);
const loading = ref(false);

const typeOptions = [
  { value: "", label: "全部" },
  { value: "post", label: "帖子" },
  { value: "appeal", label: "申诉" },
];

const showPosts = ref(true);
const showAppeals = ref(true);

function excerpt(text) {
  const t = text || "";
  return t.length > 80 ? `${t.slice(0, 80)}…` : t;
}
function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}
function reasonLabel(r) {
  return { self_regret: "自行注销反悔", wrongful_ban: "被误封请求撤回", other: "其他" }[r] || r;
}

async function fetchReviews() {
  loading.value = true;
  try {
    const params = type.value ? { type: type.value } : {};
    const data = await listReviews(params);
    // 按类型过滤时后端只填充对应数组，未返回的按空数组处理。
    posts.value = data.posts || [];
    appeals.value = data.appeals || [];
    showPosts.value = !type.value || type.value === "post";
    showAppeals.value = !type.value || type.value === "appeal";
  } catch (e) {
    toast.error(e?.msg || "加载失败");
  } finally {
    loading.value = false;
  }
}

function setType(t) {
  type.value = t;
  fetchReviews();
}

async function handleReviewPost(postId, status) {
  try {
    await reviewPost(postId, status);
    toast.success(status === "approved" ? "已通过" : "已驳回");
    fetchReviews();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

async function handleReviewAppeal(appealId, status) {
  try {
    await reviewAppeal(appealId, status);
    toast.success(status === "approved" ? "已通过，账号已恢复" : "已驳回");
    fetchReviews();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

onMounted(fetchReviews);
</script>

<style scoped>
.review-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.review-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 0;
  border-top: 1px solid var(--color-border);
}
.review-info {
  flex: 1;
  min-width: 0;
}
.review-info p {
  margin: 6px 0 0;
}
.review-title {
  font-weight: 600;
}
</style>
