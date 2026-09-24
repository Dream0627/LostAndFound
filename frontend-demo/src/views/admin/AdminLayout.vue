<!--
  管理后台布局：左侧子导航 + 右侧子路由出口。
  仅 postadmin / mainadmin 可进入（路由守卫已限制）；申诉审核与用户管理仅 mainadmin 可见。
-->
<template>
  <div class="admin">
    <aside class="admin-side card">
      <h3 class="admin-side-title">管理后台</h3>
      <nav class="admin-nav">
        <router-link to="/admin/reviews" class="admin-link">待审批</router-link>
        <router-link to="/admin/deleted-posts" class="admin-link">已删帖子</router-link>
        <router-link v-if="auth.isMainAdmin" to="/admin/appeals" class="admin-link">申诉审核</router-link>
        <router-link v-if="auth.isMainAdmin" to="/admin/users" class="admin-link">用户管理</router-link>
      </nav>
      <p class="muted text-sm admin-role">
        当前角色：{{ roleLabel }}
        <span v-if="!auth.isMainAdmin">（超级管理员专属入口已隐藏）</span>
      </p>
    </aside>

    <section class="admin-main">
      <router-view />
    </section>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const roleLabel = computed(
  () => ({ postadmin: "帖子管理员", mainadmin: "超级管理员" })[auth.role] || auth.role
);
</script>

<style scoped>
.admin {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 18px;
  align-items: start;
}
@media (max-width: 720px) {
  .admin {
    grid-template-columns: 1fr;
  }
}
.admin-side {
  position: sticky;
  top: 78px;
}
.admin-side-title {
  font-size: 15px;
}
.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.admin-link {
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-size: 14px;
}
.admin-link:hover {
  background: #f4f6f9;
  text-decoration: none;
}
.admin-link.router-link-active {
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  font-weight: 600;
}
.admin-role {
  margin-top: 14px;
}
</style>
