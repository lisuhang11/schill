<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="header-content">
        <div class="logo" @click="router.push('/home')">
          <span>SChill</span>
        </div>
        <div class="nav-menu">
          <el-button text @click="router.push('/home')">首页</el-button>
          <template v-if="userStore.isLoggedIn">
            <el-button text @click="router.push('/create-post')">发布</el-button>
            <el-dropdown @command="handleCommand">
              <span class="user-info">
                <el-avatar :size="32" :src="userStore.userInfo?.avatar || ''">
                  {{ userStore.userInfo?.username?.charAt(0) || 'U' }}
                </el-avatar>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人主页</el-dropdown-item>
                  <el-dropdown-item command="settings">设置</el-dropdown-item>
                  <el-dropdown-item command="following">关注</el-dropdown-item>
                  <el-dropdown-item command="followers">粉丝</el-dropdown-item>
                  <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-button text @click="router.push('/login')">登录</el-button>
            <el-button type="primary" @click="router.push('/register')">注册</el-button>
          </template>
        </div>
      </div>
    </el-header>
    <el-main class="main">
      <slot />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const handleCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      if (userStore.userInfo) {
        router.push(`/user/${userStore.userInfo.id}`)
      }
      break
    case 'settings':
      router.push('/settings')
      break
    case 'following':
      router.push('/following')
      break
    case 'followers':
      router.push('/followers')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        userStore.logout()
        ElMessage.success('退出登录成功')
        router.push('/home')
      } catch {
      }
      break
  }
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
}

.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.logo {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  cursor: pointer;
}

.nav-menu {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
</style>
