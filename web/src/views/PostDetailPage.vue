<template>
  <Layout>
    <div class="post-detail-page" v-loading="loading">
      <div v-if="postDetail" class="post-detail">
        <el-card>
          <div class="post-header">
            <el-avatar :size="48" :src="''">
              {{ 'U' }}
            </el-avatar>
            <div class="post-info">
              <div class="username">{{ '用户' + postDetail.post.userId }}</div>
              <div class="time">{{ formatTime(postDetail.post.createdAt) }}</div>
            </div>
          </div>
          <h1 class="post-title">{{ postDetail.post.title }}</h1>
          <div class="post-content" v-html="postDetail.content"></div>
          <div class="post-stats">
            <span><el-icon><View /></el-icon> {{ postDetail.post.viewCount }} 浏览</span>
            <span><el-icon><ChatDotRound /></el-icon> {{ postDetail.post.commentCount }} 评论</span>
            <span><el-icon><Star /></el-icon> {{ postDetail.post.likeCount }} 点赞</span>
          </div>
        </el-card>
        <CommentList :postId="postId" />
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { View, ChatDotRound, Star } from '@element-plus/icons-vue'
import { getPostDetail, incViewCount } from '@/api/content'
import type { PostDetail } from '@/types'
import { formatTime } from '@/utils/date'
import Layout from '@/components/Layout.vue'
import CommentList from '@/components/CommentList.vue'

const route = useRoute()
const postId = Number(route.params.id)
const loading = ref(true)
const postDetail = ref<PostDetail | null>(null)

const fetchPostDetail = async () => {
  try {
    const res = await getPostDetail(postId)
    postDetail.value = res
    await incViewCount(postId)
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchPostDetail()
})
</script>

<style scoped>
.post-detail-page {
  max-width: 800px;
  margin: 0 auto;
}

.post-detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.post-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.post-info {
  flex: 1;
}

.username {
  font-weight: 500;
  font-size: 16px;
}

.time {
  font-size: 12px;
  color: #909399;
}

.post-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  line-height: 1.4;
}

.post-content {
  font-size: 16px;
  line-height: 1.8;
  color: #303133;
  margin-bottom: 20px;
  white-space: pre-wrap;
}

.post-stats {
  display: flex;
  gap: 24px;
  font-size: 14px;
  color: #909399;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.post-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
