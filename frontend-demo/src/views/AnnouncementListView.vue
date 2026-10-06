<!--
  公告列表：全站公开，展示管理员发布的公告。
-->
<template>
  <div>
    <div class="page-head">
      <div>
        <h2>公告</h2>
        <p class="muted text-sm">平台公告与重要通知</p>
      </div>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="announcements.length === 0" text="暂无公告" icon="📢" />

      <div v-else class="ann-list">
        <article v-for="a in announcements" :key="a.id" class="card">
          <div class="row-between">
            <h3 class="ann-title">{{ a.title }}</h3>
            <span class="muted text-sm">{{ formatTime(a.created_at) }}</span>
          </div>
          <p class="muted text-sm">发布人：{{ a.author_name || ("管理员 " + a.admin_id) }}</p>
          <p class="ann-content">{{ a.content }}</p>
        </article>
      </div>

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
import { listAnnouncements } from "@/api/announcement";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const announcements = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchList() {
  loading.value = true;
  try {
    const data = await listAnnouncements({ page: page.value, page_size: pageSize.value });
    announcements.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    announcements.value = [];
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

function onPageSizeChange(size) {
  pageSize.value = size;
  page.value = 1;
  fetchList();
}

onMounted(fetchList);
</script>

<style scoped>
.ann-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ann-title {
  font-size: 16px;
  margin: 0;
}
.ann-content {
  white-space: pre-wrap;
  margin: 10px 0 0;
  font-size: 14px;
}
</style>