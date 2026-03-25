<template>
  <Layout>
    <div class="home-page">
      <div class="post-list">
        <el-skeleton :loading="loading" animated>
          <div v-if="postList.length === 0 && !loading" class="empty">
            <el-empty description="暂无帖子" />
          </div>
          <div v-else>
            <PostCard v-for="post in postList" :key="post.id" :post="post" />
          </div>
        </el-skeleton>
        <div v-if="!loading && hasMore" class="load-more">
          <el-button @click="loadMore" :loading="loadingMore">加载更多</el-button>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPostList } from '@/api/content'
import type { PostInfo } from '@/types'
import Layout from '@/components/Layout.vue'
import PostCard from '@/components/PostCard.vue'

const loading = ref(true)
const loadingMore = ref(false)
const postList = ref<PostInfo[]>([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)

const fetchPostList = async (isLoadMore = false) => {
  try {
    const res = await getPostList({ page: page.value, pageSize })
    if (isLoadMore) {
      postList.value = [...postList.value, ...res.list]
    } else {
      postList.value = res.list
    }
    hasMore.value = postList.value.length < res.total
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

const loadMore = () => {
  page.value++
  loadingMore.value = true
  fetchPostList(true)
}

onMounted(() => {
  fetchPostList()
})
</script>

<style scoped>
.home-page {
  max-width: 800px;
  margin: 0 auto;
}

.post-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.empty {
  padding: 40px 0;
}

.load-more {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}
</style>
