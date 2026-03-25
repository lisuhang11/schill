import request from '@/utils/request'
import type { FollowInfo } from '@/types'

export interface FollowReq {
  targetUserId: number
}

export interface FollowResp {}

export interface UnfollowResp {}

export interface FollowStatusResp {
  isFollow: boolean
}

export interface GetFollowingListReq {
  userId?: number
  page?: number
  pageSize?: number
}

export interface GetFollowerListReq {
  userId?: number
  page?: number
  pageSize?: number
}

export interface FollowListResp {
  total: number
  list: FollowInfo[]
}

export interface CheckFollowStatusReq {
  targetUserId: number
}

export interface BatchCheckFollowStatusReq {
  targetUserIds: number[]
}

export interface FollowStatusItem {
  userId: number
  isFollow: boolean
}

export interface BatchCheckFollowStatusResp {
  status: FollowStatusItem[]
}

export const follow = (data: FollowReq) => {
  return request.post<FollowResp>('/api/relation/follow', data)
}

export const unfollow = (targetUserId: number) => {
  return request.delete<UnfollowResp>(`/api/relation/follow/${targetUserId}`)
}

export const getFollowingList = (params: GetFollowingListReq) => {
  return request.get<FollowListResp>('/api/relation/following', { params })
}

export const getFollowerList = (params: GetFollowerListReq) => {
  return request.get<FollowListResp>('/api/relation/follower', { params })
}

export const checkFollowStatus = (params: CheckFollowStatusReq) => {
  return request.get<FollowStatusResp>('/api/relation/follow/status', { params })
}

export const batchCheckFollowStatus = (data: BatchCheckFollowStatusReq) => {
  return request.post<BatchCheckFollowStatusResp>('/api/relation/follow/status/batch', data)
}
