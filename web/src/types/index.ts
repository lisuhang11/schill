export interface ApiResponse<T = any> {
  code: number
  msg: string
  data?: T
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  phone: string
  email: string
  avatar: string
  status: number
  isAdmin: boolean
  createdAt: number
}

export interface UserProfileInfo {
  userId: number
  gender: number
  birthday?: string
  signature?: string
  location?: string
  website?: string
  company?: string
  jobTitle?: string
  education?: string
}

export interface UserStatInfo {
  userId: number
  postCount: number
  commentCount: number
  followerCount: number
  followingCount: number
  likeCount: number
  collectionCount: number
  lastActiveTime: number
}

export interface PostInfo {
  id: number
  userId: number
  title: string
  cover: string
  type: number
  status: number
  viewCount: number
  likeCount: number
  commentCount: number
  createdAt: number
  updatedAt: number
}

export interface PostTopic {
  postId: number
  topicId: number
  topicName: string
}

export interface PostDetail {
  post: PostInfo
  content: string
  topics: PostTopic[]
}

export interface CommentInfo {
  id: number
  groupId: number
  userId: number
  parentId: number
  replyToUserId: number
  content: string
  level: number
  replyCount: number
  likeCount: number
  createdAt: number
  username: string
  avatar: string
  replyToUsername: string
  isLiked: boolean
}

export interface CommentItem {
  root: CommentInfo
  replies: CommentInfo[]
  hasMoreReplies: boolean
}

export interface FollowInfo {
  userId: number
  username: string
  avatar: string
  followTime: number
}
