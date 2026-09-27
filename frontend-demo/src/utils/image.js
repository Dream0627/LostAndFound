// 帖子图片地址归一化。
// 历史数据里可能存了写死的绝对地址(如 http://127.0.0.1:8080/uploads/...)，
// 这类地址会让浏览器把图片指向“访问者本机”而 404。这里统一提取出 /uploads/ 开头的
// 相对路径，改成同源访问；非 /uploads 的地址(外链图片)原样返回。
export function resolveImageUrl(raw) {
  if (!raw) return "";
  try {
    const u = new URL(raw, window.location.origin);
    if (u.pathname.startsWith("/uploads/")) {
      return u.pathname + u.search;
    }
  } catch (e) {
    // raw 不是合法 URL，按相对路径处理
  }
  return raw;
}
