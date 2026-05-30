package logic

import (
	"context"
	"strconv"
	"sync"
	"time"

	"SChill/common/mq"
	"SChill/service/interaction/rpc/internal/model"
	"SChill/service/interaction/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

const batchSize = 100

type InteractionConsumer struct {
	svcCtx        *svc.ServiceContext
	ctx           context.Context
	cancel        context.CancelFunc
	consumerGroup sarama.ConsumerGroup
	wg            sync.WaitGroup
}

func NewInteractionConsumer(svcCtx *svc.ServiceContext) *InteractionConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &InteractionConsumer{
		svcCtx: svcCtx,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *InteractionConsumer) Start() error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Fetch.Default = 1048576
	config.Version = sarama.V2_8_0_0

	group, err := sarama.NewConsumerGroup(
		c.svcCtx.Config.KafkaConsumerConf.Brokers,
		c.svcCtx.Config.KafkaConsumerConf.Group,
		config,
	)
	if err != nil {
		return err
	}
	c.consumerGroup = group
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(c.ctx, c.svcCtx.Config.KafkaConsumerConf.Topics, c); err != nil && c.ctx.Err() == nil {
				logx.Errorf("interaction consumer error: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	logx.Infof("Interaction consumer started, topics: %v, group: %s",
		c.svcCtx.Config.KafkaConsumerConf.Topics,
		c.svcCtx.Config.KafkaConsumerConf.Group)
	return nil
}

func (c *InteractionConsumer) Stop() {
	c.cancel()
	if c.consumerGroup != nil {
		_ = c.consumerGroup.Close()
	}
	c.wg.Wait()
}

func (c *InteractionConsumer) sendToDLQ(msg *sarama.ConsumerMessage, reason string) {
	if c.svcCtx.Config.KafkaConsumerConf.DLQTopic == "" {
		return
	}
	if err := c.svcCtx.KafkaProducer.SendRawMessage(&sarama.ProducerMessage{
		Topic: c.svcCtx.Config.KafkaConsumerConf.DLQTopic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(msg.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("x-source-topic"), Value: []byte(msg.Topic)},
			{Key: []byte("x-original-offset"), Value: []byte(strconv.FormatInt(msg.Offset, 10))},
			{Key: []byte("x-error"), Value: []byte(reason)},
		},
	}); err != nil {
		logx.Errorf("Failed to send message to DLQ: %v", err)
	}
}

