<template>
  <Layout>
    <div class="follow-list-page">
      <el-card>
        <template #header>
          <div class="card-header">
            <h2>关注列表</h2>
          </div>
        </template>
        <el-skeleton :loading="loading" animated>
          <div v-if="list.length === 0 && !loading" class="empty">
            <el-empty description="暂无关注" />
          </div>
          <div v-else class="user-list">
            <div v-for="item in list" :key="item.userId" class="user-item">
              <el-avatar :size="50" :src="item.avatar" @click="goToProfile(item.userId)">
                {{ item.username?.charAt(0) || 'U' }}
              </el-avatar>
              <div class="user-info" @click="goToProfile(item.userId)">
                <div class="username">{{ item.username }}</div>
                <div class="follow-time">关注于 {{ formatTime(item.followTime) }}</div>
              </div>
              <FollowButton :targetUserId="item.userId" />
            </div>
          </div>
        </el-skeleton>
      </el-card>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getFollowingList } from '@/api/relation'
import type { FollowInfo } from '@/types'
import { formatTime } from '@/utils/date'
import Layout from '@/components/Layout.vue'
import FollowButton from '@/components/FollowButton.vue'

const router = useRouter()
const loading = ref(true)
const list = ref<FollowInfo[]>([])

const fetchList = async () => {
  try {
    const res = await getFollowingList({ page: 1, pageSize: 100 })
    list.value = res.list
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const goToProfile = (userId: number) => {
  router.push(`/user/${userId}`)
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.follow-list-page {
  max-width: 800px;
  margin: 0 auto;
}

.card-header h2 {
  margin: 0;
  font-size: 20px;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
}

.user-info {
  flex: 1;
  cursor: pointer;
}

.username {
  font-weight: 500;
  font-size: 16px;
  margin-bottom: 4px;
}

.follow-time {
  font-size: 12px;
  color: #909399;
}

.empty {
  padding: 40px 0;
}
</style>
