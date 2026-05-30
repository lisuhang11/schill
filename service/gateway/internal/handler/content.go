package handler

import (
	"net/http"

	"SChill/service/content/rpc/contentcenter"
	"SChill/service/gateway/internal/svc"
	"SChill/service/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPostListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, _ := currentUserID(r)
		resp, err := svcCtx.ContentRpc.GetPostList(r.Context(), &contentcenter.GetPostListReq{
			UserId:        uintQueryParam(r, "userId", 0),
			Page:          pageParam(r, "page", 1),
			PageSize:      pageParam(r, "pageSize", 20),
			CurrentUserId: currentUserID,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func CreatePostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.CreatePostReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.ContentRpc.CreatePost(r.Context(), &contentcenter.CreatePostReq{
			UserId:     userID,
			Title:      req.Title,
			Cover:      req.Cover,
			Visibility: req.Visibility,
			Contents:   toRpcContents(req.Contents),
			Topics:     req.Topics,
			Tags:       req.Tags,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func GetPostDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id")
			return
		}
		currentUserID, _ := currentUserID(r)
		resp, err := svcCtx.ContentRpc.GetPostDetail(r.Context(), &contentcenter.GetPostDetailReq{
			PostId:        postID,
			CurrentUserId: currentUserID,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func UpdatePostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		var req types.UpdatePostReq
		if err := httpx.Parse(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := svcCtx.ContentRpc.UpdatePost(r.Context(), &contentcenter.UpdatePostReq{
			PostId:     postID,
			UserId:     userID,
			Title:      req.Title,
			Cover:      req.Cover,
			Visibility: req.Visibility,
			Contents:   toRpcContents(req.Contents),
			Topics:     req.Topics,
			Tags:       req.Tags,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func DeletePostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, valid := parseUintParam(r, "id")
		if !valid {
			fail(w, http.StatusBadRequest, "invalid post id")
			return
		}
		userID, valid := currentUserID(r)
		if !valid {
			fail(w, http.StatusUnauthorized, "missing X-User-Id")
			return
		}
		resp, err := svcCtx.ContentRpc.DeletePost(r.Context(), &contentcenter.DeletePostReq{
			PostId: postID,
			UserId: userID,
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func GetTopicListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := svcCtx.ContentRpc.GetTopicList(r.Context(), &contentcenter.GetTopicListReq{
			Page:     pageParam(r, "page", 1),
			PageSize: pageParam(r, "pageSize", 20),
			Sort:     r.URL.Query().Get("sort"),
		})
		if err != nil {
			fail(w, http.StatusBadGateway, err.Error())
			return
		}
		ok(w, resp)
	}
}

func toRpcContents(items []types.PostContentItem) []*contentcenter.PostContentItem {
	contents := make([]*contentcenter.PostContentItem, 0, len(items))
	for _, item := range items {
		contents = append(contents, &contentcenter.PostContentItem{
			Type:    item.Type,
			Content: item.Content,
			Sort:    item.Sort,
		})
	}
	return contents
}
