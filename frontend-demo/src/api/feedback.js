// 反馈接口：登录用户提交反馈（超级管理员在后台 admin.js 列表 / 审批）。
import http from "./http";

// 提交反馈：{ content }（1~1000 字符，默认待处理状态）
export function submitFeedback(content) {
  return http.post("/feedbacks", { content });
}