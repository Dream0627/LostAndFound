<!--
  顶部导航栏：
    - 左侧品牌；
    - 右侧按登录态与角色显隐入口（广场 / 发布 / 消息 / 我的 / 管理后台 / 登录注册）。
  管理员（postadmin/mainadmin）额外显示“管理后台”。
-->
<template>
  <header class="navbar">
    <div class="nav-inner">
      <router-link to="/posts" class="brand">
        <span class="brand-logo">🎒</span>
        <span class="brand-name">LAF 失物招领</span>
      </router-link>

      <nav class="nav-links">
        <router-link to="/posts" class="nav-link">广场</router-link>

        <template v-if="auth.isLoggedIn">
          <router-link to="/posts/new" class="nav-link">发布</router-link>
          <router-link to="/conversations" class="nav-link">消息</router-link>
          <router-link v-if="auth.isPostAdmin" to="/admin" class="nav-link">管理后台</router-link>
        </template>
      </nav>

      <div class="nav-user">
        <template v-if="auth.isLoggedIn">
          <router-link to="/me" class="user-chip">
            <span class="avatar">{{ initial }}</span>
            <span class="user-name">{{ auth.user?.name }}</span>
            <span class="role-badge" :class="roleClass">{{ roleLabel }}</span>
          </router-link>
          <button class="btn btn-ghost btn-sm" @click="handleLogout">退出</button>
        </template>
        <template v-else>
          <router-link to="/login" class="btn btn-ghost btn-sm">登录</router-link>
          <router-link to="/register" class="btn btn-sm">注册</router-link>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();

const initial = computed(() => (auth.user?.name || "?").slice(0, 1));
const roleLabel = computed(
  () => ({ student: "学生", postadmin: "帖子管理员", mainadmin: "超级管理员" })[auth.role] || auth.role
);
const roleClass = computed(() => {
  if (auth.role === "mainadmin") return "role-main";
  if (auth.role === "postadmin") return "role-post";
  return "role-student";
});

function handleLogout() {
  auth.logout();
  toast.success("已退出登录");
  router.push({ name: "post-list" });
}
</script>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--color-border);
}
.nav-inner {
  max-width: var(--maxw);
  margin: 0 auto;
  padding: 0 16px;
  height: 62px;
  display: flex;
  align-items: center;
  gap: 20px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: var(--color-text);
  font-size: 17px;
}
.brand:hover {
  text-decoration: none;
}
.brand-logo {
  font-size: 22px;
}
.nav-links {
  display: flex;
  gap: 6px;
  flex: 1;
}
.nav-link {
  padding: 6px 12px;
  border-radius: 999px;
  color: var(--color-text-muted);
  font-size: 14px;
}
.nav-link:hover {
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  text-decoration: none;
}
.nav-link.router-link-active {
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  font-weight: 600;
}
.nav-user {
  display: flex;
  align-items: center;
  gap: 10px;
}
.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 999px;
  border: 1px solid var(--color-border);
  color: var(--color-text);
}
.user-chip:hover {
  text-decoration: none;
  border-color: var(--color-primary);
}
.avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}
.user-name {
  font-size: 14px;
}
.role-badge {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: 999px;
}
.role-student {
  background: #eef1f4;
  color: #6b7684;
}
.role-post {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}
.role-main {
  background: #fde8ef;
  color: #c02a6a;
}
</style>
