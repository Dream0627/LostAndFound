<!--
  帖子广场：筛选（类型 / 完成状态 / 管理员可加状态） + 列表 + 分页。
  后端已将“未完成”排在前面，这里只需把筛选条件透传。
-->
<template>
  <div>
    <div class="page-head">
      <div>
        <h2>帖子广场</h2>
        <p class="muted text-sm">发现身边的失物与招领信息</p>
      </div>
      <router-link v-if="auth.isLoggedIn" to="/posts/new" class="btn">＋ 发布帖子</router-link>
    </div>

    <div class="card filters">
      <div class="filter-group">
        <span class="filter-label">类型</span>
        <label class="chip" :class="{ active: filters.type.length === 0 }">
          <input type="checkbox" :checked="filters.type.length === 0" @change="setAllTypes" /> 全部
        </label>
        <label class="chip" :class="{ active: filters.type.includes('lost') }">
          <input type="checkbox" :checked="filters.type.includes('lost')" @change="toggleType('lost')" /> 失物
        </label>
        <label class="chip" :class="{ active: filters.type.includes('found') }">
          <input type="checkbox" :checked="filters.type.includes('found')" @change="toggleType('found')" /> 招领
        </label>
      </div>

      <div class="filter-group">
        <span class="filter-label">完成状态</span>
        <label v-for="opt in finishedOptions" :key="opt.value" class="chip" :class="{ active: filters.finished === opt.value }">
          <input type="radio" :checked="filters.finished === opt.value" @change="setFinished(opt.value)" /> {{ opt.label }}
        </label>
      </div>

      <div v-if="auth.isPostAdmin" class="filter-group">
        <span class="filter-label">审核状态（管理员）</span>
        <label v-for="opt in statusOptions" :key="opt" class="chip" :class="{ active: filters.status.includes(opt) }">
          <input type="checkbox" :checked="filters.status.includes(opt)" @change="toggleStatus(opt)" /> {{ statusLabel(opt) }}
        </label>
      </div>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="posts.length === 0" text="没有符合条件的帖子" icon="🔍">
        <router-link v-if="auth.isLoggedIn" to="/posts/new" class="btn">发布第一条</router-link>
      </EmptyState>

      <div v-else class="grid grid-posts mt-16">
        <PostCard v-for="post in posts" :key="post.id" :post="post" :show-status="auth.isPostAdmin" />
      </div>

      <Pagination :page="page" :page-size="pageSize" :total="total" @change="onPageChange" />
    </template>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { listPosts } from "@/api/post";
import { useAuthStore } from "@/stores/auth";
import PostCard from "@/components/PostCard.vue";
import Pagination from "@/components/Pagination.vue";
import EmptyState from "@/components/EmptyState.vue";

const auth = useAuthStore();

const posts = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(12);
const loading = ref(false);

const filters = reactive({
  type: ["lost", "found"], // 默认全选，等价于不筛选（发送时若为全集则不传）
  finished: "", // '' | 'true' | 'false'
  status: [], // 仅管理员有效
});

const finishedOptions = [
  { value: "", label: "全部" },
  { value: "false", label: "进行中" },
  { value: "true", label: "已完成" },
];
const statusOptions = ["pending", "approved", "rejected"];

function statusLabel(s) {
  return { pending: "待审核", approved: "已通过", rejected: "已驳回" }[s] || s;
}

function setAllTypes() {
  filters.type = ["lost", "found"];
  page.value = 1;
  fetchList();
}
function toggleType(t) {
  const idx = filters.type.indexOf(t);
  if (idx >= 0) filters.type.splice(idx, 1);
  else filters.type.push(t);
  // 不允许清空到 0（否则语义含糊），退化为全选。
  if (filters.type.length === 0) filters.type = ["lost", "found"];
  page.value = 1;
  fetchList();
}
function setFinished(v) {
  filters.finished = v;
  page.value = 1;
  fetchList();
}
function toggleStatus(s) {
  const idx = filters.status.indexOf(s);
  if (idx >= 0) filters.status.splice(idx, 1);
  else filters.status.push(s);
  page.value = 1;
  fetchList();
}

function buildParams() {
  const params = { page: page.value, page_size: pageSize.value };
  // 仅当类型不是“全集”时才传，避免无意义的筛选。
  if (filters.type.length === 1) params.type = [filters.type[0]];
  if (filters.finished) params.finished = filters.finished;
  if (auth.isPostAdmin && filters.status.length > 0) params.status = [...filters.status];
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
  } finally {
    loading.value = false;
  }
}

function onPageChange(p) {
  page.value = p;
  fetchList();
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(fetchList);

// 登录/登出后角色变化（管理员筛选），刷新列表。
watch(() => auth.user?.role, () => {
  page.value = 1;
  fetchList();
});
</script>

<style scoped>
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}
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
</style>
