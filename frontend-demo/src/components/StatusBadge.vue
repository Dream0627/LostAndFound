<!-- 状态/类型徽章：统一渲染帖子类型(lost/found)、审核状态、完成状态。 -->
<template>
  <span class="tag" :class="cls">{{ text }}</span>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  kind: { type: String, required: true }, // 'type' | 'status' | 'finished'
  value: { type: String, required: true }, // 具体取值
});

const MAP = {
  type: {
    lost: { text: "失物", cls: "tag-lost" },
    found: { text: "招领", cls: "tag-found" },
  },
  status: {
    pending: { text: "待审核", cls: "tag-pending" },
    approved: { text: "已通过", cls: "tag-approved" },
    rejected: { text: "已驳回", cls: "tag-rejected" },
  },
};

const text = computed(() => {
  if (props.kind === "finished") return props.value ? "已完成" : "进行中";
  return MAP[props.kind]?.[props.value]?.text || props.value;
});
const cls = computed(() => {
  if (props.kind === "finished") return props.value ? "tag-finished" : "tag-open";
  return MAP[props.kind]?.[props.value]?.cls || "";
});
</script>
