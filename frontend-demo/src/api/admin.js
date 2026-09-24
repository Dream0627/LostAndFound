// 管理员接口。
// postadmin / mainadmin：帖子状态、已删列表、审核帖子/申诉、待审批列表。
// mainadmin 专属：注销/恢复用户。
import http from "./http";

// 直接修改帖子状态（postadmin/mainadmin）：status = pending|approved|rejected
export function updatePostStatus(postId, status) {
  return http.patch(`/admin/posts/${postId}/status`, { status });
}

// 已删除帖子列表（postadmin/mainadmin，分页）
export function listDeletedPosts(params = {}) {
  return http.get("/admin/posts/deleted", { params });
}

// 注销用户（仅 mainadmin）
export function deleteUser(userId) {
  return http.delete(`/admin/users/${userId}`);
}

// 恢复用户（仅 mainadmin）
export function recoverUser(userId) {
  return http.patch(`/admin/users/${userId}/recover`);
}

// 审核申诉（仅 mainadmin）：status = approved|rejected；approved 会自动级联恢复账号
export function reviewAppeal(appealId, status) {
  return http.patch(`/admin/appeals/${appealId}/review`, { status });
}

// 待审批列表（仅 mainadmin）：{ posts:[], appeals:[] }
// params: { type?: 'post'|'appeal', page, page_size }
export function listReviews(params = {}) {
  return http.get("/admin/reviews", { params });
}
