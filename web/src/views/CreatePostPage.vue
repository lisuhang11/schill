<template>
  <Layout>
    <div class="create-post-page">
      <el-card>
        <template #header>
          <div class="card-header">
            <h2>发布帖子</h2>
          </div>
        </template>
        <el-form :model="postForm" :rules="rules" ref="postFormRef" label-width="80px">
          <el-form-item label="标题" prop="title">
            <el-input v-model="postForm.title" placeholder="请输入帖子标题" maxlength="100" show-word-limit />
          </el-form-item>
          <el-form-item label="封面" prop="cover">
            <el-input v-model="postForm.cover" placeholder="请输入封面图片URL（可选）" />
          </el-form-item>
          <el-form-item label="内容" prop="content">
            <el-input
              v-model="postForm.content"
              type="textarea"
              :rows="12"
              placeholder="请输入帖子内容"
              maxlength="10000"
              show-word-limit
            />
          </el-form-item>
          <el-form-item label="话题">
            <div class="topic-input-wrapper">
              <el-tag
                v-for="(topic, index) in postForm.topics"
                :key="index"
                closable
                @close="removeTopic(index)"
                style="margin-right: 8px; margin-bottom: 8px;"
              >
                {{ topic }}
              </el-tag>
              <el-input
                v-if="showTopicInput"
                v-model="newTopic"
                placeholder="输入话题后按回车添加"
                size="small"
                style="width: 200px;"
                @keyup.enter="addTopic"
                @blur="hideTopicInput"
                ref="topicInputRef"
              />
              <el-button v-else type="primary" size="small" text @click="showTopicInputBox">
                + 添加话题
              </el-button>
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSubmit" :loading="loading">
              发布
            </el-button>
            <el-button @click="router.back()">取消</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createPost } from '@/api/content'
import Layout from '@/components/Layout.vue'

const router = useRouter()
const postFormRef = ref<FormInstance>()
const topicInputRef = ref<HTMLElement>()
const loading = ref(false)
const showTopicInput = ref(false)
const newTopic = ref('')

const postForm = reactive({
  title: '',
  cover: '',
  visibility: 90,
  content: '',
  topics: [] as string[]
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入帖子标题', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入帖子内容', trigger: 'blur' }
  ]
}

const showTopicInputBox = () => {
  showTopicInput.value = true
  nextTick(() => {
    if (topicInputRef.value) {
      (topicInputRef.value as any).focus()
    }
  })
}

const addTopic = () => {
  const topic = newTopic.value.trim()
  if (topic && !postForm.topics.includes(topic)) {
    postForm.topics.push(topic)
  }
  newTopic.value = ''
  showTopicInput.value = false
}

const removeTopic = (index: number) => {
  postForm.topics.splice(index, 1)
}

const hideTopicInput = () => {
  if (!newTopic.value.trim()) {
    showTopicInput.value = false
  }
}

const handleSubmit = async () => {
  if (!postFormRef.value) return
  await postFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const contents = [{
          type: 2,
          content: postForm.content,
          sort: 10
        }]
        const res = await createPost({
          title: postForm.title,
          cover: postForm.cover,
          visibility: postForm.visibility,
          contents: contents,
          topics: postForm.topics,
          tags: ''
        })
        ElMessage.success('发布成功')
        router.push(`/post/${res.postId}`)
      } catch (error) {
        console.error(error)
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.create-post-page {
  max-width: 800px;
  margin: 0 auto;
}

.card-header h2 {
  margin: 0;
  font-size: 20px;
}

.topic-input-wrapper {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}
</style>
