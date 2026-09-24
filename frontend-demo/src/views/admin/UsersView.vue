<!--
  用户管理（仅 mainadmin）：按用户 ID 执行“注销 / 恢复”。
  后端未提供用户列表接口，故这里以 ID 输入为入口，配合危险操作二次确认。
-->
<template>
  <div>
    <h2>用户管理</h2>
    <p class="muted text-sm">
      按用户 ID 执行注销或恢复。注销会软删除该用户及其名下帖子/评论；恢复仅恢复与其“同批删除”的内容。
    </p>

    <div class="card">
      <div class="field">
        <label>用户 ID</label>
        <input v-model.trim="userId" class="input" placeholder="请输入用户 ID（正整数）" @keydown.enter="handleRecover" />
      </div>
      <div class="row">
        <button class="btn btn-danger" :disabled="busy || !validId" @click="handleDelete">
          注销该用户
        </button>
        <button class="btn" :disabled="busy || !validId" @click="handleRecover">
          恢复该用户
        </button>
      </div>
    </div>

    <div class="card">
      <h3>操作日志</h3>
      <EmptyState v-if="logs.length === 0" text="暂无操作记录" icon="🧾" />
      <ul v-else class="log-list">
        <li v-for="(l, i) in logs" :key="i" class="log-item" :class="l.ok ? 'ok' : 'err'">
          <span class="log-time">{{ l.time }}</span>
          <span>{{ l.text }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { deleteUser, recoverUser } from "@/api/admin";
import { useToastStore } from "@/stores/toast";
import EmptyState from "@/components/EmptyState.vue";

const toast = useToastStore();
const userId = ref("");
const busy = ref(false);
const logs = ref([]);

const validId = computed(() => /^\d+$/.test(userId.value) && Number(userId.value) > 0);

function pushLog(text, ok) {
  logs.value.unshift({
    text,
    ok,
    time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
  });
}

async function handleDelete() {
  const id = Number(userId.value);
  if (!window.confirm(`确定注销用户 #${id}？该操作会软删除其名下内容。`)) return;
  busy.value = true;
  try {
    await deleteUser(id);
    toast.success("用户已注销");
    pushLog(`注销用户 #${id} 成功`, true);
  } catch (e) {
    toast.error(e?.msg || "注销失败");
    pushLog(`注销用户 #${id} 失败：${e?.msg || "未知错误"}`, false);
  } finally {
    busy.value = false;
  }
}

async function handleRecover() {
  const id = Number(userId.value);
  if (!validId.value) return;
  busy.value = true;
  try {
    await recoverUser(id);
    toast.success("用户已恢复");
    pushLog(`恢复用户 #${id} 成功`, true);
  } catch (e) {
    toast.error(e?.msg || "恢复失败");
    pushLog(`恢复用户 #${id} 失败：${e?.msg || "未知错误"}`, false);
  } finally {
    busy.value = false;
  }
}
</script>

<style scoped>
.log-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.log-item {
  display: flex;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
}
.log-item:last-child {
  border-bottom: none;
}
.log-item.ok {
  color: var(--color-primary-dark);
}
.log-item.err {
  color: var(--color-danger);
}
.log-time {
  color: var(--color-text-muted);
  white-space: nowrap;
}
</style>
