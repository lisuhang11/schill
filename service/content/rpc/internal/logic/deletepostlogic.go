package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeletePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePostLogic {
	return &DeletePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePostLogic) DeletePost(in *pb.DeletePostReq) (*pb.DeletePostResp, error) {
	var post model.Post
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.PostId).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if post.UserID != in.UserId {
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var postTopics []model.PostTopic
		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Find(&postTopics).Error; err != nil {
			return err
		}

		for _, pt := range postTopics {
			if err := tx.WithContext(l.ctx).Model(&model.Topic{}).Where("id = ?", pt.TopicID).Update("quote_num", gorm.Expr("quote_num - ?", 1)).Error; err != nil {
				return err
			}
		}

		if err := tx.WithContext(l.ctx).Delete(&post).Error; err != nil {
			return err
		}

		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostContent{}).Error; err != nil {
			return err
		}

		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostTopic{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("删除帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.DeletePostResp{
		Success: true,
	}, nil
}
