<!-- 登录页：用户名(学号) + 密码；成功后跳回来源页或帖子广场。 -->
<template>
  <div class="auth-wrap">
    <div class="card auth-card">
      <h2 class="auth-title">欢迎回来 👋</h2>
      <p class="muted text-sm">登录后即可发布帖子、参与对话与评论。</p>

      <div v-if="errorMsg" class="alert alert-error mt-16">{{ errorMsg }}</div>

      <form class="mt-16" @submit.prevent="handleSubmit">
        <div class="field">
          <label>学号 / 工号</label>
          <input v-model.trim="form.username" class="input" placeholder="请输入纯数字学号" />
        </div>
        <div class="field">
          <label>密码</label>
          <input v-model="form.password" type="password" class="input" placeholder="请输入密码" />
        </div>
        <button class="btn btn-block" :disabled="loading" type="submit">
          {{ loading ? "登录中…" : "登录" }}
        </button>
      </form>

      <p class="auth-foot text-sm">
        还没有账号？<router-link to="/register">立即注册</router-link>
        <span class="sep">·</span>
        <router-link to="/appeal">账号被注销？去申诉</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const route = useRoute();

const form = reactive({ username: "", password: "" });
const loading = ref(false);
const errorMsg = ref("");

async function handleSubmit() {
  errorMsg.value = "";
  if (!form.username || !form.password) {
    errorMsg.value = "请输入学号与密码";
    return;
  }
  loading.value = true;
  try {
    const user = await auth.login({ username: form.username, password: form.password });
    toast.success(`登录成功，欢迎 ${user.name}`);
    const redirect = route.query.redirect || "/posts";
    router.replace(redirect);
  } catch (e) {
    errorMsg.value = e?.msg || "登录失败，请检查账号密码";
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
  max-width: 400px;
}
.auth-title {
  margin-bottom: 4px;
}
.auth-foot {
  margin-top: 18px;
  text-align: center;
  color: var(--color-text-muted);
}
.auth-foot .sep {
  margin: 0 8px;
  color: var(--color-border);
}
</style>
