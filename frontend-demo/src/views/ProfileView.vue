<!--
  个人中心：资料展示/编辑、修改密码、我的帖子、注销账号。
-->
<template>
  <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

  <div v-else>
    <h2>我的</h2>

    <div class="card">
      <div class="row-between">
        <div class="row">
          <span class="avatar-lg">{{ initial }}</span>
          <div>
            <div class="profile-name">{{ profile?.user?.name }}</div>
            <div class="muted text-sm">
              {{ profile?.user?.username }} ·
              {{ roleLabel }}
            </div>
          </div>
        </div>
        <button class="btn btn-ghost btn-sm" @click="editing = !editing">
          {{ editing ? "取消编辑" : "编辑资料" }}
        </button>
      </div>

      <form v-if="editing" class="mt-16" @submit.prevent="handleUpdateProfile">
        <div class="field">
          <label>姓名</label>
          <input v-model.trim="profileForm.name" class="input" />
        </div>
        <div class="field">
          <label>用户名（学号，纯数字）</label>
          <input v-model.trim="profileForm.username" class="input" />
        </div>
        <button class="btn" :disabled="saving" type="submit">{{ saving ? "保存中…" : "保存" }}</button>
      </form>
    </div>

    <div class="card">
      <h3>修改密码</h3>
      <form @submit.prevent="handleChangePassword">
        <div class="field">
          <label>原密码</label>
          <input v-model="pwForm.old_password" type="password" class="input" />
        </div>
        <div class="field">
          <label>新密码（8–16 位）</label>
          <input v-model="pwForm.new_password" type="password" class="input" />
        </div>
        <div class="field">
          <label>确认新密码</label>
          <input v-model="pwForm.confirm_password" type="password" class="input" />
        </div>
        <button class="btn" :disabled="pwSaving" type="submit">{{ pwSaving ? "提交中…" : "修改密码" }}</button>
      </form>
    </div>

    <div class="card">
      <h3>我发布的帖子（{{ myPosts.length }}）</h3>
      <EmptyState v-if="myPosts.length === 0" text="还没有发布过帖子" icon="📝">
        <router-link to="/posts/new" class="btn">去发布</router-link>
      </EmptyState>
      <div v-else class="grid grid-posts">
        <PostCard v-for="p in myPosts" :key="p.id" :post="p" :show-status="auth.isPostAdmin" />
      </div>
    </div>

    <div class="card danger-zone">
      <h3>危险操作</h3>
      <p class="muted text-sm">注销后本人账号及名下帖子/评论将被软删除，可通过申诉恢复。</p>
      <button class="btn btn-danger" @click="handleDeleteAccount">注销我的账号</button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { getProfile, updateProfile, changePassword, deleteAccount } from "@/api/auth";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";
import PostCard from "@/components/PostCard.vue";
import EmptyState from "@/components/EmptyState.vue";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();

const profile = ref(null);
const myPosts = ref([]);
const loading = ref(true);

const editing = ref(false);
const profileForm = reactive({ name: "", username: "" });
const saving = ref(false);

const pwForm = reactive({ old_password: "", new_password: "", confirm_password: "" });
const pwSaving = ref(false);

const initial = computed(() => (profile.value?.user?.name || "?").slice(0, 1));
const roleLabel = computed(
  () => ({ student: "学生", postadmin: "帖子管理员", mainadmin: "超级管理员" })[auth.role] || auth.role
);

async function loadProfile() {
  loading.value = true;
  try {
    const data = await getProfile();
    profile.value = data;
    myPosts.value = data.posts || [];
    profileForm.name = data.user?.name || "";
    profileForm.username = data.user?.username || "";
    auth.setUser(data.user);
  } catch (e) {
    toast.error(e?.msg || "加载资料失败");
  } finally {
    loading.value = false;
  }
}

async function handleUpdateProfile() {
  saving.value = true;
  try {
    const data = await updateProfile({ name: profileForm.name, username: profileForm.username });
    profile.value = data;
    auth.setUser(data.user);
    toast.success("资料已更新");
    editing.value = false;
  } catch (e) {
    toast.error(e?.msg || "更新失败");
  } finally {
    saving.value = false;
  }
}

async function handleChangePassword() {
  if (pwForm.new_password !== pwForm.confirm_password) {
    toast.error("两次输入的新密码不一致");
    return;
  }
  pwSaving.value = true;
  try {
    await changePassword({ ...pwForm });
    toast.success("密码已修改");
    pwForm.old_password = pwForm.new_password = pwForm.confirm_password = "";
  } catch (e) {
    toast.error(e?.msg || "修改失败");
  } finally {
    pwSaving.value = false;
  }
}

async function handleDeleteAccount() {
  if (!window.confirm("确定注销账号？此操作将软删除你的账号与名下内容。")) return;
  try {
    await deleteAccount();
    toast.success("账号已注销，可通过申诉恢复");
    auth.logout();
    router.push("/posts");
  } catch (e) {
    toast.error(e?.msg || "注销失败");
  }
}

onMounted(loadProfile);
</script>

<style scoped>
.avatar-lg {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}
.profile-name {
  font-size: 18px;
  font-weight: 700;
}
.danger-zone {
  border-color: #f3c9c9;
}
</style>
