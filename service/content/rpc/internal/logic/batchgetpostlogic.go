package logic

import (
	"context"
	"strings"

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
		return &pb.BatchGetPostResp{Posts: []*pb.PostInfo{}}, nil
	}

	postIDs := uniqueUint64s(in.PostIds)

	var posts []model.Post
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("id IN ? AND deleted_at IS NULL", postIDs).
		Find(&posts).Error; err != nil {
		logx.Errorf("batch get posts failed: postIds=%v err=%v", postIDs, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	derivedMeta := l.batchLoadDerivedPostMeta(posts)
	postMap := make(map[uint64]*pb.PostInfo, len(posts))
	for _, post := range posts {
		title := post.Title
		cover := post.Cover
		if meta, ok := derivedMeta[post.ID]; ok {
			if title == "" {
				title = meta.Title
			}
			if cover == "" {
				cover = meta.Cover
			}
		}

		postMap[post.ID] = &pb.PostInfo{
			Id:              post.ID,
			UserId:          post.UserID,
			Title:           title,
			Cover:           cover,
			CommentCount:    post.CommentCount,
			CollectionCount: post.CollectionCount,
			UpvoteCount:     post.UpvoteCount,
			ShareCount:      post.ShareCount,
			Visibility:      post.Visibility,
			IsTop:           post.IsTop,
			IsEssence:       post.IsEssence,
			IsLock:          post.IsLock,
			LatestRepliedAt: post.LatestRepliedAt,
			Tags:            post.Tags,
			CreatedAt:       post.CreatedAt.Unix(),
			UpdatedAt:       post.UpdatedAt.Unix(),
		}
	}

	resp := &pb.BatchGetPostResp{
		Posts: make([]*pb.PostInfo, 0, len(postIDs)),
	}
	for _, postID := range postIDs {
		if post, ok := postMap[postID]; ok {
			resp.Posts = append(resp.Posts, post)
		}
	}

	return resp, nil
}

func (l *BatchGetPostLogic) batchLoadDerivedPostMeta(posts []model.Post) map[uint64]postDerivedMeta {
	postIDs := make([]uint64, 0, len(posts))
	for _, post := range posts {
		if post.Title == "" || post.Cover == "" {
			postIDs = append(postIDs, post.ID)
		}
	}
	if len(postIDs) == 0 {
		return nil
	}

	var contents []model.PostContent
	if err := l.svcCtx.DBRead.WithContext(l.ctx).
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Order("post_id ASC, sort ASC, id ASC").
		Find(&contents).Error; err != nil {
		logx.Errorf("load post content for batch derived meta failed: %v", err)
		return nil
	}

	result := make(map[uint64]postDerivedMeta, len(postIDs))
	for _, content := range contents {
		meta := result[content.PostID]
		if meta.Title == "" {
			meta.Title = deriveTitleFromContentString(content.Content)
		}
		if meta.Cover == "" && content.Type == 3 {
			meta.Cover = strings.TrimSpace(content.Content)
		}
		result[content.PostID] = meta
	}

	return result
}

func uniqueUint64s(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
