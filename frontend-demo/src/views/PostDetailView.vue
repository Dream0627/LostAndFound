<!--
  帖子详情：帖子信息 + 管理员操作 + 申领/召领入口 + 评论列表/发表/删除。
-->
<template>
  <div v-if="loading" class="loading"><span class="spinner"></span> 加载中…</div>

  <EmptyState v-else-if="!post" text="帖子不存在或无权查看" icon="🚫">
    <router-link to="/posts" class="btn btn-ghost">返回广场</router-link>
  </EmptyState>

  <div v-else>
    <button class="btn btn-ghost btn-sm" @click="$router.back()">← 返回</button>

    <div class="card mt-16">
      <div class="pd-head">
        <div class="row">
          <StatusBadge kind="type" :value="post.type" />
          <StatusBadge kind="finished" :value="post.is_finished" />
          <StatusBadge v-if="auth.isPostAdmin || isOwner" kind="status" :value="post.status" />
        </div>
        <div class="row">
          <!-- 待审核/被驳回：走审核接口 -->
          <button v-if="canReview" class="btn btn-ghost btn-sm" @click="handleReview('approved')">通过</button>
          <button v-if="canReview" class="btn btn-ghost btn-sm" @click="handleReview('rejected')">驳回</button>
          <!-- 已通过：管理员仍可直接改状态（下架/打回待审） -->
          <button v-if="canAdjust" class="btn btn-ghost btn-sm" @click="handleStatusChange('pending')">打回待审</button>
          <button v-if="canAdjust" class="btn btn-ghost btn-sm" @click="handleStatusChange('rejected')">下架</button>
          <button v-if="canDelete" class="btn btn-danger btn-sm" @click="handleDelete">删除</button>
        </div>
      </div>

      <h2 class="pd-title">{{ post.title }}</h2>

      <div class="pd-meta muted text-sm">
        <span>作者：{{ postAuthorName }}</span>
        <span>·</span>
        <span>{{ formatTime(post.created_at) }}</span>
        <template v-if="post.location_name">
          <span>·</span>
          <span>📍 {{ post.location_name }}</span>
        </template>
      </div>

      <!-- 作者查看自己的非公开帖时给出明确提示 -->
      <div v-if="isOwner && !auth.isPostAdmin && post.status === 'pending'" class="alert alert-warning mt-8">
        该帖子正在等待管理员审核，暂未对其他用户公开。
      </div>
      <div v-else-if="isOwner && !auth.isPostAdmin && post.status === 'rejected'" class="alert alert-warning mt-8">
        该帖子未通过审核（已驳回），仅你和管理员可见。
      </div>

      <p v-if="post.supplement" class="pd-supplement muted">补充说明：{{ post.supplement }}</p>

      <img v-if="post.image_url" :src="postImageSrc" class="pd-image" alt="帖子图片" />

      <p class="pd-content">{{ post.content }}</p>

      <div v-if="post.is_finished" class="alert alert-info mt-16">该帖子已完成寻找，无法再发起对话。</div>
      <div v-else-if="canStartConversation" class="pd-actions">
        <button class="btn" :disabled="starting" @click="handleStartConversation">
          {{ starting ? "处理中…" : post.type === "found" ? "申领（联系拾得者）" : "召领（联系失主）" }}
        </button>
        <span class="muted text-sm">点击将与对方开启一对一会话</span>
      </div>
    </div>

    <!-- 评论区 -->
    <div class="card">
      <h3>评论（{{ cTotal }}）</h3>

      <div v-if="auth.isLoggedIn" class="comment-form">
        <textarea v-model.trim="commentText" class="textarea" maxlength="1000" placeholder="说点什么…"></textarea>
        <div class="row mt-8">
          <button class="btn" :disabled="posting" @click="handleComment">{{ posting ? "发送中…" : "发表评论" }}</button>
        </div>
      </div>
      <p v-else class="muted">
        <router-link to="/login">登录</router-link> 后可发表评论
      </p>

      <EmptyState v-if="comments.length === 0" text="还没有评论" icon="💬" />

      <ul v-else class="comment-list">
        <li v-for="c in comments" :key="c.id" class="comment-item">
          <div class="row-between">
            <span class="comment-user">{{ c.author_name || ('用户 ' + c.user_id) }}</span>
            <div class="row">
              <span class="muted text-sm">{{ formatTime(c.created_at) }}</span>
              <button v-if="canDeleteComment(c)" class="btn btn-ghost btn-sm" @click="handleDeleteComment(c)">删除</button>
            </div>
          </div>
          <p class="comment-content">{{ c.content }}</p>
        </li>
      </ul>

      <Pagination :page="cPage" :page-size="cPageSize" :total="cTotal" @change="onCommentPageChange" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  getPost,
  deletePost,
  reviewPost,
  listPostComments,
  startConversation,
} from "@/api/post";
import { createComment, deleteComment } from "@/api/comment";
import { updatePostStatus } from "@/api/admin";
import { useAuthStore } from "@/stores/auth";
import { useToastStore } from "@/stores/toast";
import StatusBadge from "@/components/StatusBadge.vue";
import EmptyState from "@/components/EmptyState.vue";
import Pagination from "@/components/Pagination.vue";
import { resolveImageUrl } from "@/utils/image";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const toast = useToastStore();

