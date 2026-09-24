// 统一 HTTP 客户端（axios 实例 + 请求/响应拦截器）。
// 约定后端统一响应信封 { code, msg, data }：
//   - 成功 code === 0  → 直接把 data 交给业务层；
//   - 失败 code !== 0  → 抛出带有后端 msg 的错误对象，交给调用方/拦截器提示。
// 鉴权：请求拦截器自动注入 Authorization: Bearer <token>。
import axios from "axios";

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
  timeout: 15000,
  // 关键：后端用 c.QueryArray("type") 读取重复参数，
  // 需要序列化成 type=lost&type=found，而不是 axios 默认的 type[]=lost&type[]=found。
  paramsSerializer: {
    serialize(params) {
      const parts = [];
      Object.keys(params || {}).forEach((key) => {
        const value = params[key];
        if (value === undefined || value === null || value === "") return;
        if (Array.isArray(value)) {
          value.forEach((item) => {
            if (item === undefined || item === null || item === "") return;
            parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(item)}`);
          });
        } else {
          parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`);
        }
      });
      return parts.join("&");
    },
  },
});

// token 读写集中在这里，避免各处硬编码 localStorage 键名。
const TOKEN_KEY = "laf_access_token";
export const tokenStore = {
  get: () => localStorage.getItem(TOKEN_KEY) || "",
  set: (token) => localStorage.setItem(TOKEN_KEY, token || ""),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

// 请求拦截器：有 token 就带上。
http.interceptors.request.use((config) => {
  const token = tokenStore.get();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器：
//   - code === 0 直接返回 data；
//   - code !== 0 抛出 { code, msg }，并把 401 当作登录失效处理（清 token + 跳登录）。
http.interceptors.response.use(
  (res) => {
    const body = res.data;
    // 文件流等非信封响应，原样返回。
    if (!body || typeof body !== "object" || !("code" in body)) {
      return body;
    }
    if (body.code === 0) {
      return body.data;
    }
    return Promise.reject({ code: body.code, msg: body.msg || "请求失败" });
  },
  (err) => {
    const status = err.response?.status;
    const body = err.response?.data;
    const message = body?.msg || err.message || "网络异常，请稍后重试";
    if (status === 401) {
      tokenStore.clear();
      // 通知 auth store 重置（避免循环依赖，用自定义事件解耦）。
      window.dispatchEvent(new CustomEvent("laf:unauthorized"));
    }
    return Promise.reject({ code: body?.code ?? status, msg: message, status });
  }
);

export default http;
