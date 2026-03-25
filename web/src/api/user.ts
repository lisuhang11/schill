import request from '@/utils/request'
import type { UserInfo, UserProfileInfo, UserStatInfo } from '@/types'

export interface LoginReq {
  username: string
  password: string
}

export interface LoginResp {
  userId: number
  accessToken: string
  accessExpireIn: number
  refreshToken: string
  refreshExpireIn: number
}

export interface RegisterReq {
  username: string
  password: string
}

export interface RegisterResp {
  userId: number
}

export interface RefreshReq {
  refreshToken: string
}

export interface RefreshResp {
  accessToken: string
  accessExpireIn: number
  refreshToken: string
  refreshExpireIn: number
}

export interface UpdateAvatarReq {
  avatar: string
}

export interface UpdateAvatarResp {
  avatar: string
}

export interface UpdateUserProfileReq {
  userProfile: UserProfileInfo
}

export interface UpdateUserProfileResp {
  userProfile: UserProfileInfo
}

export const login = (data: LoginReq) => {
  return request.post<LoginResp>('/api/user/login', data)
}

export const register = (data: RegisterReq) => {
  return request.post<RegisterResp>('/api/user/register', data)
}

export const refreshToken = (data: RefreshReq) => {
  return request.post<RefreshResp>('/api/user/token/refresh', data)
}

export const getUserInfo = (id: number) => {
  return request.get<{ userInfo: UserInfo }>(`/api/user/${id}`)
}

export const getUserProfileInfo = (id: number) => {
  return request.get<{ profile?: UserProfileInfo }>(`/api/user/${id}/profile`)
}

export const getUserStat = (id: number) => {
  return request.get<{ stat: UserStatInfo }>(`/api/user/${id}/stat`)
}

export const updateAvatar = (data: UpdateAvatarReq) => {
  return request.post<UpdateAvatarResp>('/api/user/avatar', data)
}

export const updateUserProfileInfo = (data: UpdateUserProfileReq) => {
  return request.post<UpdateUserProfileResp>('/api/user/info', data)
}
