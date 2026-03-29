// 用户相关类型
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

// 内容相关类型
export interface PostContentItem {
  type: number
  content: string
  sort: number
}

export interface PostTopic {
  postId: number
  topicId: number
  topicName: string
}

export interface PostInfo {
  id: number
  userId: number
  title: string
  cover: string
  commentCount: number
  collectionCount: number
  upvoteCount: number
  shareCount: number
  visibility: number
  isTop: number
  isEssence: number
  isLock: number
  latestRepliedAt: number
  tags: string
  createdAt: number
  updatedAt: number
}

export interface PostDetail {
  post: PostInfo
  contents: PostContentItem[]
  topics: PostTopic[]
}

// 评论相关类型
export interface CommentInfo {
  id: number
  postId: number
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

// 关系相关类型
export interface FollowInfo {
  userId: number
  username: string
  avatar: string
  followTime: number
}

export interface FollowStatusItem {
  userId: number
  isFollow: boolean
}
