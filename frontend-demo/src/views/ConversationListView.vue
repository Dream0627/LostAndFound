<!--
  我的消息：会话列表（发起方或楼主参与的会话），点击进入聊天。
-->
<template>
  <div>
    <div class="page-head">
      <h2>我的消息</h2>
    </div>

    <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

    <template v-else>
      <EmptyState v-if="conversations.length === 0" text="暂无会话，去帖子详情页发起申领/召领吧" icon="✉️">
        <router-link to="/posts" class="btn">去逛广场</router-link>
      </EmptyState>

      <ul v-else class="conv-list">
        <li
          v-for="c in conversations"
          :key="c.id"
          class="conv-item"
          @click="$router.push(`/conversations/${c.id}`)"
        >
          <div class="conv-avatar">#{{ c.id }}</div>
          <div class="conv-body">
            <div class="conv-title">
              帖子 #{{ c.post_id }}
              <span class="muted text-sm">
                · 我{{ c.initiator_id === auth.user?.id ? "发起的申领/召领" : "是楼主" }}
              </span>
            </div>
            <div class="muted text-sm">发起人 ID：{{ c.initiator_id }} · 楼主 ID：{{ c.owner_id }}</div>
          </div>
          <div class="conv-time muted text-sm">{{ formatTime(c.created_at) }}</div>
        </li>
      </ul>

      <Pagination :page="page" :page-size="pageSize" :total="total" @change="onPageChange" />
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listMyConversations } from "@/api/conversation";
import { useAuthStore } from "@/stores/auth";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";

const auth = useAuthStore();
const conversations = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const loading = ref(false);

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchList() {
  loading.value = true;
  try {
    const data = await listMyConversations({ page: page.value, page_size: pageSize.value });
    conversations.value = data.list || [];
    total.value = data.total || 0;
  } catch (e) {
    conversations.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

function onPageChange(p) {
  page.value = p;
  fetchList();
}

onMounted(fetchList);
</script>

<style scoped>
.page-head {
  margin-bottom: 16px;
}
.conv-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.conv-item {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 14px 16px;
  cursor: pointer;
  transition: border-color 0.12s ease, box-shadow 0.12s ease;
}
.conv-item:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}
.conv-avatar {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}
.conv-body {
  flex: 1;
}
.conv-title {
  font-weight: 600;
}
</style>
