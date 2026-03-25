package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetPostLogic {
	return &BatchGetPostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchGetPostLogic) BatchGetPost(in *pb.BatchGetPostReq) (*pb.BatchGetPostResp, error) {
	if len(in.PostIds) == 0 {
		return &pb.BatchGetPostResp{
			Posts: []*pb.PostInfo{},
		}, nil
	}

	var posts []model.Post
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id IN ?", in.PostIds).Find(&posts).Error
	if err != nil {
		logx.Errorf("批量查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var postList []*pb.PostInfo
	for _, post := range posts {
		postList = append(postList, &pb.PostInfo{
			Id:           post.ID,
			UserId:       post.UserID,
			Title:        post.Title,
			Cover:        post.Cover,
			Type:         post.Type,
			Status:       post.Status,
			ViewCount:    int64(post.ViewCount),
			LikeCount:    int64(post.LikeCount),
			CommentCount: int64(post.CommentCount),
			CreatedAt:    post.CreatedAt.Unix(),
			UpdatedAt:    post.UpdatedAt.Unix(),
		})
	}

	return &pb.BatchGetPostResp{
		Posts: postList,
	}, nil
}
