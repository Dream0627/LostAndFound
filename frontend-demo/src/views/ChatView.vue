<!--
  会话聊天页：消息列表 + 发送 + 完成寻找申请（发起/处理）。
  说明：后端消息按 id 倒序返回（新消息在前），这里反转后按时间升序展示。
-->
<template>
  <div class="chat-wrap">
    <div class="chat-header card">
      <div class="row-between">
        <div class="row">
          <button class="btn btn-ghost btn-sm" @click="$router.push('/conversations')">← 消息</button>
          <h3 class="chat-title">会话 #{{ conversationId }}</h3>
        </div>
        <div class="row">
          <button
            v-if="!pendingRequest"
            class="btn btn-sm"
            :disabled="finishLoading"
            @click="handleCreateFinishRequest"
          >
            发起完成寻找
          </button>
        </div>
      </div>

      <div v-if="pendingRequest" class="finish-banner mt-8">
        <template v-if="pendingRequest.requester_id === auth.user?.id">
          <span>你已发起完成寻找申请，等待对方处理…</span>
          <button class="btn btn-danger btn-sm" :disabled="finishLoading" @click="handleHandle('rejected')">
            撤回（拒绝）
          </button>
        </template>
        <template v-else>
          <span>对方申请完成寻找，是否同意？同意后帖子将置为已完成。</span>
          <div class="row">
            <button class="btn btn-sm" :disabled="finishLoading" @click="handleHandle('agreed')">同意</button>
            <button class="btn btn-ghost btn-sm" :disabled="finishLoading" @click="handleHandle('rejected')">
              拒绝
            </button>
          </div>
        </template>
      </div>
    </div>

    <div ref="listRef" class="chat-body">
      <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>
      <EmptyState v-else-if="messages.length === 0" text="还没有消息，打个招呼吧" icon="👋" />

      <ul v-else class="msg-list">
        <li v-for="m in orderedMessages" :key="m.id" class="msg" :class="{ mine: isMine(m) }">
          <div class="msg-bubble">
            <div class="msg-text">{{ m.content }}</div>
            <div class="msg-time">{{ formatTime(m.created_at) }}</div>
          </div>
        </li>
      </ul>
    </div>

    <div class="chat-input card">
      <textarea
        v-model.trim="draft"
        class="textarea"
        maxlength="1000"
        placeholder="输入消息，Enter 发送（Shift+Enter 换行）"
        @keydown.enter.exact.prevent="handleSend"
      ></textarea>
      <div class="row-between mt-8">
        <span class="muted text-sm">{{ draft.length }} / 1000</span>
        <button class="btn" :disabled="sending || !draft" @click="handleSend">
          {{ sending ? "发送中…" : "发送" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import {
  listMessages,
  sendMessage,
  createFinishRequest,
  handleFinishRequest,
} from "@/api/conversation";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";
import EmptyState from "@/components/EmptyState.vue";

const route = useRoute();
const auth = useAuthStore();
const toast = useToastStore();

const conversationId = route.params.id;
const messages = ref([]);
const loading = ref(false);
const sending = ref(false);
const draft = ref("");
const pendingRequest = ref(null); // 最近的待处理完成申请（如有）
const finishLoading = ref(false);
const listRef = ref(null);

// 后端返回新消息在前，反转成时间升序（旧→新）更符合聊天直觉。
const orderedMessages = computed(() => [...messages.value].reverse());

function isMine(m) {
  return m.sender_id === auth.user?.id;
}

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchMessages(scrollToBottom = false) {
  loading.value = true;
  try {
    const data = await listMessages(conversationId, { page: 1, page_size: 50 });
    messages.value = data.list || [];
    if (scrollToBottom) {
      await nextTick();
      if (listRef.value) listRef.value.scrollTop = listRef.value.scrollHeight;
    }
  } catch (e) {
    messages.value = [];
    toast.error(e?.msg || "加载消息失败");
  } finally {
    loading.value = false;
  }
}

async function handleSend() {
  if (!draft.value) return;
  sending.value = true;
  try {
    await sendMessage(conversationId, draft.value);
    draft.value = "";
    await fetchMessages(true);
  } catch (e) {
    toast.error(e?.msg || "发送失败");
  } finally {
    sending.value = false;
  }
}

async function handleCreateFinishRequest() {
  finishLoading.value = true;
  try {
    const req = await createFinishRequest(conversationId);
    pendingRequest.value = req;
    toast.success("已发起完成寻找申请");
  } catch (e) {
    toast.error(e?.msg || "发起失败");
  } finally {
    finishLoading.value = false;
  }
}

async function handleHandle(status) {
  if (!pendingRequest.value) return;
  finishLoading.value = true;
  try {
    await handleFinishRequest(conversationId, pendingRequest.value.id, status);
    toast.success(status === "agreed" ? "已同意，帖子已完成" : "已拒绝");
    pendingRequest.value = null;
    fetchMessages();
  } catch (e) {
    toast.error(e?.msg || "处理失败");
  } finally {
    finishLoading.value = false;
  }
}

onMounted(() => fetchMessages(true));
</script>

<style scoped>
.chat-wrap {
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: calc(100vh - 200px);
  min-height: 460px;
}
.chat-header {
  padding: 14px 16px;
}
.chat-title {
  margin: 0;
}
.finish-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  background: var(--color-warning-soft);
  color: var(--color-warning);
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  font-size: 14px;
}
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px 4px;
}
.msg-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.msg {
  display: flex;
}
.msg.mine {
  justify-content: flex-end;
}
.msg-bubble {
  max-width: 70%;
  padding: 9px 13px;
  border-radius: 14px;
  background: #fff;
  border: 1px solid var(--color-border);
}
.msg.mine .msg-bubble {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.msg-text {
  white-space: pre-wrap;
  word-break: break-word;
}
.msg-time {
  font-size: 11px;
  opacity: 0.7;
  margin-top: 4px;
  text-align: right;
}
.chat-input {
  padding: 12px;
}
.chat-input .textarea {
  min-height: 70px;
}
</style>
