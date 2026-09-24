// 认证与用户接口。
import http from "./http";

// 注册：{ username(纯数字学号), name, password(8-16), role }
export function register(data) {
  return http.post("/auth/register", data);
}

// 登录：{ username, password } → { access_token, token_type, expires_in, user }
export function login(data) {
  return http.post("/auth/login", data);
}

// 获取个人资料：{ user, posts }
export function getProfile() {
  return http.get("/auth/profile");
}

// 更新个人资料：{ name?, username? }（字段为空则不改）
export function updateProfile(data) {
  return http.patch("/auth/profile", data);
}

// 修改密码：{ old_password, new_password, confirm_password }
export function changePassword(data) {
  return http.patch("/auth/password", data);
}

// 注销本人账号（软删除本人及本人帖子/评论）
export function deleteAccount() {
  return http.delete("/auth/account");
}
