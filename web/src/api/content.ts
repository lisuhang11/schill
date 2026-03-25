import request from '@/utils/request'
import type { PostInfo, PostDetail } from '@/types'

export interface CreatePostReq {
  title: string
  cover?: string
  type: number
  content: string
  topics?: string[]
}

export interface CreatePostResp {
  postId: number
}

export interface UpdatePostReq {
  postId: number
  title: string
  cover?: string
  type: number
  content: string
  topics?: string[]
}

export interface UpdatePostResp {}

export interface DeletePostResp {}

export interface GetPostListReq {
  userId?: number
  page?: number
  pageSize?: number
}

export interface GetPostListResp {
  total: number
  list: PostInfo[]
}

export interface BatchGetPostReq {
  postIds: number[]
}

export interface BatchGetPostResp {
  posts: PostInfo[]
}

export interface IncViewCountResp {}

export const createPost = (data: CreatePostReq) => {
  return request.post<CreatePostResp>('/api/content/post', data)
}

export const updatePost = (data: UpdatePostReq) => {
  return request.put<UpdatePostResp>('/api/content/post', data)
}

export const deletePost = (postId: number) => {
  return request.delete<DeletePostResp>(`/api/content/post/${postId}`)
}

export const getPostDetail = (postId: number) => {
  return request.get<PostDetail>(`/api/content/post/${postId}`)
}

export const getPostList = (params: GetPostListReq) => {
  return request.get<GetPostListResp>('/api/content/post/list', { params })
}

export const batchGetPost = (data: BatchGetPostReq) => {
  return request.post<BatchGetPostResp>('/api/content/post/batch', data)
}

export const incViewCount = (postId: number) => {
  return request.post<IncViewCountResp>(`/api/content/post/${postId}/view`)
}
