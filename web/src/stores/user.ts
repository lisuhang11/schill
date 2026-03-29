import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo } from '@/types'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>('')
  const refreshToken = ref<string>('')
  const userId = ref<number>(0)
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setRefreshToken = (newToken: string) => {
    refreshToken.value = newToken
    localStorage.setItem('refreshToken', newToken)
  }

  const setUserId = (id: number) => {
    userId.value = id
  }

  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
  }

  const logout = () => {
    token.value = ''
    refreshToken.value = ''
    userId.value = 0
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  const init = () => {
    const savedToken = localStorage.getItem('token')
    const savedRefreshToken = localStorage.getItem('refreshToken')
    if (savedToken) token.value = savedToken
    if (savedRefreshToken) refreshToken.value = savedRefreshToken
  }

  return {
    token,
    refreshToken,
    userId,
    userInfo,
    isLoggedIn,
    setToken,
    setRefreshToken,
    setUserId,
    setUserInfo,
    logout,
    init
  }
})
