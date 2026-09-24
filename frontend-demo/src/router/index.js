// 路由表：Vue Router 4，history 模式。
// 通过 meta.requiresAuth / meta.roles 控制访问；守卫里统一校验登录态与角色。
import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes = [
  { path: "/", redirect: "/posts" },
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/LoginView.vue"),
    meta: { guestOnly: true, title: "登录" },
  },
  {
    path: "/register",
    name: "register",
    component: () => import("@/views/RegisterView.vue"),
    meta: { guestOnly: true, title: "注册" },
  },
  {
    path: "/appeal",
    name: "appeal",
    component: () => import("@/views/AppealView.vue"),
    meta: { title: "账号申诉" },
  },
  {
    path: "/posts",
    name: "post-list",
    component: () => import("@/views/PostListView.vue"),
    meta: { title: "帖子广场" },
  },
  {
    path: "/posts/new",
    name: "post-create",
    component: () => import("@/views/PostCreateView.vue"),
    meta: { requiresAuth: true, title: "发布帖子" },
  },
  {
    path: "/posts/:id",
    name: "post-detail",
    component: () => import("@/views/PostDetailView.vue"),
    meta: { title: "帖子详情" },
  },
  {
    path: "/me",
    name: "profile",
    component: () => import("@/views/ProfileView.vue"),
    meta: { requiresAuth: true, title: "我的" },
  },
  {
    path: "/conversations",
    name: "conversation-list",
    component: () => import("@/views/ConversationListView.vue"),
    meta: { requiresAuth: true, title: "我的消息" },
  },
  {
    path: "/conversations/:id",
    name: "conversation-chat",
    component: () => import("@/views/ChatView.vue"),
    meta: { requiresAuth: true, title: "会话" },
  },
  {
    path: "/admin",
    component: () => import("@/views/admin/AdminLayout.vue"),
    meta: { requiresAuth: true, roles: ["postadmin", "mainadmin"] },
    children: [
      { path: "", redirect: "/admin/reviews" },
      {
        path: "reviews",
        name: "admin-reviews",
        component: () => import("@/views/admin/ReviewsView.vue"),
        meta: { title: "待审批" },
      },
      {
        path: "deleted-posts",
        name: "admin-deleted-posts",
        component: () => import("@/views/admin/DeletedPostsView.vue"),
        meta: { title: "已删帖子" },
      },
      {
        path: "appeals",
        name: "admin-appeals",
        component: () => import("@/views/admin/AppealsView.vue"),
        meta: { roles: ["mainadmin"], title: "申诉审核" },
      },
      {
        path: "users",
        name: "admin-users",
        component: () => import("@/views/admin/UsersView.vue"),
        meta: { roles: ["mainadmin"], title: "用户管理" },
      },
    ],
  },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/views/NotFoundView.vue"),
    meta: { title: "页面不存在" },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
});

// 全局前置守卫：
//  - requiresAuth 且未登录 → 跳登录页并记住来源；
//  - roles 限制且当前角色不在白名单 → 回帖子广场；
//  - guestOnly（登录/注册页）已登录 → 回帖子广场。
router.beforeEach((to) => {
  const auth = useAuthStore();
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  if (Array.isArray(to.meta.roles) && to.meta.roles.length > 0) {
    if (!auth.isLoggedIn || !to.meta.roles.includes(auth.user?.role)) {
      return { name: "post-list" };
    }
  }
  if (to.meta.guestOnly && auth.isLoggedIn) {
    return { name: "post-list" };
  }
  return true;
});

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · LAF 失物招领` : "LAF 校园失物招领";
});

export default router;
