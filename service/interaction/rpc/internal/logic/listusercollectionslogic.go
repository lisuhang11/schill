package logic

import (
	"context"

	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserCollectionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserCollectionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserCollectionsLogic {
	return &ListUserCollectionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserCollectionsLogic) ListUserCollections(in *pb.ListUserCollectionsReq) (*pb.ListUserCollectionsResp, error) {
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

	offset := int((page - 1) * pageSize)

	total, err := l.svcCtx.PostCollectionDAO.CountByUser(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("count user collections failed: userId=%d err=%v", in.UserId, err)
		return nil, err
	}

	collections, err := l.svcCtx.PostCollectionDAO.ListByUser(l.ctx, in.UserId, offset, int(pageSize))
	if err != nil {
		logx.Errorf("list user collections failed: userId=%d err=%v", in.UserId, err)
		return nil, err
	}

	postIds := make([]uint64, 0, len(collections))
	for _, c := range collections {
		postIds = append(postIds, c.PostID)
	}

	return &pb.ListUserCollectionsResp{
		PostIds: postIds,
		Total:   total,
	}, nil
}
