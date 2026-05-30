package logic

import (
	"context"
	"encoding/json"
	"io"
	"time"

	errutil "SChill/common/error"
	"SChill/common/redis"
	"SChill/service/search/api/internal/svc"
	"SChill/service/search/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
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

func (l *SearchUserLogic) SearchUser(req *types.SearchUserReq) (resp *types.SearchUserResp, err error) {
	if req.Keyword == "" {
		return &types.SearchUserResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	cacheKey := redis.SearchResultKey + "user:" + getCacheKey(req.Keyword, req.Page, req.PageSize)
	var cachedResp types.SearchUserResp
	err = l.svcCtx.Redis.GetJSON(l.ctx, cacheKey, &cachedResp)
	if err == nil {
		logx.Infof("从缓存获取用户搜索结果成功: keyword=%s", req.Keyword)
		return &cachedResp, nil
	}

	logx.Infof("缓存未命中，从 ES 获取用户搜索结果: keyword=%s", req.Keyword)

	from := (req.Page - 1) * req.PageSize

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Keyword,
				"fields": []string{"username", "username.ngram"},
			},
		},
		"from": from,
		"size": req.PageSize,
		"sort": []map[string]interface{}{
			{"follower_count": map[string]string{"order": "desc"}},
			{"last_active_time": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		logx.Errorf("序列化查询条件失败: %v", err)
		return &types.SearchUserResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	res, err := l.svcCtx.ESClient.Search(l.ctx, l.svcCtx.Config.Elasticsearch.UserIndex, string(queryBytes))
	if err != nil {
		logx.Errorf("ES 搜索失败: %v", err)
		return &types.SearchUserResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES 返回错误: %s", res.String())
		return &types.SearchUserResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logx.Errorf("读取 ES 响应失败: %v", err)
		return &types.SearchUserResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	var esResp ESSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		logx.Errorf("解析 ES 响应失败: %v", err)
		return &types.SearchUserResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	list := make([]types.SearchUserItem, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		var doc ESUserDoc
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logx.Errorf("解析文档失败: %v", err)
			continue
		}

		list = append(list, types.SearchUserItem{
			Id:              doc.Id,
			Username:        doc.Username,
			Status:          doc.Status,
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

	result := &types.SearchUserResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: esResp.Hits.Total.Value,
		List:  list,
	}

	err = l.svcCtx.Redis.SetJSON(l.ctx, cacheKey, result, time.Duration(redis.SearchExpire)*time.Second)
	if err != nil {
		logx.Errorf("设置用户搜索缓存失败: %v", err)
	} else {
		logx.Infof("设置用户搜索缓存成功: keyword=%s", req.Keyword)
	}

	return result, nil
}
