<template>
  <el-card class="post-card" shadow="hover" @click="goToDetail">
    <div class="post-header">
      <el-avatar :size="40" :src="''">
        {{ 'U' }}
      </el-avatar>
      <div class="post-info">
        <div class="username">{{ '用户' + post.userId }}</div>
        <div class="time">{{ formatTime(post.createdAt) }}</div>
      </div>
    </div>
    <div class="post-title">{{ post.title }}</div>
    <div class="post-stats">
      <span><el-icon><View /></el-icon> {{ post.viewCount }}</span>
      <span><el-icon><ChatDotRound /></el-icon> {{ post.commentCount }}</span>
      <span><el-icon><Star /></el-icon> {{ post.likeCount }}</span>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { View, ChatDotRound, Star } from '@element-plus/icons-vue'
import type { PostInfo } from '@/types'
import { formatTime } from '@/utils/date'

interface Props {
  post: PostInfo
}

const props = defineProps<Props>()
const router = useRouter()

const goToDetail = () => {
  router.push(`/post/${props.post.id}`)
}
</script>

<style scoped>
.post-card {
  cursor: pointer;
}

.post-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.post-info {
  flex: 1;
}

.username {
  font-weight: 500;
  font-size: 14px;
}

.time {
  font-size: 12px;
  color: #909399;
}

.post-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 12px;
  line-height: 1.5;
}

.post-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #909399;
}

.post-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
