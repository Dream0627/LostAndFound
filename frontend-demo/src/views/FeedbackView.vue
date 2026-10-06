<!--
  意见反馈：登录用户向平台提交反馈；超级管理员在后台（admin/feedbacks）审批。
-->
<template>
  <div>
    <h2>意见反馈</h2>
    <p class="muted text-sm">把你的建议或问题告诉我们，超级管理员会在后台查看并处理。</p>

    <div class="card mt-16">
      <div class="field">
        <label>反馈内容（1–1000 字）</label>
        <textarea
          v-model.trim="content"
          class="textarea"
          rows="6"
          maxlength="1000"
          placeholder="描述你遇到的问题或建议…"
        ></textarea>
      </div>
      <div class="row-between mt-8">
        <span class="muted text-sm">{{ content.length }} / 1000</span>
        <button class="btn" :disabled="submitting || !content" @click="handleSubmit">
          {{ submitting ? "提交中…" : "提交反馈" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { submitFeedback } from "@/api/feedback";
import { useToastStore } from "@/stores/toast";

const toast = useToastStore();
const content = ref("");
const submitting = ref(false);

async function handleSubmit() {
  submitting.value = true;
  try {
    await submitFeedback(content.value);
    toast.success("反馈已提交，感谢你的建议");
    content.value = "";
  } catch (e) {
    toast.error(e?.msg || "提交失败");
  } finally {
    submitting.value = false;
  }
}
</script>