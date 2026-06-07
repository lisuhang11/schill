package handler

import (
	"net/http"

	"SChill/service/comment/rpc/commentcenter"
	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"
	userpb "SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// gatewayCommentInfo enriches a CommentInfo with user names and avatars.
type gatewayCommentInfo struct {
	Id              uint64 `json:"id"`
	PostId          uint64 `json:"postId"`
	UserId          uint64 `json:"userId"`
	ParentId        uint64 `json:"parentId"`
	ReplyToUserId   uint64 `json:"replyToUserId"`
	Content         string `json:"content"`
	Level           int32  `json:"level"`
	ReplyCount      int32  `json:"replyCount"`
	LikeCount       int32  `json:"likeCount"`
	DislikeCount    int32  `json:"dislikeCount"`
	CreatedAt       int64  `json:"createdAt"`
	Username        string `json:"username"`
	Avatar          string `json:"avatar"`
	ReplyToUsername string `json:"replyToUsername"`
	IsLiked         bool   `json:"isLiked"`
	IsDisliked      bool   `json:"isDisliked"`
}

// gatewayCommentItem groups a root comment with its top replies.
type gatewayCommentItem struct {
	Root           gatewayCommentInfo   `json:"root"`
	Replies        []gatewayCommentInfo `json:"replies"`
	HasMoreReplies bool                 `json:"hasMoreReplies"`
}

// buildUserMap calls user-rpc to batch-fetch basic info and returns a map userID -> *UserBasicInfo.
func buildUserMap(ctx *svc.ServiceContext, userIDs []uint64, r *http.Request) map[uint64]*userpb.UserBasicInfo {
	result := make(map[uint64]*userpb.UserBasicInfo, len(userIDs))
	if len(userIDs) == 0 {
		return result
	}
	resp, err := ctx.UserRpc.BatchGetUserBasicInfo(r.Context(), &userpb.BatchGetUserBasicInfoReq{
		UserIds: userIDs,
	})
	if err != nil {
		return result
	}
	for _, u := range resp.Users {
		result[u.UserId] = u
	}
	return result
}

// collectCommentUserIDs gathers all userId and replyToUserId from a flat CommentInfo list.
func collectCommentUserIDs(list []*commentcenter.CommentInfo) []uint64 {
	seen := make(map[uint64]bool)
	for _, c := range list {
		if c.UserId > 0 {
			seen[c.UserId] = true
		}
		if c.ReplyToUserId > 0 {
			seen[c.ReplyToUserId] = true
		}
	}
	ids := make([]uint64, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

// enrichCommentInfo fills username/avatar/replyToUsername from user map.
func enrichCommentInfo(c *commentcenter.CommentInfo, userMap map[uint64]*userpb.UserBasicInfo) gatewayCommentInfo {
	info := gatewayCommentInfo{
		Id:            c.Id,
		PostId:        c.PostId,
		UserId:        c.UserId,
		ParentId:      c.ParentId,
		ReplyToUserId: c.ReplyToUserId,
		Content:       c.Content,
		Level:         c.Level,
		ReplyCount:    c.ReplyCount,
		LikeCount:     c.LikeCount,
		DislikeCount:  c.DislikeCount,
		CreatedAt:     c.CreatedAt,
	}
	if u, ok := userMap[c.UserId]; ok {
		info.Username = u.Username
		info.Avatar = u.Avatar
	}
	if c.ReplyToUserId > 0 {
		if u, ok := userMap[c.ReplyToUserId]; ok {
			info.ReplyToUsername = u.Username
		}
	}
	return info
}

func GetCommentListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id")
			return
		}
		resp, err := svcCtx.CommentRpc.GetCommentList(r.Context(), &commentcenter.GetCommentListReq{
			PostId:   postID,
			Cursor:   pageParam(r, "cursor", 0),
			PageSize: pageParam(r, "pageSize", 20),
			SortType: r.URL.Query().Get("sortType"),
		})
		if err != nil {
			rpcFail(w, err)
			return
		}

		// Collect all user IDs from the flat comment list
		userIDs := collectCommentUserIDs(resp.List)
		userMap := buildUserMap(svcCtx, userIDs, r)

		// Assemble tree: root comments become items, and we fetch top replies
		items := make([]gatewayCommentItem, 0, len(resp.List))
		for _, root := range resp.List {
			replies := make([]gatewayCommentInfo, 0)
			hasMoreReplies := false

			if root.ReplyCount > 0 {
				replyResp, replyErr := svcCtx.CommentRpc.GetReplyList(r.Context(), &commentcenter.GetReplyListReq{
					CommentId: root.Id,
					Cursor:    0,
					PageSize:  3,
				})
				if replyErr == nil {
					for _, rp := range replyResp.List {
						replies = append(replies, enrichCommentInfo(rp, userMap))
					}
					hasMoreReplies = replyResp.HasMore
				}
			}

			items = append(items, gatewayCommentItem{
				Root:           enrichCommentInfo(root, userMap),
				Replies:        replies,
				HasMoreReplies: hasMoreReplies,
			})
		}

		ok(w, &struct {
			Total      int64                 `json:"total"`
			List       []gatewayCommentItem  `json:"list"`
			HasMore    bool                  `json:"hasMore"`
			NextCursor int64                 `json:"nextCursor"`
		}{
			Total:      resp.Total,
			List:       items,
			HasMore:    resp.HasMore,
			NextCursor: resp.NextCursor,
		})
	}
}

func GetReplyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		resp, err := svcCtx.CommentRpc.GetReplyList(r.Context(), &commentcenter.GetReplyListReq{
			CommentId: commentID,
			Cursor:    pageParam(r, "cursor", 0),
			PageSize:  pageParam(r, "pageSize", 20),
		})
		if err != nil {
			rpcFail(w, err)
			return
		}

		userIDs := collectCommentUserIDs(resp.List)
		userMap := buildUserMap(svcCtx, userIDs, r)

		list := make([]gatewayCommentInfo, 0, len(resp.List))
		for _, c := range resp.List {
			list = append(list, enrichCommentInfo(c, userMap))
		}

		ok(w, &struct {
			Total      int64                  `json:"total"`
			List       []gatewayCommentInfo   `json:"list"`
			HasMore    bool                   `json:"hasMore"`
			NextCursor int64                  `json:"nextCursor"`
		}{
			Total:      resp.Total,
			List:       list,
			HasMore:    resp.HasMore,
			NextCursor: resp.NextCursor,
		})
	}
}

func CreateCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.CreateCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.CommentRpc.CreateComment(r.Context(), &commentcenter.CreateCommentReq{
			UserId:        userID,
			PostId:        req.PostId,
			ParentId:      req.ParentId,
			ReplyToUserId: req.ReplyToUserId,
			Content:       req.Content,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func DeleteCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		resp, err := svcCtx.CommentRpc.DeleteComment(r.Context(), &commentcenter.DeleteCommentReq{
			CommentId: commentID,
			UserId:    userID,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}

func VoteCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.VoteCommentReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.CommentRpc.VoteComment(r.Context(), &commentcenter.VoteCommentReq{
			CommentId: commentID,
			UserId:    userID,
			VoteType:  req.VoteType,
		})
		if err != nil {
			rpcFail(w, err)
			return
		}
		ok(w, resp)
	}
}
