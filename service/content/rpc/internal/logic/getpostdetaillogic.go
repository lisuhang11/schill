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

type GetPostDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostDetailLogic {
	return &GetPostDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPostDetailLogic) GetPostDetail(in *pb.GetPostDetailReq) (*pb.GetPostDetailResp, error) {
	var post model.Post
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.PostId).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var postContents []model.PostContent
	err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id = ?", in.PostId).Order("sort ASC").Find(&postContents).Error
	if err != nil {
		logx.Errorf("查询帖子内容失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var title string
	var cover string
	var contents []*pb.PostContentItem
	for _, pc := range postContents {
		if pc.Type == 1 {
			title = pc.Content
		} else if pc.Type == 3 && cover == "" {
			cover = pc.Content
		} else {
			contents = append(contents, &pb.PostContentItem{
				Type:    pc.Type,
				Content: pc.Content,
				Sort:    pc.Sort,
			})
		}
	}

	var postTopics []model.PostTopic
	err = l.svcCtx.DB.WithContext(l.ctx).Where("post_id = ?", in.PostId).Find(&postTopics).Error
	if err != nil {
		logx.Errorf("查询帖子话题失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var topics []*pb.PostTopic
	for _, pt := range postTopics {
		var topic model.Topic
		err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", pt.TopicID).First(&topic).Error
		if err != nil {
			logx.Errorf("查询话题信息失败: %v", err)
			continue
		}
		topics = append(topics, &pb.PostTopic{
			PostId:    pt.PostID,
			TopicId:   pt.TopicID,
			TopicName: topic.Name,
		})
	}

	return &pb.GetPostDetailResp{
		Post: &pb.PostInfo{
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
		},
		Contents: contents,
		Topics:   topics,
	}, nil
}
