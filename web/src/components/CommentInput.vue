<template>
  <div class="comment-input">
    <el-input
      v-model="content"
      type="textarea"
      :rows="3"
      placeholder="写下你的评论..."
      :disabled="!userStore.isLoggedIn"
    />
    <div class="comment-actions">
      <span v-if="!userStore.isLoggedIn" class="login-hint">请先登录</span>
      <el-button 
        type="primary" 
        @click="handleSubmit" 
        :loading="loading"
        :disabled="!content.trim() || !userStore.isLoggedIn"
      >
        发表评论
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createComment } from '@/api/comment'
import { useUserStore } from '@/stores/user'

interface Props {
  groupId: number
  parentId?: number
  replyToUserId?: number
}

const props = defineProps<Props>()
const emit = defineEmits(['commentCreated'])

const userStore = useUserStore()
const content = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  if (!content.value.trim()) return
  loading.value = true
  try {
    await createComment({
      groupId: props.groupId,
      parentId: props.parentId || 0,
      replyToUserId: props.replyToUserId || 0,
      content: content.value
    })
    ElMessage.success('评论成功')
    content.value = ''
    emit('commentCreated')
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.comment-input {
  margin-bottom: 20px;
}

.comment-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 10px;
  gap: 10px;
}

.login-hint {
  font-size: 14px;
  color: #909399;
}
</style>
