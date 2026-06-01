package logic

import (
	"context"
	"encoding/json"
	"io"
	"time"

	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/search/rpc/internal/svc"
	"SChill/service/search/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type ESUserDoc struct {
	Id              int64     `json:"id"`
	Username        string    `json:"username"`
	Status          int8      `json:"status"`
	IsAdmin         bool      `json:"is_admin"`
	CreatedAt       time.Time `json:"created_at"`
	PostCount       int32     `json:"post_count"`
	CommentCount    int32     `json:"comment_count"`
	FollowerCount   int32     `json:"follower_count"`
	LikeCount       int32     `json:"like_count"`
	CollectionCount int32     `json:"collection_count"`
	LastActiveTime  int64     `json:"last_active_time"`
}

func (l *SearchUserLogic) SearchUser(in *pb.SearchReq) (*pb.SearchUserResp, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize)
	if in.Keyword == "" {
		return &pb.SearchUserResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	cacheKey := redis.SearchResultKey + "user:" + getCacheKey(in.Keyword, page, pageSize)
	var cachedResp pb.SearchUserResp
	if err := l.svcCtx.Redis.GetJSON(l.ctx, cacheKey, &cachedResp); err == nil {
		return &cachedResp, nil
	}

	from := (page - 1) * pageSize
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  in.Keyword,
				"fields": []string{"username", "username.ngram"},
			},
		},
		"from": from,
		"size": pageSize,
		"sort": []map[string]interface{}{
			{"follower_count": map[string]string{"order": "desc"}},
			{"last_active_time": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return internalSearchUserResp(), nil
	}

	res, err := l.svcCtx.ESClient.Search(l.ctx, l.svcCtx.Config.Elasticsearch.UserIndex, string(queryBytes))
	if err != nil {
		logx.Errorf("search user from ES failed: %v", err)
		return internalSearchUserResp(), nil
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES search response error: %s", res.String())
		return internalSearchUserResp(), nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return internalSearchUserResp(), nil
	}

	var esResp ESSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		return internalSearchUserResp(), nil
	}

	list := make([]*pb.SearchUserItem, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		var doc ESUserDoc
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logx.Errorf("unmarshal ES user doc failed: %v", err)
			continue
		}

		list = append(list, &pb.SearchUserItem{
			Id:              doc.Id,
			Username:        doc.Username,
			Status:          int32(doc.Status),
			IsAdmin:         doc.IsAdmin,
			PostCount:       doc.PostCount,
			CommentCount:    doc.CommentCount,
			FollowerCount:   doc.FollowerCount,
			LikeCount:       doc.LikeCount,
			CollectionCount: doc.CollectionCount,
			LastActiveTime:  doc.LastActiveTime,
			CreatedAt:       doc.CreatedAt.Unix(),
		})
	}

	result := &pb.SearchUserResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: esResp.Hits.Total.Value,
		List:  list,
	}
	if err := l.svcCtx.Redis.SetJSON(l.ctx, cacheKey, result, time.Duration(redis.SearchExpire)*time.Second); err != nil {
		logx.Errorf("cache search user result failed: %v", err)
	}

	return result, nil
}

func internalSearchUserResp() *pb.SearchUserResp {
	return &pb.SearchUserResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}
}
