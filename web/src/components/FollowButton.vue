<template>
  <el-button
    :type="isFollow ? 'info' : 'primary'"
    @click="handleFollow"
    :loading="loading"
  >
    {{ isFollow ? '已关注' : '关注' }}
  </el-button>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { follow, unfollow, checkFollowStatus } from '@/api/relation'

interface Props {
  targetUserId: number
}

const props = defineProps<Props>()
const loading = ref(false)
const isFollow = ref(false)

const fetchFollowStatus = async () => {
  try {
    const res = await checkFollowStatus({ targetUserId: props.targetUserId })
    isFollow.value = res.isFollow
  } catch (error) {
    console.error(error)
  }
}

const handleFollow = async () => {
  loading.value = true
  try {
    if (isFollow.value) {
      await unfollow(props.targetUserId)
      ElMessage.success('取消关注成功')
      isFollow.value = false
    } else {
      await follow({ targetUserId: props.targetUserId })
      ElMessage.success('关注成功')
      isFollow.value = true
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFollowStatus()
})
</script>
