// 定位结果缓存：发布/编辑帖子的多个步骤可复用同一次定位，避免重复调用。
import { defineStore } from "pinia";

export const useLocationStore = defineStore("location", {
  state: () => ({
    // 最近一次确认的地点：{ location:{id,name,...}, distance_meters, match_type, supplement }
    current: null,
  }),
  actions: {
    set(result) {
      this.current = result;
    },
    clear() {
      this.current = null;
    },
  },
});
