<!--
  通用分页器（直观版）：首页/上一页 + 页码按钮组 + 下一页/末页 + 每页数量选择 + 跳页输入。
  props 受控；emit change(page) 切换页码，emit changePageSize(size) 修改每页数量。
  页码过多时用省略号折叠，始终保留首尾页与当前页附近窗口。
-->
<template>
  <div v-if="total > 0" class="pager">
    <div class="pager-main">
      <button class="btn btn-ghost btn-sm" :disabled="page <= 1" @click="go(1)">首页</button>
      <button class="btn btn-ghost btn-sm" :disabled="page <= 1" @click="go(page - 1)">上一页</button>

      <button
        v-for="(item, idx) in pageItems"
        :key="idx"
        class="page-btn"
        :class="{ active: item === page, ellipsis: item === '...' }"
        :disabled="item === '...'"
        @click="item !== '...' && go(item)"
      >
        {{ item }}
      </button>

      <button class="btn btn-ghost btn-sm" :disabled="page >= totalPages" @click="go(page + 1)">下一页</button>
      <button class="btn btn-ghost btn-sm" :disabled="page >= totalPages" @click="go(totalPages)">末页</button>
    </div>

    <div class="pager-side">
      <span class="pager-info">共 {{ total }} 条 · 第 {{ page }}/{{ totalPages }} 页</span>

      <label class="pager-field">
        每页
        <select class="page-size-select" :value="pageSize" @change="onSizeChange">
          <option v-for="n in sizeOptions" :key="n" :value="n">{{ n }}</option>
        </select>
        条
      </label>

      <label class="pager-field">
        跳至
        <input
          v-model.number="jumpTo"
          class="page-jump"
          type="number"
          min="1"
          :max="totalPages"
          @keydown.enter="doJump"
        />
        页
        <button class="btn btn-ghost btn-sm" @click="doJump">Go</button>
      </label>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
  page: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 },
  total: { type: Number, default: 0 },
});
const emit = defineEmits(["change", "changePageSize"]);

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)));

// 每页数量候选：固定常用值，若当前值不在其中则并入并排序，避免选择框显示空白。
const sizeOptions = computed(() => {
  const base = [10, 20, 50, 100];
  return base.includes(props.pageSize) ? base : [...base, props.pageSize].sort((a, b) => a - b);
});

// 页码按钮组：始终显示首尾页，当前页左右各保留 1 页，跨度较大处用省略号折叠。
const pageItems = computed(() => {
  const last = totalPages.value;
  const cur = props.page;
  if (last <= 7) return Array.from({ length: last }, (_, i) => i + 1);

  const items = [1];
  const start = Math.max(2, cur - 1);
  const end = Math.min(last - 1, cur + 1);
  if (start > 2) items.push("...");
  for (let i = start; i <= end; i++) items.push(i);
  if (end < last - 1) items.push("...");
  items.push(last);
  return items;
});

const jumpTo = ref(props.page);
watch(
  () => props.page,
  (p) => {
    jumpTo.value = p;
  }
);

function go(target) {
  if (target < 1 || target > totalPages.value || target === props.page) return;
  emit("change", target);
}

function doJump() {
  const target = Number(jumpTo.value);
  if (!Number.isInteger(target) || target < 1 || target > totalPages.value) {
    jumpTo.value = props.page; // 非法输入还原为当前页
    return;
  }
  go(target);
}

function onSizeChange(e) {
  emit("changePageSize", Number(e.target.value));
}
</script>

<style scoped>
.pager {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 20px;
}
.pager-main {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.page-btn {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: #fff;
  color: var(--color-text);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.12s ease;
}
.page-btn:hover:not(.active):not(.ellipsis) {
  border-color: var(--color-primary);
  color: var(--color-primary-dark);
}
.page-btn.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
  font-weight: 600;
}
.page-btn.ellipsis {
  border: none;
  background: transparent;
  cursor: default;
  color: var(--color-text-muted);
}
.pager-side {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}
.pager-info {
  font-size: 13px;
  color: var(--color-text-muted);
}
.pager-field {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  color: var(--color-text-muted);
}
.page-size-select,
.page-jump {
  height: 30px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0 6px;
  font-size: 13px;
  color: var(--color-text);
  background: #fff;
}
.page-jump {
  width: 58px;
}
</style>