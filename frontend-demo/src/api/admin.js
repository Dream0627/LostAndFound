// 管理员接口。
// postadmin / mainadmin：帖子状态、已删列表、审核帖子/申诉、待审批列表。
// mainadmin 专属：注销/恢复用户。
import http from "./http";

// 直接修改帖子状态（postadmin/mainadmin）：status = approved|rejected（不允许 pending）
export function updatePostStatus(postId, status) {
  return http.patch(`/admin/posts/${postId}/status`, { status });
}

// 直接修改帖子完成状态（postadmin/mainadmin）：finished = true|false（可反复切换）
export function updatePostFinished(postId, finished) {
  return http.patch(`/admin/posts/${postId}/finished`, { finished });
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

// 后台数据概览（仅 mainadmin）：
// { user_count, post_count, pending_post_count, pending_appeal_count, today_post_count, today_comment_count }
export function getCount() {
  return http.get("/admin/count");
}

// 公告发布（postadmin/mainadmin）：{ title(1~200), content(1~2000) }
export function createAnnouncement(data) {
  return http.post("/admin/announcements", data);
}

// 公告删除（postadmin/mainadmin，软删除）
export function deleteAnnouncement(announcementId) {
  return http.delete(`/admin/announcements/${announcementId}`);
}

// 反馈列表（仅 mainadmin，分页）：params: { status?: 'pending'|'approved'|'rejected', page, page_size }
export function listFeedbacks(params = {}) {
  return http.get("/admin/feedbacks", { params });
}

// 反馈审批（仅 mainadmin）：status = approved|rejected（只有 pending 可被审批）
export function reviewFeedback(feedbackId, status) {
  return http.patch(`/admin/feedbacks/${feedbackId}/review`, { status });
}
