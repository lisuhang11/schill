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

	postIDs := make([]uint64, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}

	var postContents []model.PostContent
	err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id IN ?", postIDs).Order("post_id, sort").Find(&postContents).Error
	if err != nil {
		logx.Errorf("查询帖子内容失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	contentMap := make(map[uint64]*struct {
		title string
		cover string
	})
	for _, pc := range postContents {
		if _, ok := contentMap[pc.PostID]; !ok {
			contentMap[pc.PostID] = &struct {
				title string
				cover string
			}{}
		}
		if pc.Type == 1 {
			contentMap[pc.PostID].title = pc.Content
		} else if pc.Type == 3 {
			contentMap[pc.PostID].cover = pc.Content
		}
	}

	var postList []*pb.PostInfo
	for _, post := range posts {
		title := ""
		cover := ""
		if cm, ok := contentMap[post.ID]; ok {
			title = cm.title
			cover = cm.cover
		}
		postList = append(postList, &pb.PostInfo{
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
		})
	}

	return &pb.BatchGetPostResp{
		Posts: postList,
	}, nil
}
