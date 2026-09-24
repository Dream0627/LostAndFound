// 帖子相关接口。
import http from "./http";

// 帖子列表（公开，带 token 可识别管理员身份）。
// params: { type?: string[], status?: string[], finished?: 'true'|'false', page, page_size }
// 返回 { list, total, page, page_size }；后端已将“未完成”排在前面。
export function listPosts(params = {}) {
  return http.get("/posts", { params });
}

// 发布帖子：multipart/form-data。
// fields: type(lost|found), title, content, image?(File),
//         location_id? | latitude?+longitude?, supplement?
export function createPost(formData) {
  return http.post("/posts", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}

// 帖子详情（公开）
export function getPost(postId) {
  return http.get(`/posts/${postId}`);
}

// 删除帖子（作者或管理员）
export function deletePost(postId) {
  return http.delete(`/posts/${postId}`);
}

// 恢复被删帖子
export function recoverPost(postId) {
  return http.patch(`/posts/${postId}/recover`);
}

// 审核帖子（管理员）：status = approved | rejected
export function reviewPost(postId, status) {
  return http.patch(`/posts/${postId}/review`, { status });
}

// 某帖评论列表（公开，分页）
export function listPostComments(postId, params = {}) {
  return http.get(`/posts/${postId}/comments`, { params });
}

// 对帖子发起申领(found)/召领(lost)，开启对话；幂等。
export function startConversation(postId) {
  return http.post(`/posts/${postId}/conversations`);
}