func (c *InteractionConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *InteractionConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *InteractionConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()

	type starKey struct {
		PostID uint64
		UserID uint64
	}
	type collectionKey struct {
		PostID uint64
		UserID uint64
	}

	starMap := make(map[starKey]*model.PostStar)
	collectionMap := make(map[collectionKey]*model.PostCollection)
	unstarUserPostIds := make([][2]uint64, 0, batchSize)
	uncollectUserPostIds := make([][2]uint64, 0, batchSize)
	pendingMsgs := make([]*sarama.ConsumerMessage, 0, batchSize)
	lastFlushTime := time.Now()

	flushBatches := func() error {
		if len(starMap) > 0 {
			starBatch := make([]*model.PostStar, 0, len(starMap))
			for _, star := range starMap {
				starBatch = append(starBatch, star)
			}
			if err := c.svcCtx.PostStarDAO.BatchCreate(ctx, starBatch); err != nil {
				return err
			}
			starMap = make(map[starKey]*model.PostStar)
		}

		if len(collectionMap) > 0 {
			collectionBatch := make([]*model.PostCollection, 0, len(collectionMap))
			for _, collection := range collectionMap {
				collectionBatch = append(collectionBatch, collection)
			}
			if err := c.svcCtx.PostCollectionDAO.BatchCreate(ctx, collectionBatch); err != nil {
				return err
			}
			collectionMap = make(map[collectionKey]*model.PostCollection)
		}

		for _, ids := range unstarUserPostIds {
			if err := c.svcCtx.PostStarDAO.Delete(ctx, ids[0], ids[1]); err != nil {
				logx.Errorf("delete post star failed: %v", err)
			}
		}
		unstarUserPostIds = unstarUserPostIds[:0]

		for _, ids := range uncollectUserPostIds {
			if err := c.svcCtx.PostCollectionDAO.Delete(ctx, ids[0], ids[1]); err != nil {
				logx.Errorf("delete post collection failed: %v", err)
			}
		}
		uncollectUserPostIds = uncollectUserPostIds[:0]

		for _, msg := range pendingMsgs {
			session.MarkMessage(msg, "")
		}
		pendingMsgs = pendingMsgs[:0]
		lastFlushTime = time.Now()
		return nil
	}

	for msg := range claim.Messages() {
		var envelope *mq.EventEnvelope
		switch msg.Topic {
		case "post-star":
			var starMsg mq.PostStarMessage
			if env, err := mq.DecodeEnvelopePayload(msg.Value, &starMsg); err != nil {
				logx.Errorf("failed to decode post-star message: %v", err)
			} else {
				envelope = env
				key := starKey{PostID: starMsg.PostID, UserID: starMsg.UserID}
				if _, exists := starMap[key]; !exists {
					starMap[key] = &model.PostStar{
						PostID:  starMsg.PostID,
						UserID:  starMsg.UserID,
						LikedAt: time.Unix(starMsg.Timestamp, 0),
					}
				}
			}
		case "post-unstar":
			var unstarMsg mq.PostUnstarMessage
			if env, err := mq.DecodeEnvelopePayload(msg.Value, &unstarMsg); err != nil {
				logx.Errorf("failed to decode post-unstar message: %v", err)
			} else {
				envelope = env
				unstarUserPostIds = append(unstarUserPostIds, [2]uint64{unstarMsg.PostID, unstarMsg.UserID})
			}
		case "post-collect":
			var collectMsg mq.PostCollectMessage
			if env, err := mq.DecodeEnvelopePayload(msg.Value, &collectMsg); err != nil {
				logx.Errorf("failed to decode post-collect message: %v", err)
			} else {
				envelope = env
				key := collectionKey{PostID: collectMsg.PostID, UserID: collectMsg.UserID}
				if _, exists := collectionMap[key]; !exists {
					collectedAt := time.Now()
					if collectMsg.Timestamp > 0 {
						collectedAt = time.Unix(collectMsg.Timestamp, 0)
					}
					collectionMap[key] = &model.PostCollection{
						PostID:      collectMsg.PostID,
						UserID:      collectMsg.UserID,
						CollectedAt: collectedAt,
					}
				}
			}
		case "post-uncollect":
			var uncollectMsg mq.PostUncollectMessage
			if env, err := mq.DecodeEnvelopePayload(msg.Value, &uncollectMsg); err != nil {
				logx.Errorf("failed to decode post-uncollect message: %v", err)
			} else {
				envelope = env
				uncollectUserPostIds = append(uncollectUserPostIds, [2]uint64{uncollectMsg.PostID, uncollectMsg.UserID})
			}
		}

		if key := mq.BuildIdempotencyKey(c.svcCtx.Config.KafkaConsumerConf.Group, envelope); key != "" {
			ok, err := c.svcCtx.RedisClient.SetNX(ctx, key, "1", mq.DefaultEventTTL)
			if err == nil && !ok {
				session.MarkMessage(msg, "")
				continue
			}
		}

		pendingMsgs = append(pendingMsgs, msg)
		totalBatchSize := len(starMap) + len(collectionMap) + len(unstarUserPostIds) + len(uncollectUserPostIds)
		if totalBatchSize >= batchSize || time.Since(lastFlushTime) > 500*time.Millisecond {
			if err := flushBatches(); err != nil {
				for _, pending := range pendingMsgs {
					c.sendToDLQ(pending, err.Error())
				}
			}
		}
	}

	if err := flushBatches(); err != nil {
		logx.Errorf("flush final interaction batches failed: %v", err)
	}
	return nil
}
