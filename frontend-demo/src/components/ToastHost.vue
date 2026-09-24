<!-- 全局轻提示容器：固定在右上角，读取 toast store 渲染。 -->
<template>
  <div class="toast-host">
    <transition-group name="toast">
      <div v-for="t in toast.items" :key="t.id" class="toast" :class="t.type">
        {{ t.text }}
      </div>
    </transition-group>
  </div>
</template>

<script setup>
import { useToastStore } from "@/stores/toast";
const toast = useToastStore();
</script>
