package logic

import (
	"context"
	"errors"

	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

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
	var comment model.Comment
	err := l.svcCtx.DB.Where("id = ? AND deleted_at IS NULL", in.CommentId).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}

	var existingVote model.CommentVote
	err = l.svcCtx.DB.Where("comment_id = ? AND user_id = ?", in.CommentId, in.UserId).First(&existingVote).Error

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err == nil {
			oldVoteType := existingVote.VoteType
			if oldVoteType == 1 {
				comment.LikeCount--
			} else if oldVoteType == 2 {
				comment.DislikeCount--
			}

			if in.VoteType == 0 {
				if err := tx.Delete(&existingVote).Error; err != nil {
					return err
				}
			} else {
				existingVote.VoteType = uint8(in.VoteType)
				if err := tx.Save(&existingVote).Error; err != nil {
					return err
				}

				if in.VoteType == 1 {
					comment.LikeCount++
				} else if in.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		} else {
			if in.VoteType != 0 {
				newVote := model.CommentVote{
					CommentID: in.CommentId,
					UserID:    in.UserId,
					VoteType:  uint8(in.VoteType),
				}
				if err := tx.Create(&newVote).Error; err != nil {
					return err
				}

				if in.VoteType == 1 {
					comment.LikeCount++
				} else if in.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		}

		if err := tx.Save(&comment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var finalVote model.CommentVote
	isLiked := false
	isDisliked := false
	err = l.svcCtx.DB.Where("comment_id = ? AND user_id = ?", in.CommentId, in.UserId).First(&finalVote).Error
	if err == nil {
		isLiked = finalVote.VoteType == 1
		isDisliked = finalVote.VoteType == 2
	}

	return &pb.VoteCommentResp{
		Success:      true,
		LikeCount:    comment.LikeCount,
		DislikeCount: comment.DislikeCount,
		IsLiked:      isLiked,
		IsDisliked:   isDisliked,
	}, nil
}
