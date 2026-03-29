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
          <div v-if="postDetail.post.cover" class="post-cover">
            <el-image :src="postDetail.post.cover" fit="cover" />
          </div>
          <div class="post-content">
            <div v-for="(item, index) in postDetail.contents" :key="index" class="content-item">
              {{ item.content }}
            </div>
          </div>
          <div v-if="postDetail.topics.length > 0" class="post-topics">
            <el-tag v-for="(topic, index) in postDetail.topics" :key="index" size="small" style="margin-right: 8px;">
              #{{ topic.topicName }}
            </el-tag>
          </div>
          <div class="post-stats">
            <span><el-icon><ChatDotRound /></el-icon> {{ postDetail.post.commentCount }} 评论</span>
            <span><el-icon><Star /></el-icon> {{ postDetail.post.upvoteCount }} 点赞</span>
            <span><el-icon><Share /></el-icon> {{ postDetail.post.shareCount }} 分享</span>
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
import { ChatDotRound, Star, Share } from '@element-plus/icons-vue'
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

.post-cover {
  margin-bottom: 20px;
  border-radius: 8px;
  overflow: hidden;
}

.post-cover :deep(.el-image) {
  width: 100%;
}

.post-content {
  font-size: 16px;
  line-height: 1.8;
  color: #303133;
  margin-bottom: 20px;
}

.content-item {
  margin-bottom: 12px;
}

.post-topics {
  margin-bottom: 20px;
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
