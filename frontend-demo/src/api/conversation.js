// 对话 / 完成寻找 相关接口。
import http from "./http";

// 我的会话列表（分页）：当前用户作为发起方或楼主参与的会话。
export function listMyConversations(params = {}) {
  return http.get("/conversations", { params });
}

// 会话消息列表（分页，仅参与方）。按 id 倒序（新消息在前）。
export function listMessages(conversationId, params = {}) {
  return http.get(`/conversations/${conversationId}/messages`, { params });
}

// 发送消息：{ content }
export function sendMessage(conversationId, content) {
  return http.post(`/conversations/${conversationId}/messages`, { content });
}

// 发起“完成寻找”申请（对话任一方均可）。
export function createFinishRequest(conversationId) {
  return http.post(`/conversations/${conversationId}/finish-requests`);
}

// 处理完成申请：status = agreed | rejected（须为发起方之外的另一方）
export function handleFinishRequest(conversationId, requestId, status) {
  return http.patch(
    `/conversations/${conversationId}/finish-requests/${requestId}`,
    { status }
  );
}
