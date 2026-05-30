package logic

import (
	"SChill/service/canal/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type CanalService struct {
	svcCtx   *svc.ServiceContext
	consumer *CanalConsumer
}

func NewCanalService(svcCtx *svc.ServiceContext) *CanalService {
	return &CanalService{
		svcCtx:   svcCtx,
		consumer: NewCanalConsumer(svcCtx),
	}
}

func (s *CanalService) Start() {
	logx.Infof("Starting Canal ES Sync Service...")
	if err := s.consumer.Start(); err != nil {
		logx.Errorf("failed to start canal consumer: %v", err)
	}
}

func (s *CanalService) Stop() {
	logx.Infof("Stopping Canal ES Sync Service...")
	s.consumer.Stop()
}

var _ service.Service = (*CanalService)(nil)
