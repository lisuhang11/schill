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

type SearchTopicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchTopicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchTopicLogic {
	return &SearchTopicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ESTopicDoc struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	QuoteNum  int64     `json:"quote_num"`
	CreatedAt time.Time `json:"created_at"`
}

func (l *SearchTopicLogic) SearchTopic(req *types.SearchTopicReq) (resp *types.SearchTopicResp, err error) {
	if req.Keyword == "" {
		return &types.SearchTopicResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	cacheKey := redis.SearchResultKey + "topic:" + getCacheKey(req.Keyword, req.Page, req.PageSize)
	var cachedResp types.SearchTopicResp
	err = l.svcCtx.Redis.GetJSON(l.ctx, cacheKey, &cachedResp)
	if err == nil {
		logx.Infof("从缓存获取话题搜索结果成功: keyword=%s", req.Keyword)
		return &cachedResp, nil
	}

	logx.Infof("缓存未命中，从 ES 获取话题搜索结果: keyword=%s", req.Keyword)

	from := (req.Page - 1) * req.PageSize

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Keyword,
				"fields": []string{"name"},
			},
		},
		"from": from,
		"size": req.PageSize,
		"sort": []map[string]interface{}{
			{"quote_num": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		logx.Errorf("序列化查询条件失败: %v", err)
		return &types.SearchTopicResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	res, err := l.svcCtx.ESClient.Search(l.ctx, l.svcCtx.Config.Elasticsearch.TopicIndex, string(queryBytes))
	if err != nil {
		logx.Errorf("ES 搜索失败: %v", err)
		return &types.SearchTopicResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES 返回错误: %s", res.String())
		return &types.SearchTopicResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logx.Errorf("读取 ES 响应失败: %v", err)
		return &types.SearchTopicResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	var esResp ESSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		logx.Errorf("解析 ES 响应失败: %v", err)
		return &types.SearchTopicResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	list := make([]types.SearchTopicItem, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		var doc ESTopicDoc
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logx.Errorf("解析文档失败: %v", err)
			continue
		}

		list = append(list, types.SearchTopicItem{
			Id:        doc.Id,
			Name:      doc.Name,
			QuoteNum:  doc.QuoteNum,
			CreatedAt: doc.CreatedAt.Unix(),
		})
	}

	result := &types.SearchTopicResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: esResp.Hits.Total.Value,
		List:  list,
	}

	err = l.svcCtx.Redis.SetJSON(l.ctx, cacheKey, result, time.Duration(redis.SearchExpire)*time.Second)
	if err != nil {
		logx.Errorf("设置话题搜索缓存失败: %v", err)
	} else {
		logx.Infof("设置话题搜索缓存成功: keyword=%s", req.Keyword)
	}

	return result, nil
}
