package logic

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type InteractionConsumerService struct {
	consumer *InteractionConsumer
}

func NewInteractionConsumerService(consumer *InteractionConsumer) *InteractionConsumerService {
	return &InteractionConsumerService{consumer: consumer}
}

func (s *InteractionConsumerService) Start() {
	if err := s.consumer.Start(); err != nil {
		logx.Errorf("start interaction consumer failed: %v", err)
	}
}

func (s *InteractionConsumerService) Stop() {
	s.consumer.Stop()
}

var _ service.Service = (*InteractionConsumerService)(nil)
