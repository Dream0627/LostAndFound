<!-- 404 页 -->
<template>
  <div class="nf">
    <div class="nf-code">404</div>
    <p class="muted">你访问的页面不存在或已被删除。</p>
    <router-link to="/posts" class="btn">返回帖子广场</router-link>
  </div>
</template>

<style scoped>
.nf {
  text-align: center;
  padding: 80px 20px;
}
.nf-code {
  font-size: 64px;
  font-weight: 800;
  color: var(--color-primary);
}
</style>
