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

type SearchPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchPostLogic {
	return &SearchPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ESSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type ESPostDoc struct {
	Id              int64     `json:"id"`
	UserId          int64     `json:"user_id"`
	Title           string    `json:"title"`
	Cover           string    `json:"cover"`
	Summary         string    `json:"summary"`
	Content         string    `json:"content"`
	CommentCount    int64     `json:"comment_count"`
	CollectionCount int64     `json:"collection_count"`
	UpvoteCount     int64     `json:"upvote_count"`
	ShareCount      int64     `json:"share_count"`
	Visibility      int8      `json:"visibility"`
	IsTop           bool      `json:"is_top"`
	IsEssence       bool      `json:"is_essence"`
	IsLock          bool      `json:"is_lock"`
	Tags            []string  `json:"tags"`
	CreatedAt       time.Time `json:"created_at"`
	Username        string    `json:"username"`
	Avatar          string    `json:"avatar"`
	TopicNames      []string  `json:"topic_names"`
}

func (l *SearchPostLogic) SearchPost(req *types.SearchPostReq) (resp *types.SearchPostResp, err error) {
	if req.Keyword == "" {
		return &types.SearchPostResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	cacheKey := redis.SearchResultKey + "post:" + getCacheKey(req.Keyword, req.Page, req.PageSize)
	var cachedResp types.SearchPostResp
	if err = l.svcCtx.Redis.GetJSON(l.ctx, cacheKey, &cachedResp); err == nil {
		return &cachedResp, nil
	}

	from := (req.Page - 1) * req.PageSize
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": req.Keyword,
				"fields": []string{
					"title^4",
					"summary^3",
					"content^2",
					"username^2",
					"topic_names",
					"tags",
				},
			},
		},
		"from": from,
		"size": req.PageSize,
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return &types.SearchPostResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}, nil
	}

	res, err := l.svcCtx.ESClient.Search(l.ctx, l.svcCtx.Config.Elasticsearch.PostIndex, string(queryBytes))
	if err != nil {
		logx.Errorf("search post from ES failed: %v", err)
		return &types.SearchPostResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}, nil
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES search response error: %s", res.String())
		return &types.SearchPostResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &types.SearchPostResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}, nil
	}

	var esResp ESSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		return &types.SearchPostResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}, nil
	}

	list := make([]types.SearchPostItem, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		var doc ESPostDoc
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logx.Errorf("unmarshal ES post doc failed: %v", err)
			continue
		}
		list = append(list, types.SearchPostItem{
			Id:              doc.Id,
			UserId:          doc.UserId,
			Title:           doc.Title,
			Cover:           doc.Cover,
			Summary:         doc.Summary,
			Username:        doc.Username,
			Avatar:          doc.Avatar,
			Content:         doc.Content,
			TopicNames:      doc.TopicNames,
			CommentCount:    doc.CommentCount,
			CollectionCount: doc.CollectionCount,
			UpvoteCount:     doc.UpvoteCount,
			ShareCount:      doc.ShareCount,
			Visibility:      doc.Visibility,
			IsTop:           doc.IsTop,
			IsEssence:       doc.IsEssence,
			IsLock:          doc.IsLock,
			Tags:            doc.Tags,
			CreatedAt:       doc.CreatedAt.Unix(),
		})
	}

	result := &types.SearchPostResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: esResp.Hits.Total.Value,
		List:  list,
	}
	if err := l.svcCtx.Redis.SetJSON(l.ctx, cacheKey, result, time.Duration(redis.SearchExpire)*time.Second); err != nil {
		logx.Errorf("cache search result failed: %v", err)
	}
	return result, nil
}
