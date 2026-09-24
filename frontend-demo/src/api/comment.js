// 评论相关接口。
import http from "./http";

// 发表评论：{ post_id, content }（作者取自 token）
export function createComment(data) {
  return http.post("/comments", data);
}

// 删除评论（作者或管理员）
export function deleteComment(commentId) {
  return http.delete(`/comments/${commentId}`);
}
