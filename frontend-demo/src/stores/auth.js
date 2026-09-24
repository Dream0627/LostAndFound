// 认证与当前用户状态。
// 负责登录/登出、token 持久化，并暴露角色判断（便于界面按角色显隐管理员入口）。
import { defineStore } from "pinia";
import { tokenStore } from "@/api/http";
import * as authApi from "@/api/auth";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: tokenStore.get(),
    // 从 localStorage 恢复上次的 user，刷新页面不至于闪一下未登录。
    user: JSON.parse(localStorage.getItem("laf_user") || "null"),
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token),
    role: (state) => state.user?.role || "",
    isPostAdmin: (state) => ["postadmin", "mainadmin"].includes(state.user?.role),
    isMainAdmin: (state) => state.user?.role === "mainadmin",
  },
  actions: {
    // 登录：保存 token 与 user。
    async login(payload) {
      const data = await authApi.login(payload);
      this.token = data.access_token;
      this.user = data.user;
      tokenStore.set(data.access_token);
      localStorage.setItem("laf_user", JSON.stringify(data.user));
      return data.user;
    },
    // 拉取/刷新个人资料（也用于刷新页面后校正 user）。
    async fetchProfile() {
      const data = await authApi.getProfile();
      this.user = data.user;
      localStorage.setItem("laf_user", JSON.stringify(data.user));
      return data;
    },
    // 局部更新本地 user（改资料后同步）。
    setUser(user) {
      this.user = user;
      localStorage.setItem("laf_user", JSON.stringify(user));
    },
    // 登出：清空本地状态。
    logout() {
      this.token = "";
      this.user = null;
      tokenStore.clear();
      localStorage.removeItem("laf_user");
    },
  },
});
