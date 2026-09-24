// 地理位置相关接口封装（复用统一 http 实例）。
// 二选一：手动选择传 location_id；自动定位传 latitude + longitude；均可带 supplement 补充说明。
import http from "./http";

// 拉取校园预设地点（按校区分组），用于手动选择器。
// 返回：[{ campus, locations: [{ id, name, category, address, latitude, longitude }, ...] }, ...]
export function fetchCampusLocations() {
  return http.get("/geo/locations");
}

// 定位/匹配：二选一提交 location_id 或 latitude/longitude，可附带 supplement。
// 返回：{ location, distance_meters, match_type('auto'|'manual'), supplement }
export function locate({ locationId, latitude, longitude, supplement } = {}) {
  const body = {};
  if (locationId) body.location_id = locationId;
  if (typeof latitude === "number" && typeof longitude === "number") {
    body.latitude = latitude;
    body.longitude = longitude;
  }
  if (supplement) body.supplement = supplement;
  return http.post("/geo/locate", body);
}
