<!--
  已删帖子：列表 + 恢复（PATCH /posts/:id/recover）。
  注意：恢复帖子不会自动恢复其评论。
-->
<template>
  <div>
    <h2>已删除帖子</h2>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="posts.length === 0" text="没有已删除的帖子" icon="🗑️" />
      <div v-else class="card">
        <ul class="deleted-list">
          <li v-for="p in posts" :key="p.id" class="deleted-item">
            <div class="review-info">
              <div class="row">
                <StatusBadge kind="type" :value="p.type" />
                <span class="review-title">{{ p.title }}</span>
              </div>
              <p class="muted text-sm">
                帖子 #{{ p.id }} · 作者 ID：{{ p.user_id }} · {{ formatTime(p.created_at) }}
              </p>
            </div>
            <button class="btn btn-sm" @click="handleRecover(p.id)">恢复</button>
          </li>
        </ul>
      </div>

      <Pagination :page="page" :page-size="pageSize" :total="total" @change="onPageChange" />
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listDeletedPosts } from "@/api/admin";
import { recoverPost } from "@/api/post";
import { useToastStore } from "@/stores/toast";
import StatusBadge from "@/components/StatusBadge.vue";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const toast = useToastStore();
const posts = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const loading = ref(false);

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchList() {
  loading.value = true;
  try {
    const data = await listDeletedPosts({ page: page.value, page_size: pageSize.value });
    posts.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    toast.error(e?.msg || "加载失败");
  } finally {
    loading.value = false;
  }
}

async function handleRecover(postId) {
  try {
    await recoverPost(postId);
    toast.success("已恢复（评论不会一并恢复）");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "恢复失败");
  }
}

function onPageChange(p) {
  page.value = p;
  fetchList();
}

onMounted(fetchList);
</script>

<style scoped>
.deleted-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.deleted-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid var(--color-border);
}
.deleted-item:last-child {
  border-bottom: none;
}
.review-info {
  flex: 1;
}
.review-info p {
  margin: 6px 0 0;
}
.review-title {
  font-weight: 600;
}
</style>
