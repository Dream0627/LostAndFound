// 申诉接口（公开，无需登录）。
import http from "./http";

// 提交申诉：{ username, reason(self_regret|wrongful_ban|other), content? }
// reason=other 时 content 必填。
export function submitAppeal(data) {
  return http.post("/appeals", data);
}
