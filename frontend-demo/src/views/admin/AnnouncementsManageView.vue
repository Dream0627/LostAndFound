<!--
  公告管理（postadmin/mainadmin）：发布公告 + 列表 + 删除（软删除）。
-->
<template>
  <div>
    <h2>公告管理</h2>
    <p class="muted text-sm">发布全站公告；公告会在广场「公告」入口中展示。</p>

    <div class="card">
      <h3>发布公告</h3>
      <div class="field">
        <label>标题（1–200 字）</label>
        <input v-model.trim="form.title" class="input" maxlength="200" placeholder="如：图书馆失物认领安排" />
      </div>
      <div class="field">
        <label>内容（1–2000 字）</label>
        <textarea v-model.trim="form.content" class="textarea" rows="5" maxlength="2000" placeholder="公告正文…"></textarea>
      </div>
      <button class="btn" :disabled="publishing || !form.title || !form.content" @click="handlePublish">
        {{ publishing ? "发布中…" : "发布公告" }}
      </button>
    </div>

    <div class="card mt-16">
      <h3>已发布公告（{{ total }}）</h3>
      <EmptyState v-if="announcements.length === 0" text="暂无公告" icon="📢" />
      <ul v-else class="ann-manage-list">
        <li v-for="a in announcements" :key="a.id" class="ann-item">
          <div class="row-between">
            <div>
              <div class="ann-title">{{ a.title }}</div>
              <div class="muted text-sm">{{ formatTime(a.created_at) }}</div>
            </div>
            <button class="btn btn-danger btn-sm" @click="handleDelete(a)">删除</button>
          </div>
          <p class="ann-content">{{ a.content }}</p>
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
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import { listAnnouncements } from "@/api/announcement";
import { createAnnouncement, deleteAnnouncement } from "@/api/admin";
import { useToastStore } from "@/stores/toast";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const toast = useToastStore();

const announcements = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const publishing = ref(false);
const form = reactive({ title: "", content: "" });

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchList() {
  try {
    const data = await listAnnouncements({ page: page.value, page_size: pageSize.value });
    announcements.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    announcements.value = [];
    total.value = 0;
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

async function handlePublish() {
  publishing.value = true;
  try {
    await createAnnouncement({ title: form.title, content: form.content });
    toast.success("公告已发布");
    form.title = "";
    form.content = "";
    page.value = 1;
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "发布失败");
  } finally {
    publishing.value = false;
  }
}

async function handleDelete(a) {
  if (!window.confirm(`确定删除公告「${a.title}」？`)) return;
  try {
    await deleteAnnouncement(a.id);
    toast.success("公告已删除");
    fetchList();
  } catch (e) {
    toast.error(e?.msg || "删除失败");
  }
}

onMounted(fetchList);
</script>

<style scoped>
.ann-manage-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.ann-item {
  padding: 12px 0;
  border-top: 1px solid var(--color-border);
}
.ann-item:first-child {
  border-top: none;
}
.ann-title {
  font-weight: 600;
}
.ann-content {
  white-space: pre-wrap;
  margin: 8px 0 0;
  font-size: 14px;
}
</style>