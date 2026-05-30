package logic

import (
	"context"
	"strings"

	errutil "SChill/common/error"
	contentpb "SChill/service/content/rpc/pb"
	"SChill/service/feed/rpc/internal/svc"
	"SChill/service/feed/rpc/pb"
	interactionpb "SChill/service/interaction/rpc/pb"
	relationpb "SChill/service/relation/rpc/pb"
	userpb "SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetFeedListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFeedListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedListLogic {
	return &GetFeedListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFeedListLogic) GetFeedList(req *pb.GetFeedListReq) (*pb.GetFeedListResp, error) {
	if req.Page <= 0 || req.PageSize <= 0 || req.PageSize > 100 {
		return &pb.GetFeedListResp{
			Code: int32(errutil.ErrInvalidParams),
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	feedType := normalizeFeedType(req.FeedType)
	userID := req.CurrentUserId
	if feedType == "following" && userID == 0 {
		return &pb.GetFeedListResp{
			Code: int32(errutil.ErrUnauthorized),
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	rpcCtx := metadata.AppendToOutgoingContext(l.ctx, "x-feed-type", feedType)
	postResp, err := l.svcCtx.ContentRpc.GetPostList(rpcCtx, &contentpb.GetPostListReq{
		Page:          req.Page,
		PageSize:      req.PageSize,
		CurrentUserId: userID,
	})
	if err != nil {
		code, msg := errutil.ParseRpcError(err)
		return &pb.GetFeedListResp{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	postIDs := make([]uint64, 0, len(postResp.List))
	userIDs := make([]uint64, 0, len(postResp.List))
	seenUsers := make(map[uint64]struct{}, len(postResp.List))
	for _, item := range postResp.List {
		postIDs = append(postIDs, item.Id)
		if _, ok := seenUsers[item.UserId]; !ok {
			seenUsers[item.UserId] = struct{}{}
			userIDs = append(userIDs, item.UserId)
		}
	}

	summaryMap := l.loadSummaries(postIDs)
	authorMap := l.loadAuthors(userIDs)
	starMap, collectionMap := l.loadViewerInteraction(userID, postIDs)
	followMap := l.loadFollowStatus(userID, userIDs)

	list := make([]*pb.FeedItem, 0, len(postResp.List))
	for _, item := range postResp.List {
		summary := summaryMap[item.Id]
		if summary == "" {
			summary = item.Title
		}
		author := authorMap[item.UserId]
		list = append(list, &pb.FeedItem{
			PostId:          item.Id,
			AuthorId:        item.UserId,
			Title:           item.Title,
			Summary:         summary,
			Cover:           item.Cover,
			Tags:            splitTags(item.Tags),
			Visibility:      item.Visibility,
			IsTop:           item.IsTop == 1,
			IsEssence:       item.IsEssence == 1,
			IsLock:          item.IsLock == 1,
			CommentCount:    item.CommentCount,
			UpvoteCount:     item.UpvoteCount,
			CollectionCount: item.CollectionCount,
			ShareCount:      item.ShareCount,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
			LatestRepliedAt: item.LatestRepliedAt,
			FeedType:        feedType,
			Source:          feedType,
			Author: &pb.FeedAuthor{
				Id:       author.Id,
				Username: author.Username,
				Nickname: author.Nickname,
				Avatar:   author.Avatar,
			},
			ViewerState: &pb.FeedViewerState{
				IsStarred:         starMap[item.Id],
				IsCollected:       collectionMap[item.Id],
				IsFollowingAuthor: followMap[item.UserId],
			},
		})
	}

	return &pb.GetFeedListResp{
		Code:  int32(errutil.Success),
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: postResp.Total,
		List:  list,
	}, nil
}

func normalizeFeedType(feedType string) string {
	switch strings.ToLower(strings.TrimSpace(feedType)) {
	case "", "latest":
		return "latest"
	case "following":
		return "following"
	default:
		return "latest"
	}
}

func (l *GetFeedListLogic) loadSummaries(postIDs []uint64) map[uint64]string {
	result := make(map[uint64]string, len(postIDs))
	if len(postIDs) == 0 {
		return result
	}

	resp, err := l.svcCtx.ContentRpc.BatchGetPostSummary(l.ctx, &contentpb.BatchGetPostSummaryReq{
		PostIds: postIDs,
	})
	if err != nil {
		logx.Errorf("load post summaries failed: %v", err)
		return result
	}

	for _, item := range resp.Posts {
		result[item.Id] = item.Summary
	}
	return result
}

func (l *GetFeedListLogic) loadAuthors(userIDs []uint64) map[uint64]*userpb.UserBasicInfo {
	result := make(map[uint64]*userpb.UserBasicInfo, len(userIDs))
	if len(userIDs) == 0 {
		return result
	}

	resp, err := l.svcCtx.UserRpc.BatchGetUserBasicInfo(l.ctx, &userpb.BatchGetUserBasicInfoReq{
		UserIds: userIDs,
	})
	if err != nil {
		logx.Errorf("load authors failed: %v", err)
		for _, userID := range userIDs {
			result[userID] = fallbackAuthor(userID)
		}
		return result
	}

	for _, item := range resp.Users {
		result[item.Id] = item
	}
	for _, userID := range userIDs {
		if _, ok := result[userID]; !ok {
			result[userID] = fallbackAuthor(userID)
		}
	}
	return result
}

func (l *GetFeedListLogic) loadViewerInteraction(userID uint64, postIDs []uint64) (map[uint64]bool, map[uint64]bool) {
	starMap := make(map[uint64]bool, len(postIDs))
	collectionMap := make(map[uint64]bool, len(postIDs))
	if len(postIDs) == 0 || userID == 0 {
		return starMap, collectionMap
	}

	starResp, err := l.svcCtx.InteractionRpc.BatchCheckPostStar(l.ctx, &interactionpb.BatchCheckPostStarReq{
		UserId:  userID,
		PostIds: postIDs,
	})
	if err != nil {
		logx.Errorf("load star status failed: %v", err)
	} else {
		for postID, status := range starResp.StarStatus {
			starMap[postID] = status
		}
	}

	collectionResp, err := l.svcCtx.InteractionRpc.BatchCheckPostCollection(l.ctx, &interactionpb.BatchCheckPostCollectionReq{
		UserId:  userID,
		PostIds: postIDs,
	})
	if err != nil {
		logx.Errorf("load collection status failed: %v", err)
	} else {
		for postID, status := range collectionResp.CollectionStatus {
			collectionMap[postID] = status
		}
	}

	return starMap, collectionMap
}

func (l *GetFeedListLogic) loadFollowStatus(userID uint64, userIDs []uint64) map[uint64]bool {
	result := make(map[uint64]bool, len(userIDs))
	if len(userIDs) == 0 || userID == 0 {
		return result
	}

	resp, err := l.svcCtx.RelationRpc.BatchCheckFollowStatus(l.ctx, &relationpb.BatchCheckFollowStatusReq{
		UserId:        userID,
		TargetUserIds: userIDs,
	})
	if err != nil {
		logx.Errorf("load follow status failed: %v", err)
		return result
	}

	for _, item := range resp.Status {
		result[item.UserId] = item.IsFollow
	}
	return result
}

func splitTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func fallbackAuthor(userID uint64) *userpb.UserBasicInfo {
	return &userpb.UserBasicInfo{
		Id:       userID,
		Username: "user",
		Nickname: "",
		Avatar:   "",
	}
}
