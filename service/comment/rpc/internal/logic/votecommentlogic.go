package logic

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// Lua 脚本实现原子更新
const voteScript = `
local voteKey = KEYS[1]
local infoKey = KEYS[2]
local newVote = ARGV[1]
local expire = ARGV[2]

-- 获取旧的投票状态
local oldVote = redis.call('get', voteKey)
if not oldVote then
    oldVote = '0'
end

-- 如果状态没有变化，直接返回当前计数
if oldVote == newVote then
    local likeCount = redis.call('hget', infoKey, 'like_count') or '0'
    local dislikeCount = redis.call('hget', infoKey, 'dislike_count') or '0'
    return {likeCount, dislikeCount, newVote}
end

-- 计算计数变化
local likeDelta = 0
local dislikeDelta = 0

-- 减去旧状态的计数
if oldVote == '1' then
    likeDelta = -1
elseif oldVote == '2' then
    dislikeDelta = -1
end

-- 加上新状态的计数
if newVote == '1' then
    likeDelta = likeDelta + 1
elseif newVote == '2' then
    dislikeDelta = dislikeDelta + 1
end

-- 应用修改
if newVote == '0' then
    redis.call('del', voteKey)
else
    redis.call('set', voteKey, newVote)
    redis.call('expire', voteKey, expire)
end

if likeDelta ~= 0 then
    redis.call('hincrby', infoKey, 'like_count', likeDelta)
end
if dislikeDelta ~= 0 then
    redis.call('hincrby', infoKey, 'dislike_count', dislikeDelta)
end

-- 返回最新的计数
local likeCount = redis.call('hget', infoKey, 'like_count') or '0'
local dislikeCount = redis.call('hget', infoKey, 'dislike_count') or '0'

return {likeCount, dislikeCount, newVote}
`

type VoteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVoteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VoteCommentLogic {
	return &VoteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VoteCommentLogic) VoteComment(in *pb.VoteCommentReq) (*pb.VoteCommentResp, error) {
	if in.CommentId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if in.UserId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if in.VoteType < 0 || in.VoteType > 2 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	// 步骤一：检查评论是否存在
	var comment model.Comment
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND deleted_at IS NULL", in.CommentId).First(&comment).Error
	if err != nil {
		logx.Errorf("查询评论失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrCommentNotExist)
	}

	// 步骤二：防刷限流（用户当日投票次数）
	date := time.Now().Format("20060102")
	userVoteKey := fmt.Sprintf("%s%d:%s", redis.UserVoteCountKey, in.UserId, date)
	voteCount, err := l.svcCtx.Redis.Incr(l.ctx, userVoteKey)
	if err != nil {
		logx.Errorf("用户投票计数失败: %v", err)
		// 不影响主流程
	} else {
		// 设置过期时间
		if voteCount == 1 {
			l.svcCtx.Redis.Expire(l.ctx, userVoteKey, time.Hour*24)
		}
		// 每日最大 200 次
		if voteCount > 200 {
			return nil, errutil.RpcBusinessError(errutil.ErrTooManyRequests)
		}
	}

	// 步骤三：确保评论信息在 Redis 中存在
	commentInfoKey := fmt.Sprintf("%s%d", redis.CommentInfoKey, in.CommentId)
	exists, err := l.svcCtx.Redis.Exists(l.ctx, commentInfoKey)
	if err == nil && exists == 0 {
		// 缓存不存在，从数据库重建
		l.rebuildCommentInfo(comment)
	}

	// 步骤四：执行 Lua 脚本原子更新
	voteKey := fmt.Sprintf("%s%d:user:%d", redis.CommentVoteKey, in.CommentId, in.UserId)
	newVoteStr := fmt.Sprintf("%d", in.VoteType)
	expire := redis.VoteExpire

	// 执行 Lua 脚本
	result, err := l.svcCtx.Redis.Eval(l.ctx, voteScript, []string{voteKey, commentInfoKey}, newVoteStr, expire)
	if err != nil {
		logx.Errorf("执行投票Lua脚本失败: %v", err)
		// 降级到数据库处理
		return l.voteCommentDB(in)
	}

	// 解析 Lua 脚本返回结果
	results, ok := result.([]interface{})
	if !ok || len(results) < 3 {
		logx.Errorf("解析Lua脚本返回结果失败: %v", result)
		return l.voteCommentDB(in)
	}

	likeCount := parseInt32(results[0])
	dislikeCount := parseInt32(results[1])
	finalVoteType := parseInt32(results[2])

	// 步骤五：发送 Kafka 消息异步落库
	voteEvent := mq.VoteEvent{
		CommentID: in.CommentId,
		UserID:    in.UserId,
		VoteType:  in.VoteType,
		Timestamp: time.Now().Unix(),
	}

	go func() {
		if err := l.svcCtx.KafkaProducer.SendMessage(l.svcCtx.Config.KqProducerConf.TopicCommentVote, voteEvent); err != nil {
			logx.Errorf("发送投票事件消息失败: %v", err)
		}
	}()

	// 构建响应
	isLiked := finalVoteType == 1
	isDisliked := finalVoteType == 2

	// Vote only changes like_count/dislike_count in the comment info hash.
	// For "hot" sorted list, update the ZSet score for this comment instead of
	// invalidating the entire list. For "time" sorted list, the score (created_at) is unchanged.
	l.updateCommentScoreInLists(comment.PostID, comment.ID, likeCount, int32(comment.ReplyCount), comment.CreatedAt)
	if comment.ParentID > 0 {
		invalidateReplyCache(l.ctx, l.svcCtx, comment.ParentID)
	}

	return &pb.VoteCommentResp{
		Success:      true,
		LikeCount:    likeCount,
		DislikeCount: dislikeCount,
		IsLiked:      isLiked,
		IsDisliked:   isDisliked,
	}, nil
}

// 降级到数据库处理
func (l *VoteCommentLogic) voteCommentDB(in *pb.VoteCommentReq) (*pb.VoteCommentResp, error) {
	var comment model.Comment
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND deleted_at IS NULL", in.CommentId).First(&comment).Error
	if err != nil {
		logx.Errorf("查询评论失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrCommentNotExist)
	}

	var existingVote model.CommentVote
	existingVoteErr := l.svcCtx.DB.WithContext(l.ctx).Where("comment_id = ? AND user_id = ?", in.CommentId, in.UserId).First(&existingVote).Error

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if existingVoteErr == nil {
			oldVoteType := existingVote.VoteType
			if oldVoteType == 1 {
				comment.LikeCount--
			} else if oldVoteType == 2 {
				comment.DislikeCount--
			}

			if in.VoteType == 0 {
				if err := tx.WithContext(l.ctx).Delete(&existingVote).Error; err != nil {
					return err
				}
			} else {
				existingVote.VoteType = uint8(in.VoteType)
				if err := tx.WithContext(l.ctx).Save(&existingVote).Error; err != nil {
					return err
				}

				if in.VoteType == 1 {
					comment.LikeCount++
				} else if in.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		} else if existingVoteErr != gorm.ErrRecordNotFound {
			// 有其他错误
			if in.VoteType != 0 {
				newVote := model.CommentVote{
					CommentID: in.CommentId,
					UserID:    in.UserId,
					VoteType:  uint8(in.VoteType),
				}
				if err := tx.WithContext(l.ctx).Create(&newVote).Error; err != nil {
					return err
				}

				if in.VoteType == 1 {
					comment.LikeCount++
				} else if in.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		} else {
			return existingVoteErr
		}

		if err := tx.WithContext(l.ctx).Save(&comment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("评论投票失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var finalVote model.CommentVote
	isLiked := false
	isDisliked := false
	err = l.svcCtx.DB.WithContext(l.ctx).Where("comment_id = ? AND user_id = ?", in.CommentId, in.UserId).First(&finalVote).Error
	if err == nil {
		isLiked = finalVote.VoteType == 1
		isDisliked = finalVote.VoteType == 2
	} else if err != gorm.ErrRecordNotFound {
		logx.Errorf("查询最终投票状态失败: %v", err)
	}

	return &pb.VoteCommentResp{
		Success:      true,
		LikeCount:    comment.LikeCount,
		DislikeCount: comment.DislikeCount,
		IsLiked:      isLiked,
		IsDisliked:   isDisliked,
	}, nil
}

// 重建评论信息缓存
func (l *VoteCommentLogic) rebuildCommentInfo(comment model.Comment) {
	ctx := context.Background()
	commentInfoKey := fmt.Sprintf("%s%d", redis.CommentInfoKey, comment.ID)

	info := map[string]interface{}{
		"id":            comment.ID,
		"post_id":       comment.PostID,
		"user_id":       comment.UserID,
		"parent_id":     comment.ParentID,
		"level":         comment.Level,
		"reply_count":   comment.ReplyCount,
		"like_count":    comment.LikeCount,
		"dislike_count": comment.DislikeCount,
		"created_at":    comment.CreatedAt.Unix(),
	}
	if comment.ReplyToUserID != nil {
		info["reply_to_user_id"] = *comment.ReplyToUserID
	}

	l.svcCtx.Redis.HMSet(ctx, commentInfoKey, info)
}

// updateCommentScoreInLists updates the ZSet score for a single comment in the "hot" sorted list.
// The "time" sorted list uses created_at as the score, which doesn't change on vote.
func (l *VoteCommentLogic) updateCommentScoreInLists(postID, commentID uint64, likeCount, replyCount int32, createdAt time.Time) {
	ctx := context.Background()

	// Hot score formula: likeCount + replyCount*3 - ageInHours
	ageInHours := math.Max(0, time.Since(createdAt).Hours())
	hotScore := float64(likeCount) + float64(replyCount)*3 - ageInHours

	hotListKey := buildCommentListKey(postID, "hot")
	commentIDStr := strconv.FormatUint(commentID, 10)

	// Only update if the key exists (not a full rebuild)
	exists, err := l.svcCtx.Redis.Exists(ctx, hotListKey)
	if err == nil && exists > 0 {
		if zaddErr := l.svcCtx.Redis.ZAdd(ctx, hotListKey, hotScore, commentIDStr); zaddErr != nil {
			logx.Errorf("update hot score failed: commentId=%d err=%v", commentID, zaddErr)
		}
	}
}

func parseInt32(v interface{}) int32 {
	switch val := v.(type) {
	case string:
		res, _ := strconv.ParseInt(val, 10, 32)
		return int32(res)
	case []byte:
		res, _ := strconv.ParseInt(string(val), 10, 32)
		return int32(res)
	case int:
		return int32(val)
	case int64:
		return int32(val)
	case float64:
		return int32(val)
	default:
		return 0
	}
}
