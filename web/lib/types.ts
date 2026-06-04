export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; message: string; status?: number };

export type BackendEnvelope = {
  code?: number;
  msg?: string;
  data?: unknown;
};

export type PageRequest = {
  page?: number;
  pageSize?: number;
};

export type PostVisibility = 0 | 10 | 20 | 50 | 90;

export const POST_VISIBILITY_OPTIONS: Array<{
  value: PostVisibility;
  label: string;
  description: string;
}> = [
  { value: 90, label: "公开", description: "所有人可见" },
  { value: 20, label: "粉丝", description: "仅粉丝可见" },
  { value: 50, label: "互关", description: "仅互相关注可见" },
  { value: 10, label: "充电", description: "付费或充电后可见" },
  { value: 0, label: "私密", description: "仅自己可见" }
];

export type PostContentType = 2;

export type PostContentItem = {
  type: PostContentType;
  content: string;
  sort: number;
};

export type PostInfo = {
  id: number;
  userId: number;
  title: string;
  cover: string;
  summary?: string;
  commentCount: number;
  collectionCount: number;
  upvoteCount: number;
  shareCount: number;
  visibility: PostVisibility | number;
  isTop: number | boolean;
  isEssence: number | boolean;
  isLock: number | boolean;
  latestRepliedAt: number;
  tags: string;
  createdAt: number;
  updatedAt: number;
};

export type PostTopic = {
  postId: number;
  topicId: number;
  topicName: string;
};

export type PostListResponse = BackendEnvelope & {
  total: number;
  list: PostInfo[];
};

export type PostDetailResponse = BackendEnvelope & {
  post: PostInfo;
  contents: PostContentItem[];
  topics: PostTopic[];
};

export type CreatePostRequest = {
  title: string;
  cover: string;
  visibility: PostVisibility;
  contents: PostContentItem[];
  topics: string[];
  tags: string;
};

export type UpdatePostRequest = CreatePostRequest & {
  postId: number;
};

export type CreatePostResponse = BackendEnvelope & {
  postId: number;
};

export type RegisterRequest = {
  username: string;
  password: string;
};

export type RegisterResponse = {
  userId: number;
};

export type LoginRequest = {
  username: string;
  password: string;
};

export type LoginResponse = {
  userId: number;
  accessToken: string;
  accessExpireIn: number;
  refreshToken: string;
  refreshExpireIn: number;
};

export type TopicInfo = {
  id: number;
  name: string;
  quoteNum: number;
  createdAt: number;
  updatedAt: number;
};

export type TopicListResponse = BackendEnvelope & {
  total: number;
  list: TopicInfo[];
};

export type CommentInfo = {
  id: number;
  postId: number;
  userId: number;
  parentId: number;
  replyToUserId: number;
  content: string;
  level: number;
  replyCount: number;
  likeCount: number;
  dislikeCount: number;
  createdAt: number;
  username: string;
  avatar: string;
  replyToUsername: string;
  isLiked: boolean;
  isDisliked: boolean;
};

export type CommentItem = {
  root: CommentInfo;
  replies: CommentInfo[];
  hasMoreReplies: boolean;
};

export type CommentListResponse = BackendEnvelope & {
  total: number;
  list: CommentItem[];
};

export type CreateCommentRequest = {
  postId: number;
  parentId: number;
  replyToUserId: number;
  content: string;
};

export type CreateCommentResponse = BackendEnvelope & {
  comment: CommentInfo;
};

export type VoteCommentResponse = BackendEnvelope & {
  likeCount: number;
  dislikeCount: number;
  isLiked: boolean;
  isDisliked: boolean;
};

export type InteractionToggleResponse = BackendEnvelope & {
  isStarred?: boolean;
  isCollected?: boolean;
  starCount?: number;
  shareCount?: number;
};

export type PostStarStatusResponse = BackendEnvelope & {
  isStarred: boolean;
  starCount: number;
};

export type PostCollectionStatusResponse = BackendEnvelope & {
  isCollected: boolean;
};

export type FollowInfo = {
  userId: number;
  username: string;
  avatar: string;
  followTime: number;
};

export type FollowListResponse = BackendEnvelope & {
  total: number;
  list: FollowInfo[];
};

export type FeedAuthor = {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
};

export type FeedViewerState = {
  isStarred: boolean;
  isCollected: boolean;
  isFollowingAuthor: boolean;
};

export type FeedItem = {
  postId: number;
  authorId: number;
  title: string;
  summary: string;
  cover: string;
  tags: string[];
  visibility: PostVisibility | number;
  isTop: boolean;
  isEssence: boolean;
  isLock: boolean;
  commentCount: number;
  upvoteCount: number;
  collectionCount: number;
  shareCount: number;
  createdAt: number;
  updatedAt: number;
  latestRepliedAt: number;
  feedType: string;
  source: string;
  author: FeedAuthor;
  viewerState: FeedViewerState;
};

export type FeedListRequest = PageRequest & {
  feedType: "latest" | "following" | string;
  currentUserId?: number;
};

export type FeedListResponse = BackendEnvelope & {
  total: number;
  list: FeedItem[];
};

export type SearchPostItem = {
  id: number;
  userId: number;
  username: string;
  avatar: string;
  content: string;
  topicNames: string[];
  commentCount: number;
  collectionCount: number;
  upvoteCount: number;
  shareCount: number;
  visibility: PostVisibility | number;
  isTop: boolean;
  isEssence: boolean;
  isLock: boolean;
  tags: string[];
  createdAt: number;
};

export type SearchUserItem = {
  id: number;
  username: string;
  status: number;
  isAdmin: boolean;
  postCount: number;
  commentCount: number;
  followerCount: number;
  likeCount: number;
  collectionCount: number;
  lastActiveTime: number;
  createdAt: number;
};

export type SearchTopicItem = {
  id: number;
  name: string;
  quoteNum: number;
  createdAt: number;
};

export type SearchResponse<T> = BackendEnvelope & {
  total: number;
  list: T[];
};

export type UpdateProfileRequest = {
  gender?: number;
  birthday?: string;
  signature?: string;
  location?: string;
  website?: string;
  company?: string;
  jobTitle?: string;
  education?: string;
};

export type ApiGap = {
  capability: string;
  expectedRoute: string;
  currentEvidence: string;
  impact: string;
};
