package logic

const (
	maxBatchStatusPostIDs = 256
	batchStatusChunkSize  = 64
)

func uniquePostIDs(postIDs []uint64) []uint64 {
	if len(postIDs) == 0 {
		return nil
	}

	seen := make(map[uint64]struct{}, len(postIDs))
	result := make([]uint64, 0, len(postIDs))
	for _, postID := range postIDs {
		if postID == 0 {
			continue
		}
		if _, ok := seen[postID]; ok {
			continue
		}
		seen[postID] = struct{}{}
		result = append(result, postID)
	}
	return result
}
