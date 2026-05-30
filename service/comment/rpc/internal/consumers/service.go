package consumers

import (
	"github.com/zeromicro/go-zero/core/service"
)

type CommentConsumerService struct {
	consumer *CommentConsumer
}

func NewCommentConsumerService(consumer *CommentConsumer) *CommentConsumerService {
	return &CommentConsumerService{consumer: consumer}
}

func (s *CommentConsumerService) Start() {
	go func() {
		s.consumer.Start()
	}()
}

func (s *CommentConsumerService) Stop() {
	s.consumer.Stop()
}

var _ service.Service = (*CommentConsumerService)(nil)