const postId = route.params.id;
const post = ref(null);
const loading = ref(true);
const starting = ref(false);

const comments = ref([]);
const cPage = ref(1);
const cPageSize = ref(20);
const cTotal = ref(0);
const commentText = ref("");
const posting = ref(false);

const isOwner = computed(() => auth.isLoggedIn && post.value?.user_id === auth.user?.id);
// 未完成的非 approved 帖：管理员可走审核流（通过/驳回）。
const canReview = computed(
  () => auth.isPostAdmin && !post.value?.is_finished && post.value?.status !== "approved"
);
// 已通过的帖子：管理员可随时改状态（下架=rejected、打回待审=pending）。
const canAdjust = computed(
  () => auth.isPostAdmin && !post.value?.is_finished && post.value?.status === "approved"
);
const canDelete = computed(
  () => auth.isLoggedIn && (auth.isPostAdmin || post.value?.user_id === auth.user?.id)
);
const canStartConversation = computed(
  () => auth.isLoggedIn && post.value && post.value.user_id !== auth.user?.id
);

const postAuthorName = computed(
  () => post.value?.author_name || (post.value ? `用户 ${post.value.user_id}` : "")
);
const postImageSrc = computed(() => resolveImageUrl(post.value?.image_url));

function canDeleteComment(c) {
  return auth.isLoggedIn && (auth.isPostAdmin || c.user_id === auth.user?.id);
}

function formatTime(value) {
  if (!value) return "";
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString("zh-CN", { hour12: false });
}

async function fetchPost() {
  loading.value = true;
  try {
    post.value = await getPost(postId);
  } catch (e) {
    post.value = null;
  } finally {
    loading.value = false;
  }
}

async function fetchComments() {
  try {
    const data = await listPostComments(postId, { page: cPage.value, page_size: cPageSize.value });
    comments.value = data.list || [];
    cTotal.value = data.total || 0;
  } catch (e) {
    comments.value = [];
    cTotal.value = 0;
  }
}

async function handleReview(status) {
  const tip = status === "approved" ? "确定通过该帖子的审核？" : "确定驳回该帖子？驳回后仅作者与管理员可见。";
  if (!window.confirm(tip)) return;
  try {
    await reviewPost(postId, status);
    toast.success(status === "approved" ? "已通过审核" : "已驳回");
    fetchPost();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

// 管理员直接修改“已通过”帖子的状态（PATCH /admin/posts/:id/status）。
async function handleStatusChange(status) {
  const tip =
    status === "rejected"
      ? "确定下架该帖子？下架后其他用户将无法在广场看到它。"
      : "确定把该帖子打回待审核？";
  if (!window.confirm(tip)) return;
  try {
    await updatePostStatus(postId, status);
    toast.success(status === "rejected" ? "已下架" : "已设为待审核");
    fetchPost();
  } catch (e) {
    toast.error(e?.msg || "操作失败");
  }
}

async function handleDelete() {
  if (!window.confirm("确定删除该帖子？删除后其评论也会一并删除。")) return;
  try {
    await deletePost(postId);
    toast.success("已删除");
    router.push("/posts");
  } catch (e) {
    toast.error(e?.msg || "删除失败");
  }
}

async function handleStartConversation() {
  starting.value = true;
  try {
    const conv = await startConversation(postId);
    toast.success("已开启会话");
    router.push(`/conversations/${conv.id}`);
  } catch (e) {
    toast.error(e?.msg || "无法发起会话");
  } finally {
    starting.value = false;
  }
}

async function handleComment() {
  if (!commentText.value) return;
  posting.value = true;
  try {
    await createComment({ post_id: Number(postId), content: commentText.value });
    commentText.value = "";
    toast.success("评论已发表");
    cPage.value = 1;
    fetchComments();
  } catch (e) {
    toast.error(e?.msg || "评论失败");
  } finally {
    posting.value = false;
  }
}

async function handleDeleteComment(c) {
  if (!window.confirm("确定删除该评论？")) return;
  try {
    await deleteComment(c.id);
    toast.success("已删除评论");
    fetchComments();
  } catch (e) {
    toast.error(e?.msg || "删除失败");
  }
}

function onCommentPageChange(p) {
  cPage.value = p;
  fetchComments();
}

onMounted(() => {
  fetchPost();
  fetchComments();
});
</script>

<style scoped>
.pd-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
}
.pd-title {
  margin: 6px 0;
}
.pd-meta {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.pd-supplement {
  margin: 10px 0 0;
}
.pd-image {
  width: 100%;
  max-height: 420px;
  object-fit: contain;
  border-radius: var(--radius);
  margin: 14px 0;
  background: #f7f9fb;
}
.pd-content {
  white-space: pre-wrap;
  margin: 14px 0 0;
  font-size: 15px;
}
.pd-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}
.comment-form {
  margin-bottom: 16px;
}
.comment-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.comment-item {
  padding: 12px 0;
  border-top: 1px solid var(--color-border);
}
.comment-user {
  font-weight: 600;
  font-size: 14px;
}
.comment-content {
  margin: 6px 0 0;
  white-space: pre-wrap;
}
</style>
