<template>
  <Layout>
    <div class="user-settings-page">
      <el-card>
        <template #header>
          <div class="card-header">
            <h2>个人设置</h2>
          </div>
        </template>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本资料" name="profile">
            <el-form :model="profileForm" :rules="rules" ref="profileFormRef" label-width="100px">
              <el-form-item label="头像">
                <el-avatar :size="80" :src="profileForm.avatar">
                  {{ userStore.userInfo?.username?.charAt(0) || 'U' }}
                </el-avatar>
              </el-form-item>
              <el-form-item label="头像URL" prop="avatar">
                <el-input v-model="profileForm.avatar" placeholder="请输入头像URL" />
              </el-form-item>
              <el-form-item label="性别" prop="gender">
                <el-radio-group v-model="profileForm.gender">
                  <el-radio :value="0">保密</el-radio>
                  <el-radio :value="1">男</el-radio>
                  <el-radio :value="2">女</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="生日" prop="birthday">
                <el-date-picker
                  v-model="profileForm.birthday"
                  type="date"
                  placeholder="选择生日"
                  value-format="YYYY-MM-DD"
                />
              </el-form-item>
              <el-form-item label="个性签名" prop="signature">
                <el-input
                  v-model="profileForm.signature"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入个性签名"
                  maxlength="500"
                  show-word-limit
                />
              </el-form-item>
              <el-form-item label="所在地" prop="location">
                <el-input v-model="profileForm.location" placeholder="请输入所在地" maxlength="100" />
              </el-form-item>
              <el-form-item label="个人网站" prop="website">
                <el-input v-model="profileForm.website" placeholder="请输入个人网站" maxlength="200" />
              </el-form-item>
              <el-form-item label="公司" prop="company">
                <el-input v-model="profileForm.company" placeholder="请输入公司名称" maxlength="100" />
              </el-form-item>
              <el-form-item label="职位" prop="jobTitle">
                <el-input v-model="profileForm.jobTitle" placeholder="请输入职位" maxlength="50" />
              </el-form-item>
              <el-form-item label="教育背景" prop="education">
                <el-input v-model="profileForm.education" placeholder="请输入教育背景" maxlength="50" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSaveProfile" :loading="loading">
                  保存
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getUserProfileInfo, updateUserProfileInfo } from '@/api/user'
import type { UserProfileInfo } from '@/types'
import { useUserStore } from '@/stores/user'
import Layout from '@/components/Layout.vue'

const userStore = useUserStore()
const activeTab = ref('profile')
const profileFormRef = ref<FormInstance>()
const loading = ref(false)

const profileForm = reactive<UserProfileInfo>({
  userId: 0,
  gender: 0,
  birthday: '',
  signature: '',
  location: '',
  website: '',
  company: '',
  jobTitle: '',
  education: ''
})

const rules: FormRules = {
  signature: [
    { max: 500, message: '个性签名不能超过 500 个字符', trigger: 'blur' }
  ],
  location: [
    { max: 100, message: '所在地不能超过 100 个字符', trigger: 'blur' }
  ],
  website: [
    { max: 200, message: '个人网站不能超过 200 个字符', trigger: 'blur' }
  ],
  company: [
    { max: 100, message: '公司名称不能超过 100 个字符', trigger: 'blur' }
  ],
  jobTitle: [
    { max: 50, message: '职位不能超过 50 个字符', trigger: 'blur' }
  ],
  education: [
    { max: 50, message: '教育背景不能超过 50 个字符', trigger: 'blur' }
  ]
}

const fetchProfile = async () => {
  if (!userStore.userInfo) return
  try {
    const res = await getUserProfileInfo(userStore.userInfo.id)
    if (res.profile) {
      Object.assign(profileForm, res.profile)
    }
  } catch (error) {
    console.error(error)
  }
}

const handleSaveProfile = async () => {
  if (!profileFormRef.value) return
  await profileFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await updateUserProfileInfo({ userProfile: profileForm })
        ElMessage.success('保存成功')
      } catch (error) {
        console.error(error)
      } finally {
        loading.value = false
      }
    }
  })
}

onMounted(() => {
  fetchProfile()
})
</script>

<style scoped>
.user-settings-page {
  max-width: 800px;
  margin: 0 auto;
}

.card-header h2 {
  margin: 0;
  font-size: 20px;
}
</style>
