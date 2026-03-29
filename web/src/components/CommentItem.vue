<template>
  <div class="comment-item">
    <div class="comment-root">
      <el-avatar :size="40" :src="commentItem.root.avatar">
        {{ commentItem.root.username?.charAt(0) || 'U' }}
      </el-avatar>
      <div class="comment-content-wrapper">
        <div class="comment-meta">
          <span class="username">{{ commentItem.root.username }}</span>
          <span class="time">{{ formatTime(commentItem.root.createdAt) }}</span>
        </div>
        <div class="comment-text">{{ commentItem.root.content }}</div>
        <div class="comment-actions">
          <span class="action-btn" @click="showReplyInput = !showReplyInput">
            <el-icon><ChatDotRound /></el-icon> 回复
          </span>
          <span class="action-btn" :class="{ liked: commentItem.root.isLiked }" @click="handleVote(1)">
            <el-icon><Star /></el-icon> {{ commentItem.root.likeCount }}
          </span>
        </div>
        <CommentInput 
          v-if="showReplyInput" 
          :postId="postId"
          :parentId="commentItem.root.id"
          :replyToUserId="commentItem.root.userId"
          @commentCreated="onReplyCreated"
        />
      </div>
    </div>
    <div class="comment-replies" v-if="commentItem.replies.length > 0">
      <div v-for="reply in commentItem.replies" :key="reply.id" class="reply-item">
        <el-avatar :size="32" :src="reply.avatar">
          {{ reply.username?.charAt(0) || 'U' }}
        </el-avatar>
        <div class="reply-content-wrapper">
          <div class="comment-meta">
            <span class="username">{{ reply.username }}</span>
            <span v-if="reply.replyToUsername" class="reply-to">回复 {{ reply.replyToUsername }}</span>
            <span class="time">{{ formatTime(reply.createdAt) }}</span>
          </div>
          <div class="comment-text">{{ reply.content }}</div>
        </div>
      </div>
      <div v-if="commentItem.hasMoreReplies" class="more-replies">
        查看更多回复
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Star } from '@element-plus/icons-vue'
import { voteComment } from '@/api/comment'
import type { CommentItem as CommentItemType } from '@/types'
import { formatTime } from '@/utils/date'
import CommentInput from './CommentInput.vue'

interface Props {
  commentItem: CommentItemType
  postId: number
}

const props = defineProps<Props>()
const emit = defineEmits(['replyCreated'])
const showReplyInput = ref(false)

const handleVote = async (voteType: number) => {
  try {
    await voteComment(props.commentItem.root.id, { voteType })
    ElMessage.success(voteType === 1 ? '点赞成功' : '取消成功')
  } catch (error) {
    console.error(error)
  }
}

const onReplyCreated = () => {
  showReplyInput.value = false
  emit('replyCreated')
}
</script>

<style scoped>
.comment-item {
  padding: 16px 0;
  border-bottom: 1px solid #f5f7fa;
}

.comment-root {
  display: flex;
  gap: 12px;
}

.comment-content-wrapper {
  flex: 1;
}

.comment-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.username {
  font-weight: 500;
  font-size: 14px;
  color: #409eff;
}

.reply-to {
  font-size: 14px;
  color: #909399;
}

.time {
  font-size: 12px;
  color: #909399;
}

.comment-text {
  font-size: 14px;
  color: #303133;
  line-height: 1.6;
  margin-bottom: 8px;
}

.comment-actions {
  display: flex;
  gap: 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #909399;
  cursor: pointer;
  transition: color 0.2s;
}

.action-btn:hover {
  color: #409eff;
}

.action-btn.liked {
  color: #f56c6c;
}

.comment-replies {
  margin-top: 12px;
  margin-left: 52px;
  padding-left: 12px;
  border-left: 2px solid #f5f7fa;
}

.reply-item {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.reply-content-wrapper {
  flex: 1;
}

.more-replies {
  font-size: 14px;
  color: #409eff;
  cursor: pointer;
}
</style>
