<template>
  <Layout>
    <div class="user-profile-page" v-loading="loading">
      <div v-if="userInfo" class="profile-container">
        <el-card class="profile-card">
          <div class="profile-header">
            <el-avatar :size="100" :src="userInfo.avatar">
              {{ userInfo.username?.charAt(0) || 'U' }}
            </el-avatar>
            <div class="profile-info">
              <h2>{{ userInfo.nickname || userInfo.username }}</h2>
              <p class="signature">{{ userProfile?.signature || '这个人很懒，什么都没写~' }}</p>
              <div class="stats">
                <div class="stat-item">
                  <span class="stat-value">{{ userStat?.postCount || 0 }}</span>
                  <span class="stat-label">帖子</span>
                </div>
                <div class="stat-item">
                  <span class="stat-value">{{ userStat?.followerCount || 0 }}</span>
                  <span class="stat-label">粉丝</span>
                </div>
                <div class="stat-item">
                  <span class="stat-value">{{ userStat?.followingCount || 0 }}</span>
                  <span class="stat-label">关注</span>
                </div>
              </div>
            </div>
            <div class="profile-actions">
              <FollowButton v-if="userStore.isLoggedIn && userStore.userInfo?.id !== userId" :targetUserId="userId" />
              <el-button v-if="userStore.isLoggedIn && userStore.userInfo?.id === userId" type="primary" @click="router.push('/settings')">
                编辑资料
              </el-button>
            </div>
          </div>
        </el-card>
        <div class="user-posts">
          <h3>Ta 的帖子</h3>
          <el-skeleton :loading="postsLoading" animated>
            <div v-if="userPosts.length === 0 && !postsLoading" class="empty">
              <el-empty description="暂无帖子" />
            </div>
            <div v-else>
              <PostCard v-for="post in userPosts" :key="post.id" :post="post" />
            </div>
          </el-skeleton>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUserInfo, getUserProfileInfo, getUserStat } from '@/api/user'
import { getPostList } from '@/api/content'
import type { UserInfo, UserProfileInfo, UserStatInfo, PostInfo } from '@/types'
import { useUserStore } from '@/stores/user'
import Layout from '@/components/Layout.vue'
import PostCard from '@/components/PostCard.vue'
import FollowButton from '@/components/FollowButton.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const userId = Number(route.params.id)

const loading = ref(true)
const postsLoading = ref(false)
const userInfo = ref<UserInfo | null>(null)
const userProfile = ref<UserProfileInfo | null>(null)
const userStat = ref<UserStatInfo | null>(null)
const userPosts = ref<PostInfo[]>([])

const fetchUserInfo = async () => {
  try {
    const [infoRes, profileRes, statRes] = await Promise.all([
      getUserInfo(userId),
      getUserProfileInfo(userId),
      getUserStat(userId)
    ])
    userInfo.value = infoRes.userInfo
    userProfile.value = profileRes.profile || null
    userStat.value = statRes.stat
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchUserPosts = async () => {
  postsLoading.value = true
  try {
    const res = await getPostList({ userId, page: 1, pageSize: 20 })
    userPosts.value = res.list
  } catch (error) {
    console.error(error)
  } finally {
    postsLoading.value = false
  }
}

onMounted(() => {
  fetchUserInfo()
  fetchUserPosts()
})
</script>

<style scoped>
.user-profile-page {
  max-width: 800px;
  margin: 0 auto;
}

.profile-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.profile-header {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.profile-info {
  flex: 1;
}

.profile-info h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
}

.signature {
  margin: 0 0 16px 0;
  color: #606266;
  font-size: 14px;
}

.stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.profile-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.user-posts {
  margin-top: 20px;
}

.user-posts h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
}

.empty {
  padding: 40px 0;
}
</style>
