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
	"gorm.io/gorm"
)

type UpdatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePostLogic) UpdatePost(in *pb.UpdatePostReq) (*pb.UpdatePostResp, error) {
	if in.UserId == 0 || in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrUnauthorized)
	}

	contents := normalizePostContents(in.Contents)
	if len(contents) == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = deriveTitleFromContentItems(contents)
	}
	if title == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrPostTitleEmpty)
	}

	cover := strings.TrimSpace(in.Cover)
	if cover == "" {
		cover = deriveCoverFromContentItems(contents)
	}

	post, err := loadOwnedPost(l.ctx, l.svcCtx.DBWrite, in.PostId, in.UserId)
	if err != nil {
		return nil, err
	}

	topicNames := normalizeTopicNames(in.Topics)
	postContents := make([]*model.PostContent, 0, len(contents))
	summaryItems := make([]contentSummaryItem, 0, len(contents))
	for idx, item := range contents {
		sort := item.Sort
		if sort <= 0 {
			sort = int32(idx + 1)
		}
		postContents = append(postContents, &model.PostContent{
			Content: item.Content,
			Type:    item.Type,
			Sort:    sort,
		})
		summaryItems = append(summaryItems, contentSummaryItem{
			Type:    item.Type,
			Content: item.Content,
		})
	}

	now := time.Now()
	post.Title = title
	post.Cover = cover
	post.Visibility = in.Visibility
	post.Tags = strings.TrimSpace(in.Tags)
	post.UpdatedAt = now

	if err := l.svcCtx.DBWrite.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(l.ctx).
			Model(&model.Post{}).
			Where("id = ? AND deleted_at IS NULL", in.PostId).
			Updates(map[string]interface{}{
				"title":      post.Title,
				"cover":      post.Cover,
				"visibility": post.Visibility,
				"tags":       post.Tags,
				"updated_at": post.UpdatedAt,
			}).Error; err != nil {
			return err
		}

		if err := replacePostContentsTx(l.ctx, tx, in.PostId, in.UserId, postContents); err != nil {
			return err
		}

		if err := replacePostTopicsTx(l.ctx, tx, in.PostId, topicNames); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logx.Errorf("update post failed: postId=%d err=%v", in.PostId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	invalidatePostCachesByModel(l.ctx, l.svcCtx, post)
	publishContentChangedEvent(l.ctx, l.svcCtx, post, "updated", buildContentSummaryFromItems(summaryItems), topicNames)

	return &pb.UpdatePostResp{Success: true}, nil
}
