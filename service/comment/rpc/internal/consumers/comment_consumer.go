package consumers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/config"
	"SChill/service/comment/rpc/internal/model"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const maxRetryCount = 3

type CommentConsumer struct {
	consumer sarama.ConsumerGroup
	ctx      context.Context
	cancel   context.CancelFunc
	db       *gorm.DB
	redis    *redis.Client
	config   config.Config
	producer *mq.Producer
}

func NewCommentConsumer(cfg config.Config, db *gorm.DB, redisClient *redis.Client, producer *mq.Producer) (*CommentConsumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(cfg.KqConsumerConf.Brokers, cfg.KqConsumerConf.Group, saramaConfig)
	if err != nil {
		cancel()
		return nil, err
	}

	return &CommentConsumer{
		consumer: consumer,
		ctx:      ctx,
		cancel:   cancel,
		db:       db,
		redis:    redisClient,
		config:   cfg,
		producer: producer,
	}, nil
}

func (c *CommentConsumer) Start() {
	go func() {
		for {
			topics := []string{
				c.config.KqConsumerConf.TopicCommentCreate,
				c.config.KqConsumerConf.TopicCommentDeleted,
				c.config.KqConsumerConf.TopicCommentVote,
			}
			if err := c.consumer.Consume(c.ctx, topics, c); err != nil {
				logx.Errorf("comment consumer error: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	logx.Info("comment consumer started")
}

func (c *CommentConsumer) Stop() {
	c.cancel()
	if err := c.consumer.Close(); err != nil {
		logx.Errorf("close comment consumer failed: %v", err)
	}
}

func (c *CommentConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *CommentConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *CommentConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		lag := time.Since(message.Timestamp)
		logx.Infof("comment consumer received topic=%s partition=%d offset=%d lag_ms=%d", message.Topic, message.Partition, message.Offset, lag.Milliseconds())

		var (
			err      error
			envelope *mq.EventEnvelope
		)
		switch message.Topic {
		case c.config.KqConsumerConf.TopicCommentCreate:
			var event mq.CommentCreateEvent
			if envelope, err = mq.DecodeEnvelopePayload(message.Value, &event); err != nil {
				logx.Errorf("decode comment create event failed: %v", err)
			} else if c.skipIfProcessed(envelope) {
				session.MarkMessage(message, "")
				continue
			} else {
				err = c.handleCommentCreateEvent(event)
			}
		case c.config.KqConsumerConf.TopicCommentDeleted:
			var msg mq.CommentDeletedMessage
			if envelope, err = mq.DecodeEnvelopePayload(message.Value, &msg); err != nil {
				logx.Errorf("decode comment deleted event failed: %v", err)
			} else if c.skipIfProcessed(envelope) {
				session.MarkMessage(message, "")
				continue
			} else {
				err = c.handleCommentDeletedEvent(msg)
			}
		case c.config.KqConsumerConf.TopicCommentVote:
			var event mq.VoteEvent
			if envelope, err = mq.DecodeEnvelopePayload(message.Value, &event); err != nil {
				logx.Errorf("decode comment vote event failed: %v", err)
			} else if c.skipIfProcessed(envelope) {
				session.MarkMessage(message, "")
				continue
			} else {
				err = c.handleVoteEvent(event)
			}
		}

		if err != nil {
			logx.Errorf("comment consumer handle failed: topic=%s offset=%d err=%v", message.Topic, message.Offset, err)
			c.retryOrDLQ(message)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func (c *CommentConsumer) skipIfProcessed(envelope *mq.EventEnvelope) bool {
	key := mq.BuildIdempotencyKey(c.config.KqConsumerConf.Group, envelope)
	if key == "" {
		return false
	}
	ok, err := c.redis.SetNX(context.Background(), key, "1", mq.DefaultEventTTL)
	if err != nil {
		logx.Errorf("comment consumer idempotency check failed: %v", err)
		return false
	}
	return !ok
}

func (c *CommentConsumer) retryOrDLQ(message *sarama.ConsumerMessage) {
	retryCount := 0
	for _, header := range message.Headers {
		if string(header.Key) == "x-retry-count" {
			retryCount, _ = strconv.Atoi(string(header.Value))
			break
		}
	}

	if retryCount < maxRetryCount {
		headers := make([]sarama.RecordHeader, 0, len(message.Headers)+1)
		for _, header := range message.Headers {
			if header == nil {
				continue
			}
			headers = append(headers, sarama.RecordHeader{Key: header.Key, Value: header.Value})
		}
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte("x-retry-count"),
			Value: []byte(strconv.Itoa(retryCount + 1)),
		})
		_ = c.producer.SendRawMessage(&sarama.ProducerMessage{
			Topic:   message.Topic,
			Key:     sarama.ByteEncoder(message.Key),
			Value:   sarama.ByteEncoder(message.Value),
			Headers: headers,
		})
		return
	}

	if c.config.KqConsumerConf.TopicCommentDLQ == "" {
		return
	}
	_ = c.producer.SendRawMessage(&sarama.ProducerMessage{
		Topic: c.config.KqConsumerConf.TopicCommentDLQ,
		Key:   sarama.ByteEncoder(message.Key),
		Value: sarama.ByteEncoder(message.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("x-source-topic"), Value: []byte(message.Topic)},
			{Key: []byte("x-original-offset"), Value: []byte(strconv.FormatInt(message.Offset, 10))},
		},
	})
}

func (c *CommentConsumer) handleCommentCreateEvent(event mq.CommentCreateEvent) error {
	ctx := context.Background()

	commentInfoKey := fmt.Sprintf("%s%d", redis.CommentInfoKey, event.CommentID)
	commentInfo := map[string]interface{}{
		"id":               event.CommentID,
		"post_id":          event.PostID,
		"user_id":          event.UserID,
		"parent_id":        event.ParentID,
		"reply_to_user_id": event.ReplyToUserID,
		"level":            event.Level,
		"reply_count":      0,
		"like_count":       0,
		"dislike_count":    0,
		"created_at":       event.CreatedAt,
	}
	if err := c.redis.HMSet(ctx, commentInfoKey, commentInfo); err != nil {
		return err
	}
	_, _ = c.redis.Expire(ctx, commentInfoKey, time.Duration(redis.CommentExpire)*time.Second)

	commentContentKey := fmt.Sprintf("%s%d", redis.CommentContentKey, event.CommentID)
	if err := c.redis.Set(ctx, commentContentKey, event.Content, time.Duration(redis.CommentExpire)*time.Second); err != nil {
		return err
	}

	if event.ParentID == 0 {
		postCommentsKey := fmt.Sprintf("%s%d:list", redis.PostCommentsKey, event.PostID)
		if err := c.redis.ZAdd(ctx, postCommentsKey, redis.Z{
			Score:  float64(event.CreatedAt),
			Member: event.CommentID,
		}); err != nil {
			return err
		}
		_, _ = c.redis.Expire(ctx, postCommentsKey, time.Duration(redis.CommentExpire)*time.Second)
		_, _ = c.redis.IncrBy(ctx, fmt.Sprintf("%s%d", redis.PostCommentCountKey, event.PostID), 1)
	}

	if event.ParentID > 0 {
		replyListKey := fmt.Sprintf("%s%d:list", redis.CommentRepliesKey, event.ParentID)
		if err := c.redis.ZAdd(ctx, replyListKey, redis.Z{
			Score:  float64(event.CreatedAt),
			Member: event.CommentID,
		}); err != nil {
			return err
		}
		_, _ = c.redis.Expire(ctx, replyListKey, time.Duration(redis.CommentExpire)*time.Second)

		parentCommentInfoKey := fmt.Sprintf("%s%d", redis.CommentInfoKey, event.ParentID)
		if err := c.redis.HIncrBy(ctx, parentCommentInfoKey, "reply_count", 1); err != nil {
			return err
		}
		_, _ = c.redis.IncrBy(ctx, fmt.Sprintf("%s%d", redis.CommentReplyCountKey, event.ParentID), 1)
	}

	return nil
}

func (c *CommentConsumer) handleCommentDeletedEvent(msg mq.CommentDeletedMessage) error {
	ctx := context.Background()

	var comment model.Comment
	if err := c.db.WithContext(ctx).Where("id = ?", msg.CommentID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	if err := c.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		comment.DeletedAt = &now
		comment.Status = 3

		if err := tx.WithContext(ctx).Save(&comment).Error; err != nil {
			return err
		}
		if comment.ParentID > 0 {
			if err := tx.WithContext(ctx).
				Model(&model.Comment{}).
				Where("id = ?", comment.ParentID).
				Update("reply_count", gorm.Expr("reply_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	commentInfoKey := fmt.Sprintf("%s%d", redis.CommentInfoKey, comment.ID)
	commentContentKey := fmt.Sprintf("%s%d", redis.CommentContentKey, comment.ID)
	_ = c.redis.Del(ctx, commentInfoKey, commentContentKey)

	_ = c.redis.ZRem(ctx, fmt.Sprintf("%s%d:list", redis.PostCommentsKey, comment.PostID), comment.ID)
	_ = c.redis.ZRem(ctx, fmt.Sprintf("%s%d:hot", redis.PostCommentsKey, comment.PostID), comment.ID)
	if comment.ParentID == 0 {
		_, _ = c.redis.IncrBy(ctx, fmt.Sprintf("%s%d", redis.PostCommentCountKey, comment.PostID), -1)
	}

	if comment.ParentID > 0 {
		_ = c.redis.ZRem(ctx, fmt.Sprintf("%s%d:list", redis.CommentRepliesKey, comment.ParentID), comment.ID)
		_ = c.redis.HIncrBy(ctx, fmt.Sprintf("%s%d", redis.CommentInfoKey, comment.ParentID), "reply_count", -1)
		_, _ = c.redis.IncrBy(ctx, fmt.Sprintf("%s%d", redis.CommentReplyCountKey, comment.ParentID), -1)
	}

	return nil
}

func (c *CommentConsumer) handleVoteEvent(event mq.VoteEvent) error {
	ctx := context.Background()

	var comment model.Comment
	if err := c.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", event.CommentID).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	var existingVote model.CommentVote
	existingVoteErr := c.db.WithContext(ctx).Where("comment_id = ? AND user_id = ?", event.CommentID, event.UserID).First(&existingVote).Error

	if err := c.db.Transaction(func(tx *gorm.DB) error {
		switch {
		case existingVoteErr == nil:
			if existingVote.VoteType == 1 {
				comment.LikeCount--
			} else if existingVote.VoteType == 2 {
				comment.DislikeCount--
			}

			if event.VoteType == 0 {
				if err := tx.WithContext(ctx).Delete(&existingVote).Error; err != nil {
					return err
				}
			} else {
				existingVote.VoteType = uint8(event.VoteType)
				if err := tx.WithContext(ctx).Save(&existingVote).Error; err != nil {
					return err
				}
				if event.VoteType == 1 {
					comment.LikeCount++
				} else if event.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		case existingVoteErr == gorm.ErrRecordNotFound:
			if event.VoteType != 0 {
				newVote := model.CommentVote{
					CommentID: event.CommentID,
					UserID:    event.UserID,
					VoteType:  uint8(event.VoteType),
				}
				if err := tx.WithContext(ctx).Create(&newVote).Error; err != nil {
					return err
				}
				if event.VoteType == 1 {
					comment.LikeCount++
				} else if event.VoteType == 2 {
					comment.DislikeCount++
				}
			}
		default:
			return existingVoteErr
		}

		return tx.WithContext(ctx).Save(&comment).Error
	}); err != nil {
		return err
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		var stat model.CommentStat
		err := tx.WithContext(ctx).Where("comment_id = ?", event.CommentID).First(&stat).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if err == gorm.ErrRecordNotFound {
			return tx.WithContext(ctx).Create(&model.CommentStat{
				CommentID:    event.CommentID,
				ReplyCount:   0,
				LikeCount:    uint32(comment.LikeCount),
				DislikeCount: uint32(comment.DislikeCount),
			}).Error
		}

		stat.LikeCount = uint32(comment.LikeCount)
		stat.DislikeCount = uint32(comment.DislikeCount)
		return tx.WithContext(ctx).Save(&stat).Error
	})
}
