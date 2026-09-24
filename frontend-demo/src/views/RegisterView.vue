<!-- 注册页：学号(纯数字) + 姓名 + 密码(8-16) + 确认密码 + 角色。 -->
<template>
  <div class="auth-wrap">
    <div class="card auth-card">
      <h2 class="auth-title">创建账号 ✨</h2>
      <p class="muted text-sm">学号用于登录，须为纯数字。</p>

      <div v-if="errorMsg" class="alert alert-error mt-16">{{ errorMsg }}</div>

      <form class="mt-16" @submit.prevent="handleSubmit">
        <div class="field">
          <label>学号 / 工号（纯数字）</label>
          <input v-model.trim="form.username" class="input" placeholder="如 20240001" />
        </div>
        <div class="field">
          <label>姓名</label>
          <input v-model.trim="form.name" class="input" placeholder="请输入姓名" />
        </div>
        <div class="field">
          <label>密码（8–16 位）</label>
          <input v-model="form.password" type="password" class="input" placeholder="请输入密码" />
        </div>
        <div class="field">
          <label>确认密码</label>
          <input v-model="form.confirm" type="password" class="input" placeholder="再次输入密码" />
        </div>
        <div class="field">
          <label>角色</label>
          <select v-model="form.role" class="select">
            <option value="student">学生</option>
            <option value="postadmin">帖子管理员</option>
            <option value="mainadmin">超级管理员</option>
          </select>
        </div>
        <button class="btn btn-block" :disabled="loading" type="submit">
          {{ loading ? "注册中…" : "注册" }}
        </button>
      </form>

      <p class="auth-foot text-sm">
        已有账号？<router-link to="/login">去登录</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { register } from "@/api/auth";
import { useToastStore } from "@/stores/toast";

const toast = useToastStore();
const router = useRouter();

const form = reactive({ username: "", name: "", password: "", confirm: "", role: "student" });
const loading = ref(false);
const errorMsg = ref("");

function validate() {
  if (!/^\d+$/.test(form.username)) return "学号须为纯数字";
  if (!form.name) return "请输入姓名";
  if (form.password.length < 8 || form.password.length > 16) return "密码须为 8–16 位";
  if (form.password !== form.confirm) return "两次输入的密码不一致";
  return "";
}

async function handleSubmit() {
  errorMsg.value = validate();
  if (errorMsg.value) return;
  loading.value = true;
  try {
    await register({
      username: form.username,
      name: form.name,
      password: form.password,
      role: form.role,
    });
    toast.success("注册成功，请登录");
    router.push({ name: "login" });
  } catch (e) {
    errorMsg.value = e?.msg || "注册失败，请稍后重试";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-wrap {
  display: flex;
  justify-content: center;
  padding-top: 32px;
}
.auth-card {
  width: 100%;
  max-width: 420px;
}
.auth-title {
  margin-bottom: 4px;
}
.auth-foot {
  margin-top: 18px;
  text-align: center;
  color: var(--color-text-muted);
}
</style>
