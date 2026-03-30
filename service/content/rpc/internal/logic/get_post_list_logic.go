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

	offset := (page - 1) * pageSize

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

	var posts []model.Post
	err = query.Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&posts).Error
	if err != nil {
		logx.Errorf("查询帖子列表失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	postIDs := make([]uint64, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}

	var postContents []model.PostContent
	if len(postIDs) > 0 {
		err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id IN ?", postIDs).Order("post_id, sort").Find(&postContents).Error
		if err != nil {
			logx.Errorf("查询帖子内容失败: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
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

	return &pb.GetPostListResp{
		Total: total,
		List:  postList,
	}, nil
}
