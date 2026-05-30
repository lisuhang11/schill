package logic

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"SChill/service/canal/internal/model"
	"SChill/service/canal/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	canalBatchSize    = 100
	canalFlushTimeout = time.Second
)

type CanalConsumer struct {
	svcCtx        *svc.ServiceContext
	handler       *SyncHandler
	ctx           context.Context
	cancel        context.CancelFunc
	consumerGroup sarama.ConsumerGroup
	wg            sync.WaitGroup
}

func NewCanalConsumer(svcCtx *svc.ServiceContext) *CanalConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CanalConsumer{
		svcCtx:  svcCtx,
		handler: NewSyncHandler(context.Background(), svcCtx),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (c *CanalConsumer) Start() error {
	group, err := sarama.NewConsumerGroup(c.svcCtx.Config.Kafka.Brokers, c.svcCtx.Config.Kafka.Group, c.svcCtx.KafkaClient.Config())
	if err != nil {
		return err
	}
	c.consumerGroup = group
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(c.ctx, []string{c.svcCtx.Config.Kafka.Topic}, c); err != nil && c.ctx.Err() == nil {
				logx.Errorf("canal consumer error: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
		}
	}()

	logx.Infof("Canal consumer started, topic: %s, group: %s", c.svcCtx.Config.Kafka.Topic, c.svcCtx.Config.Kafka.Group)
	return nil
}

func (c *CanalConsumer) Stop() {
	c.cancel()
	if c.consumerGroup != nil {
		_ = c.consumerGroup.Close()
	}
	c.wg.Wait()
}

func (c *CanalConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *CanalConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *CanalConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batch := make([]*model.CanalMessage, 0, canalBatchSize)
	pending := make([]*sarama.ConsumerMessage, 0, canalBatchSize)
	lastFlush := time.Now()

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := c.handler.HandleMessages(batch); err != nil {
			return err
		}
		for _, msg := range pending {
			session.MarkMessage(msg, "")
		}
		batch = batch[:0]
		pending = pending[:0]
		lastFlush = time.Now()
		return nil
	}

	for msg := range claim.Messages() {
		var canalMsg model.CanalMessage
		if err := json.Unmarshal(msg.Value, &canalMsg); err != nil {
			logx.Errorf("failed to unmarshal canal message: topic=%s partition=%d offset=%d err=%v", msg.Topic, msg.Partition, msg.Offset, err)
			session.MarkMessage(msg, "")
			continue
		}

		batch = append(batch, &canalMsg)
		pending = append(pending, msg)

		if len(batch) >= canalBatchSize || time.Since(lastFlush) >= canalFlushTimeout {
			if err := flush(); err != nil {
				logx.Errorf("flush canal batch failed: %v", err)
			}
		}
	}

	if err := flush(); err != nil {
		logx.Errorf("flush final canal batch failed: %v", err)
	}
	return nil
}
