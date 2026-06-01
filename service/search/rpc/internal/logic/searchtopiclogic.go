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

type SearchTopicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchTopicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchTopicLogic {
	return &SearchTopicLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type ESTopicDoc struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	QuoteNum  int64     `json:"quote_num"`
	CreatedAt time.Time `json:"created_at"`
}

func (l *SearchTopicLogic) SearchTopic(in *pb.SearchReq) (*pb.SearchTopicResp, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize)
	if in.Keyword == "" {
		return &pb.SearchTopicResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	cacheKey := redis.SearchResultKey + "topic:" + getCacheKey(in.Keyword, page, pageSize)
	var cachedResp pb.SearchTopicResp
	if err := l.svcCtx.Redis.GetJSON(l.ctx, cacheKey, &cachedResp); err == nil {
		return &cachedResp, nil
	}

	from := (page - 1) * pageSize
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  in.Keyword,
				"fields": []string{"name"},
			},
		},
		"from": from,
		"size": pageSize,
		"sort": []map[string]interface{}{
			{"quote_num": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return internalSearchTopicResp(), nil
	}

	res, err := l.svcCtx.ESClient.Search(l.ctx, l.svcCtx.Config.Elasticsearch.TopicIndex, string(queryBytes))
	if err != nil {
		logx.Errorf("search topic from ES failed: %v", err)
		return internalSearchTopicResp(), nil
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("ES search response error: %s", res.String())
		return internalSearchTopicResp(), nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return internalSearchTopicResp(), nil
	}

	var esResp ESSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		return internalSearchTopicResp(), nil
	}

	list := make([]*pb.SearchTopicItem, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		var doc ESTopicDoc
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logx.Errorf("unmarshal ES topic doc failed: %v", err)
			continue
		}

		list = append(list, &pb.SearchTopicItem{
			Id:        doc.Id,
			Name:      doc.Name,
			QuoteNum:  doc.QuoteNum,
			CreatedAt: doc.CreatedAt.Unix(),
		})
	}

	result := &pb.SearchTopicResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: esResp.Hits.Total.Value,
		List:  list,
	}
	if err := l.svcCtx.Redis.SetJSON(l.ctx, cacheKey, result, time.Duration(redis.SearchExpire)*time.Second); err != nil {
		logx.Errorf("cache search topic result failed: %v", err)
	}

	return result, nil
}

func internalSearchTopicResp() *pb.SearchTopicResp {
	return &pb.SearchTopicResp{Code: errutil.ErrInternalError, Msg: errutil.GetCodeMessage(errutil.ErrInternalError)}
}
