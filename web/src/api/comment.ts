import request from '@/utils/request'
import type { CommentInfo, CommentItem } from '@/types'

export interface CreateCommentReq {
  groupId: number
  parentId?: number
  replyToUserId?: number
  content: string
}

export interface CreateCommentResp {
  comment: CommentInfo
}

export interface DeleteCommentResp {}

export interface VoteCommentReq {
  voteType: number
}

export interface VoteCommentResp {
  likeCount: number
  dislikeCount: number
  isLiked: boolean
  isDisliked: boolean
}

export interface GetCommentListReq {
  groupId: number
  cursor?: number
  pageSize?: number
  sortType?: string
}

export interface GetCommentListResp {
  total: number
  list: CommentItem[]
}

export interface GetReplyListReq {
  commentId: number
  cursor?: number
  pageSize?: number
}

export const createComment = (data: CreateCommentReq) => {
  return request.post<CreateCommentResp>('/api/comment', data)
}

export const deleteComment = (commentId: number) => {
  return request.delete<DeleteCommentResp>(`/api/comment/${commentId}`)
}

export const voteComment = (commentId: number, data: VoteCommentReq) => {
  return request.post<VoteCommentResp>(`/api/comment/${commentId}/vote`, data)
}

export const getCommentList = (params: GetCommentListReq) => {
  return request.get<GetCommentListResp>('/api/comment/list', { params })
}

export const getReplyList = (params: GetReplyListReq) => {
  return request.get<GetCommentListResp>('/api/comment/replies', { params })
}
