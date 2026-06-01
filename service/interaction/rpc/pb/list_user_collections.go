package pb

// ListUserCollectionsReq 获取用户收藏列表请求（手动添加，非 proto 生成）
type ListUserCollectionsReq struct {
	UserId   uint64
	Page     int64
	PageSize int64
}

// ListUserCollectionsResp 获取用户收藏列表响应（手动添加，非 proto 生成）
type ListUserCollectionsResp struct {
	PostIds []uint64
	Total   int64
}
