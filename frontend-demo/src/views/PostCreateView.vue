<!--
  发布帖子：multipart/form-data。
  含类型/标题/内容/图片，以及可选的位置（复用 LocationPicker）。
  位置二选一：手动选择 location_id，或自动定位 latitude/longitude；二者都通过组件确认后随表单提交。
-->
<template>
  <div class="create-wrap">
    <div class="card">
      <h2>发布帖子</h2>
      <p class="muted text-sm">
        学生发布后为“待审核”，管理员发布直接通过。带 📍 的位置为可选信息，用于就近推荐。
      </p>

      <div v-if="errorMsg" class="alert alert-error mt-16">{{ errorMsg }}</div>

      <form class="mt-16" @submit.prevent="handleSubmit">
        <div class="field">
          <label>类型</label>
          <div class="row">
            <label class="radio" :class="{ active: form.type === 'lost' }">
              <input v-model="form.type" type="radio" value="lost" /> 失物（我丢了东西）
            </label>
            <label class="radio" :class="{ active: form.type === 'found' }">
              <input v-model="form.type" type="radio" value="found" /> 招领（我捡到东西）
            </label>
          </div>
        </div>

        <div class="field">
          <label>标题</label>
          <input v-model.trim="form.title" class="input" maxlength="100" placeholder="简要描述，如：遗失一张校园卡" />
        </div>

        <div class="field">
          <label>内容（1–2000 字）</label>
          <textarea v-model.trim="form.content" class="textarea" maxlength="2000" placeholder="请描述物品特征、丢失/拾取时间与地点等"></textarea>
          <div class="counter text-sm muted">{{ form.content.length }} / 2000</div>
        </div>

        <div class="field">
          <label>图片（可选）</label>
          <input ref="fileInput" type="file" accept="image/*" class="input" @change="onFileChange" />
          <div v-if="previewUrl" class="preview mt-8">
            <img :src="previewUrl" alt="预览" />
          </div>
        </div>

        <div class="field">
          <label>📍 地点（可选）</label>
          <LocationPicker @confirm="onLocationConfirm" />
          <div v-if="location" class="loc-result mt-8">
            已选择：<strong>{{ location.location?.name }}</strong>
            <span v-if="location.supplement" class="muted">（{{ location.supplement }}）</span>
          </div>
        </div>

        <div class="row mt-16">
          <button class="btn" :disabled="submitting" type="submit">
            {{ submitting ? "发布中…" : "发布帖子" }}
          </button>
          <router-link to="/posts" class="btn btn-ghost">取消</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { createPost } from "@/api/post";
import { useToastStore } from "@/stores/toast";
import LocationPicker from "@/components/LocationPicker.vue";

const router = useRouter();
const toast = useToastStore();

const form = reactive({ type: "lost", title: "", content: "" });
const file = ref(null);
const previewUrl = ref("");
const location = ref(null);
const submitting = ref(false);
const errorMsg = ref("");

function onFileChange(e) {
  const f = e.target.files?.[0] || null;
  file.value = f;
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
  previewUrl.value = f ? URL.createObjectURL(f) : "";
}

function onLocationConfirm(payload) {
  location.value = payload;
}

function validate() {
  if (!form.title) return "请填写标题";
  if (!form.content) return "请填写内容";
  return "";
}

async function handleSubmit() {
  errorMsg.value = validate();
  if (errorMsg.value) return;

  submitting.value = true;
  try {
    const fd = new FormData();
    fd.append("type", form.type);
    fd.append("title", form.title);
    fd.append("content", form.content);
    if (file.value) fd.append("image", file.value);

    // 位置：手动选择传 location_id；自动定位传 latitude/longitude（组件已解析，用其 location.id 即最稳）。
    if (location.value?.location?.id) {
      fd.append("location_id", location.value.location.id);
    }
    if (location.value?.supplement) {
      fd.append("supplement", location.value.supplement);
    }

    const created = await createPost(fd);
    toast.success("发布成功");
    router.push(`/posts/${created.id}`);
  } catch (e) {
    errorMsg.value = e?.msg || "发布失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}

onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value);
});
</script>

<style scoped>
.create-wrap {
  max-width: 640px;
  margin: 0 auto;
}
.radio {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
}
.radio.active {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary-dark);
  font-weight: 600;
}
.counter {
  text-align: right;
  margin-top: 4px;
}
.preview img {
  max-width: 100%;
  max-height: 220px;
  border-radius: var(--radius-sm);
}
.loc-result {
  font-size: 14px;
  color: var(--color-text);
}
</style>
