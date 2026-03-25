package logic

import (
	errutil "SChill/common/error"
	minioUtil "SChill/common/minio"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAvatarLogic {
	return &UpdateAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAvatarLogic) UpdateAvatar(req *types.UpdateAvatarReq) (resp *types.UpdateAvatarResp, err error) {
	logx.Infof("收到更新头像请求，avatarName: [%s]", req.AvatarName)
	// 从上下文获取UserId
	userId, err := l.getUserIdFromContext()
	if err != nil {
		return &types.UpdateAvatarResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	// 校验文件
	if len(req.Avatar) == 0 {
		return &types.UpdateAvatarResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	// 校验文件类型（仅图片）
	mimeType := http.DetectContentType([]byte(req.Avatar[:512]))
	if !strings.HasPrefix(mimeType, "image/") {
		return &types.UpdateAvatarResp{
			Code: errutil.ErrInvalidParams,
			Msg:  "仅支持图片文件",
		}, nil
	}

	// 限制文件大小（例如 10MB）
	const maxSize = 10 * 1024 * 1024
	if len(req.Avatar) > maxSize {
		return &types.UpdateAvatarResp{
			Code: errutil.ErrInvalidParams,
			Msg:  "文件大小不能超过5MB",
		}, nil
	}

	// 生成对象名
	objectName := minioUtil.GenerateMinIOObjectName(userId, mimeType)

	// 上传到 MinIO
	reader := bytes.NewReader(req.Avatar) // req.Avatar 已经是 []byte，直接使用
	fileSize := int64(len(req.Avatar))

	_, err = l.svcCtx.MinIO.PutObject(l.ctx, l.svcCtx.Config.MinIO.Bucket, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		logx.Errorf("上传到 MinIO 失败: %v", err)
		return &types.UpdateAvatarResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	// 构造公开访问的 URL（假设存储桶已设置为公开读）
	// 注意：Endpoint 需要是外部可访问的地址，如 "http://minio.example.com:9000" 或使用 Nginx 反向代理后的域名
	objectURL := fmt.Sprintf("http://%s/%s/%s",
		l.svcCtx.Config.MinIO.Endpoint, // 例如 "localhost:9000"
		l.svcCtx.Config.MinIO.Bucket,
		objectName)

	// 调用 RPC 更新数据库
	rpcResp, err := l.svcCtx.UserRpc.UpdateAvatar(l.ctx, &pb.UpdateAvatarReq{
		UserId:    userId,
		AvatarUrl: objectURL,
	})
	if err != nil {
		logx.Errorf("调用 RPC UpdateAvatar 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.UpdateAvatarResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.UpdateAvatarResp{
		Code:   errutil.Success,
		Msg:    errutil.GetCodeMessage(errutil.Success),
		Avatar: rpcResp.AvatarUrl,
	}, nil
}

func (l *UpdateAvatarLogic) getUserIdFromContext() (uint64, error) {
	logx.Infof("尝试从Context获取userId...")

	userIdVal := l.ctx.Value("userId")
	logx.Infof("userId值: %v, 类型: %T", userIdVal, userIdVal)

	switch v := userIdVal.(type) {
	case uint64:
		logx.Infof("从userId key获取到(uint64): %d", v)
		return v, nil
	case int64:
		logx.Infof("从userId key获取到(int64): %d", v)
		return uint64(v), nil
	case int:
		logx.Infof("从userId key获取到(int): %d", v)
		return uint64(v), nil
	case float64:
		logx.Infof("从userId key获取到(float64): %f", v)
		return uint64(v), nil
	case json.Number:
		logx.Infof("从userId key获取到(json.Number): %s", v)
		if userId, err := v.Int64(); err == nil {
			return uint64(userId), nil
		}
		if userId, err := v.Float64(); err == nil {
			return uint64(userId), nil
		}
	}

	logx.Errorf("无法从Context获取userId")
	return 0, nil
}
