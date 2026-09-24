<!-- 通用分页条：上一页/下一页 + 页码信息。props 受控，emit change(page)。 -->
<template>
  <div v-if="total > 0" class="pager">
    <button class="btn btn-ghost btn-sm" :disabled="page <= 1" @click="go(page - 1)">上一页</button>
    <span class="pager-info">第 {{ page }} / {{ totalPages }} 页 · 共 {{ total }} 条</span>
    <button class="btn btn-ghost btn-sm" :disabled="page >= totalPages" @click="go(page + 1)">
      下一页
    </button>
  </div>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  page: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 },
  total: { type: Number, default: 0 },
});
const emit = defineEmits(["change"]);

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)));

function go(target) {
  if (target < 1 || target > totalPages.value || target === props.page) return;
  emit("change", target);
}
</script>

<style scoped>
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  margin-top: 20px;
}
.pager-info {
  font-size: 13px;
  color: var(--color-text-muted);
}
</style>
