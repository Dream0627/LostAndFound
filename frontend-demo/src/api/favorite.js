// 收藏帖子接口：收藏 / 取消收藏。同一帖重复收藏幂等；取消为软删除（再次收藏会新插入一行）。
import http from "./http";

// 收藏帖子（需登录）
export function addFavorite(postId) {
  return http.post(`/posts/${postId}/favorite`);
}

// 取消收藏帖子（需登录）
export function removeFavorite(postId) {
  return http.delete(`/posts/${postId}/favorite`);
}