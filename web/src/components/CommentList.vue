<template>
  <el-card class="comment-list">
    <template #header>
      <div class="comment-header">
        <span>评论 ({{ total }})</span>
      </div>
    </template>
    <CommentInput :groupId="postId" @commentCreated="onCommentCreated" />
    <div class="comments">
      <el-skeleton :loading="loading" animated>
        <div v-if="commentItems.length === 0 && !loading" class="empty">
          <el-empty description="暂无评论" />
        </div>
        <div v-else>
          <CommentItem 
            v-for="item in commentItems" 
            :key="item.root.id" 
            :commentItem="item"
            :groupId="postId"
            @replyCreated="onCommentCreated"
          />
        </div>
      </el-skeleton>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getCommentList } from '@/api/comment'
import type { CommentItem as CommentItemType } from '@/types'
import CommentInput from './CommentInput.vue'
import CommentItem from './CommentItem.vue'

interface Props {
  postId: number
}

const props = defineProps<Props>()
const loading = ref(true)
const total = ref(0)
const commentItems = ref<CommentItemType[]>([])

const fetchComments = async () => {
  try {
    const res = await getCommentList({ groupId: props.postId, pageSize: 50 })
    total.value = res.total
    commentItems.value = res.list
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const onCommentCreated = () => {
  fetchComments()
}

onMounted(() => {
  fetchComments()
})
</script>

<style scoped>
.comment-header {
  font-weight: 500;
}

.comments {
  margin-top: 20px;
}

.empty {
  padding: 40px 0;
}
</style>
