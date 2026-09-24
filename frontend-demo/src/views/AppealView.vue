<!--
  账号申诉页（公开，无需登录）。
  用于“自行注销反悔 / 被误封请求撤回 / 其他”三种情形；reason=other 时说明必填。
-->
<template>
  <div class="appeal-wrap">
    <div class="card appeal-card">
      <h2>账号申诉</h2>
      <p class="muted text-sm">
        若你的账号已被注销或封禁、无法登录，可在此提交申诉，等待超级管理员审核。
      </p>

      <div v-if="success" class="alert alert-info mt-16">
        申诉已提交（编号 {{ createdId }}），请耐心等待管理员审核。
      </div>
      <div v-if="errorMsg" class="alert alert-error mt-16">{{ errorMsg }}</div>

      <form class="mt-16" @submit.prevent="handleSubmit">
        <div class="field">
          <label>被注销 / 封禁的学号（纯数字）</label>
          <input v-model.trim="form.username" class="input" placeholder="如 20240001" />
        </div>
        <div class="field">
          <label>申诉原因</label>
          <select v-model="form.reason" class="select">
            <option value="self_regret">自行注销后反悔</option>
            <option value="wrongful_ban">被误封，请求撤回</option>
            <option value="other">其他</option>
          </select>
        </div>
        <div class="field">
          <label>
            补充说明
            <span v-if="form.reason === 'other'" class="required">（必填）</span>
          </label>
          <textarea
            v-model.trim="form.content"
            class="textarea"
            maxlength="1000"
            placeholder="请说明具体情况，便于管理员核实"
          ></textarea>
        </div>
        <button class="btn btn-block" :disabled="loading" type="submit">
          {{ loading ? "提交中…" : "提交申诉" }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { submitAppeal } from "@/api/appeal";

const form = reactive({ username: "", reason: "self_regret", content: "" });
const loading = ref(false);
const errorMsg = ref("");
const success = ref(false);
const createdId = ref("");

async function handleSubmit() {
  errorMsg.value = "";
  success.value = false;
  if (!/^\d+$/.test(form.username)) {
    errorMsg.value = "学号须为纯数字";
    return;
  }
  if (form.reason === "other" && !form.content) {
    errorMsg.value = "选择“其他”时必须填写补充说明";
    return;
  }
  loading.value = true;
  try {
    const data = await submitAppeal({
      username: form.username,
      reason: form.reason,
      content: form.content,
    });
    success.value = true;
    createdId.value = data?.id ?? "";
    form.content = "";
  } catch (e) {
    errorMsg.value = e?.msg || "提交失败，请稍后重试";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.appeal-wrap {
  display: flex;
  justify-content: center;
  padding-top: 24px;
}
.appeal-card {
  width: 100%;
  max-width: 520px;
}
.required {
  color: var(--color-danger);
}
</style>
