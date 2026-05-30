package logic

import (
	"context"
	"strings"
	"time"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTopicListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTopicListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTopicListLogic {
	return &GetTopicListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTopicListLogic) GetTopicList(in *pb.GetTopicListReq) (*pb.GetTopicListResp, error) {
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	cacheKey := buildTopicListLocalCacheKey(page, pageSize, in.Sort)
	if cached, ok := loadLocalCache[*pb.GetTopicListResp](l.svcCtx.LocalCache, cacheKey); ok && cached != nil {
		return cached, nil
	}

	query := l.svcCtx.DBRead.WithContext(l.ctx).Model(&model.Topic{}).Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logx.Errorf("count topics failed: err=%v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	orderBy := "quote_num DESC, created_at DESC"
	if strings.EqualFold(strings.TrimSpace(in.Sort), "new") {
		orderBy = "created_at DESC, id DESC"
	}

	var topics []model.Topic
	if err := query.
		Order(orderBy).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&topics).Error; err != nil {
		logx.Errorf("load topic list failed: err=%v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	resp := &pb.GetTopicListResp{
		Total: total,
		List:  make([]*pb.TopicInfo, 0, len(topics)),
	}
	for _, topic := range topics {
		resp.List = append(resp.List, &pb.TopicInfo{
			Id:        topic.ID,
			Name:      topic.Name,
			QuoteNum:  topic.QuoteNum,
			CreatedAt: topic.CreatedAt.Unix(),
			UpdatedAt: topic.UpdatedAt.Unix(),
		})
	}
	storeLocalCache(l.svcCtx, cacheKey, resp, time.Minute)

	return resp, nil
}

func buildTopicListLocalCacheKey(page, pageSize int64, sort string) string {
	return "topic:list:page:" + uint64ToString(uint64(page)) + ":size:" + uint64ToString(uint64(pageSize)) + ":sort:" + strings.ToLower(strings.TrimSpace(sort))
}
