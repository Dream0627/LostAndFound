// 公告接口：全站公开的公告列表（发布 / 删除属管理员，见 admin.js）。
import http from "./http";

// 公告列表（公开，分页）：{ list, total, page, page_size }
export function listAnnouncements(params = {}) {
  return http.get("/announcements", { params });
}