// 全局轻提示（toast）状态：任意位置调用 pushToast 即可弹提示。
import { defineStore } from "pinia";

let seed = 0;

export const useToastStore = defineStore("toast", {
  state: () => ({
    items: [], // { id, text, type('info'|'success'|'error') }
  }),
  actions: {
    push(text, type = "info", duration = 2600) {
      const id = ++seed;
      this.items.push({ id, text, type });
      setTimeout(() => this.remove(id), duration);
      return id;
    },
    success(text) {
      return this.push(text, "success");
    },
    error(text) {
      return this.push(text, "error");
    },
    remove(id) {
      this.items = this.items.filter((t) => t.id !== id);
    },
  },
});
