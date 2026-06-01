import type {
  ApiResult,
  BackendEnvelope,
  CommentListResponse,
  CreateCommentRequest,
  CreateCommentResponse,
  CreatePostRequest,
  CreatePostResponse,
  FeedListRequest,
  FeedListResponse,
  FollowListResponse,
  InteractionToggleResponse,
  LoginRequest,
  LoginResponse,
  PageRequest,
  PostDetailResponse,
  PostInfo,
  PostListResponse,
  RegisterRequest,
  RegisterResponse,
  SearchPostItem,
  SearchResponse,
  SearchTopicItem,
  SearchUserItem,
  TopicListResponse,
  UpdatePostRequest,
  VoteCommentResponse
} from "@/lib/types";

const DEFAULT_API_BASE_URL = "http://localhost:8086";

function apiBaseUrl(): string {
  return (
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    process.env.API_BASE_URL ??
    DEFAULT_API_BASE_URL
  ).replace(/\/$/, "");
}

function buildUrl(path: string, query?: Record<string, string | number | undefined>) {
  const url = new URL(`${apiBaseUrl()}${path}`);
  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value !== undefined && String(value) !== "") {
      url.searchParams.set(key, String(value));
    }
  });
  return url;
}

function authHeaders(): Record<string, string> {
  if (typeof window === "undefined") {
    return {};
  }
  const accessToken = window.localStorage.getItem("schill:accessToken");
  return accessToken ? { Authorization: `Bearer ${accessToken}` } : {};
}

async function request<T>(
  path: string,
  init?: RequestInit & { query?: Record<string, string | number | undefined> }
): Promise<ApiResult<T>> {
  const { query, ...requestInit } = init ?? {};

  try {
    const response = await fetch(buildUrl(path, query), {
      cache: "no-store",
      ...requestInit,
      headers: {
        ...(requestInit.body ? { "Content-Type": "application/json" } : {}),
        ...authHeaders(),
        ...requestInit.headers
      }
    });

    if (!response.ok) {
      return { ok: false, message: `HTTP ${response.status}`, status: response.status };
    }

    const payload = (await response.json()) as (T & BackendEnvelope) | BackendEnvelope;
    if (typeof payload.code === "number" && payload.code !== 0) {
      return { ok: false, message: payload.msg || `业务错误 ${payload.code}` };
    }

    const data = Object.prototype.hasOwnProperty.call(payload, "data")
      ? (payload.data as T)
      : (payload as T);

    return { ok: true, data };
  } catch (error) {
    return {
      ok: false,
      message: error instanceof Error ? error.message : "请求失败"
    };
  }
}

export function registerUser(input: RegisterRequest) {
  return request<RegisterResponse>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function loginUser(input: LoginRequest) {
  return request<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function getPostList(params: PageRequest & { userId?: number }) {
  return request<PostListResponse>("/api/posts", {
    query: {
      userId: params.userId,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10
    }
  });
}

export function getPostDetail(postId: number) {
  return request<PostDetailResponse>(`/api/posts/${postId}`);
}

export function createPost(input: CreatePostRequest) {
  return request<CreatePostResponse>("/api/posts", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function updatePost(input: UpdatePostRequest) {
  return request<BackendEnvelope>(`/api/posts/${input.postId}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export function getTopicList(params: PageRequest & { sort?: "hot" | "new" }) {
  return request<TopicListResponse>("/api/topics", {
    query: {
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 12,
      sort: params.sort ?? "hot"
    }
  });
}

export function getCommentList(params: {
  postId: number;
  cursor?: number;
  pageSize?: number;
  sortType?: "time" | "hot" | "new";
}) {
  return request<CommentListResponse>(`/api/posts/${params.postId}/comments`, {
    query: {
      cursor: params.cursor ?? 0,
      pageSize: params.pageSize ?? 20,
      sortType: params.sortType ?? "time"
    }
  });
}

export function createComment(input: CreateCommentRequest) {
  return request<CreateCommentResponse>("/api/comments", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function voteComment(commentId: number, voteType: 1 | 2) {
  return request<VoteCommentResponse>(`/api/comments/${commentId}/vote`, {
    method: "POST",
    body: JSON.stringify({ voteType })
  });
}

export function starPost(postId: number) {
  return request<InteractionToggleResponse>(`/api/posts/${postId}/star`, {
    method: "POST"
  });
}

export function unstarPost(postId: number) {
  return request<InteractionToggleResponse>(`/api/posts/${postId}/star`, {
    method: "DELETE"
  });
}

export function collectPost(postId: number) {
  return request<InteractionToggleResponse>(`/api/posts/${postId}/collect`, {
    method: "POST"
  });
}

export function uncollectPost(postId: number) {
  return request<InteractionToggleResponse>(`/api/posts/${postId}/collect`, {
    method: "DELETE"
  });
}

export function sharePost(postId: number) {
  return request<InteractionToggleResponse>(`/api/posts/${postId}/share`, {
    method: "POST"
  });
}

export function getFollowingList(params: PageRequest & { userId?: number }) {
  return request<FollowListResponse>("/api/relation/following", {
    query: {
      userId: params.userId,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 20
    }
  });
}

export function getFollowerList(params: PageRequest & { userId?: number }) {
  return request<FollowListResponse>("/api/relation/follower", {
    query: {
      userId: params.userId,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 20
    }
  });
}

export function getFeedList(params: FeedListRequest) {
  return request<FeedListResponse>("/api/feed", {
    query: {
      feedType: params.feedType,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10,
      currentUserId: params.currentUserId
    }
  });
}

export function searchPosts(params: PageRequest & { keyword: string }) {
  return request<SearchResponse<SearchPostItem>>("/api/search/post", {
    query: {
      keyword: params.keyword,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10
    }
  });
}

export function searchUsers(params: PageRequest & { keyword: string }) {
  return request<SearchResponse<SearchUserItem>>("/api/search/user", {
    query: {
      keyword: params.keyword,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10
    }
  });
}

export function searchTopics(params: PageRequest & { keyword: string }) {
  return request<SearchResponse<SearchTopicItem>>("/api/search/topic", {
    query: {
      keyword: params.keyword,
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10
    }
  });
}

export function getUserInfo(userId: number) {
  return request<{
    userInfo: {
      id: number;
      username: string;
      avatar: string;
      status: number;
      createdAt: number;
    };
    profile: {
      userId: number;
      gender: number;
      signature: string;
      location: string;
      website: string;
      company: string;
      jobTitle: string;
      education: string;
    };
    stat: {
      userId: number;
      postCount?: number;
      commentCount?: number;
      followerCount?: number;
      followingCount?: number;
      likeCount?: number;
      collectionCount?: number;
      lastActiveTime?: number;
    };
  }>(`/api/users/${userId}`);
}

export function getUserProfile(userId: number) {
  return getUserInfo(userId);
}

export function getMyCollections(params: PageRequest) {
  return request<{ total: number; list: PostInfo[] }>("/api/users/me/collections", {
    query: {
      page: params.page ?? 1,
      pageSize: params.pageSize ?? 10
    }
  });
}

export function getUserStat(userId: number) {
  return getUserInfo(userId);
}
