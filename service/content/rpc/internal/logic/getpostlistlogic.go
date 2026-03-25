package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostListLogic {
	return &GetPostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPostListLogic) GetPostList(in *pb.GetPostListReq) (*pb.GetPostListResp, error) {
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Post{})

	if in.UserId > 0 {
		query = query.Where("user_id = ?", in.UserId)
	}

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		logx.Errorf("统计帖子数量失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if total == 0 {
		return &pb.GetPostListResp{
			Total: 0,
			List:  []*pb.PostInfo{},
		}, nil
	}

	offset := (in.Page - 1) * in.PageSize
	var posts []model.Post
	err = query.Limit(int(in.PageSize)).Offset(int(offset)).Order("created_at DESC").Find(&posts).Error
	if err != nil {
		logx.Errorf("查询帖子列表失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	postIds := make([]uint64, 0, len(posts))
	for _, post := range posts {
		postIds = append(postIds, post.ID)
	}

	var postContents []model.PostContent
	err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id IN ? AND type = 1", postIds).Find(&postContents).Error
	if err != nil {
		logx.Errorf("查询帖子标题失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	titleMap := make(map[uint64]string)
	for _, pc := range postContents {
		titleMap[pc.PostID] = pc.Content
	}

	var coverContents []model.PostContent
	err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id IN ? AND type = 3", postIds).Find(&coverContents).Error
	if err != nil {
		logx.Errorf("查询帖子封面失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	coverMap := make(map[uint64]string)
	for _, pc := range coverContents {
		if _, ok := coverMap[pc.PostID]; !ok {
			coverMap[pc.PostID] = pc.Content
		}
	}

	var list []*pb.PostInfo
	for _, post := range posts {
		list = append(list, &pb.PostInfo{
			Id:              post.ID,
			UserId:          post.UserID,
			Title:           titleMap[post.ID],
			Cover:           coverMap[post.ID],
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

	return &pb.GetPostListResp{
		Total: total,
		List:  list,
	}, nil
}
