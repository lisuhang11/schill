package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"SChill/service/canal/internal/model"
	"SChill/service/canal/internal/svc"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/logx"
)

type bulkAction string

const (
	bulkIndex  bulkAction = "index"
	bulkUpdate bulkAction = "update"
	bulkDelete bulkAction = "delete"
)

type bulkOperation struct {
	action bulkAction
	index  string
	docID  string
	doc    interface{}
}

type SyncHandler struct {
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	builder *PostDocumentBuilder
}

func NewSyncHandler(ctx context.Context, svcCtx *svc.ServiceContext) *SyncHandler {
	return &SyncHandler{
		ctx:     ctx,
		svcCtx:  svcCtx,
		builder: NewPostDocumentBuilder(svcCtx),
	}
}

func (h *SyncHandler) HandleMessages(msgs []*model.CanalMessage) error {
	ops := make([]bulkOperation, 0, len(msgs))
	pendingPostIDs := make([]uint64, 0, len(msgs))

	for _, msg := range msgs {
		messageOps, rebuildPostIDs, err := h.collectOperations(msg)
		if err != nil {
			logx.Errorf("collect canal operations failed: table=%s type=%s sql=%s data_count=%d err=%v",
				msg.Table, msg.Type, msg.SQL, len(msg.Data), err)
			return err
		}
		ops = append(ops, messageOps...)
		pendingPostIDs = append(pendingPostIDs, rebuildPostIDs...)
	}

	postDocs, err := h.builder.Build(h.ctx, pendingPostIDs)
	if err != nil {
		return err
	}
	for _, doc := range postDocs {
		ops = append(ops, bulkOperation{
			action: bulkIndex,
			index:  "post",
			docID:  strconv.FormatUint(doc.ID, 10),
			doc:    doc,
		})
	}

	if len(ops) == 0 {
		return nil
	}
	return h.bulkWriteToES(ops)
}

func (h *SyncHandler) collectOperations(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	if msg.IsDDL {
		logx.Infof("skip DDL message: %s", msg.SQL)
		return nil, nil, nil
	}

	switch msg.Table {
	case "user":
		return h.syncUser(msg)
	case "user_stat":
		return h.syncUserStat(msg)
	case "post":
		return h.syncPost(msg)
	case "post_content":
		return h.syncPostContent(msg)
	case "topic":
		return h.syncTopic(msg)
	case "post_topic":
		return h.syncPostTopic(msg)
	default:
		return nil, nil, nil
	}
}

func (h *SyncHandler) syncUser(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	ops := make([]bulkOperation, 0, len(msg.Data))
	userIDs := make([]uint64, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}

		id := h.getUint64(row, "id")
		if id == 0 {
			continue
		}
		userIDs = append(userIDs, id)
		docID := strconv.FormatUint(id, 10)
		if msg.Type == "DELETE" || (row["deleted_at"] != nil && row["deleted_at"] != "") {
			ops = append(ops, bulkOperation{action: bulkDelete, index: "user", docID: docID})
			continue
		}

		ops = append(ops, bulkOperation{
			action: bulkIndex,
			index:  "user",
			docID:  docID,
			doc: model.ESUser{
				ID:        id,
				Username:  h.getString(row, "username"),
				Status:    h.getInt8(row, "status"),
				IsAdmin:   h.getBool(row, "is_admin"),
				CreatedAt: h.getTimeUnix(row, "created_at"),
			},
		})
	}

	postIDs, err := h.builder.FindPostIDsByUserIDs(h.ctx, userIDs)
	return ops, postIDs, err
}

func (h *SyncHandler) syncUserStat(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	ops := make([]bulkOperation, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}

		userID := h.getUint64(row, "user_id")
		if userID == 0 {
			continue
		}
		ops = append(ops, bulkOperation{
			action: bulkUpdate,
			index:  "user",
			docID:  strconv.FormatUint(userID, 10),
			doc: map[string]interface{}{
				"post_count":       h.getInt32(row, "post_count"),
				"comment_count":    h.getInt32(row, "comment_count"),
				"follower_count":   h.getInt32(row, "follower_count"),
				"like_count":       h.getInt32(row, "like_count"),
				"collection_count": h.getInt32(row, "collection_count"),
				"last_active_time": h.getInt64(row, "last_active_time"),
			},
		})
	}
	return ops, nil, nil
}

func (h *SyncHandler) syncPost(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	ops := make([]bulkOperation, 0, len(msg.Data))
	postIDs := make([]uint64, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}

		id := h.getUint64(row, "id")
		if id == 0 {
			continue
		}
		if msg.Type == "DELETE" || (row["deleted_at"] != nil && row["deleted_at"] != "") {
			ops = append(ops, bulkOperation{action: bulkDelete, index: "post", docID: strconv.FormatUint(id, 10)})
			continue
		}
		postIDs = append(postIDs, id)
	}
	return ops, postIDs, nil
}

func (h *SyncHandler) syncPostContent(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	postIDs := make([]uint64, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}
		postIDs = append(postIDs, h.getUint64(row, "post_id"))
	}
	return nil, postIDs, nil
}

func (h *SyncHandler) syncTopic(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	ops := make([]bulkOperation, 0, len(msg.Data))
	topicIDs := make([]uint64, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}

		id := h.getUint64(row, "id")
		if id == 0 {
			continue
		}
		topicIDs = append(topicIDs, id)
		docID := strconv.FormatUint(id, 10)
		if msg.Type == "DELETE" || (row["deleted_at"] != nil && row["deleted_at"] != "") {
			ops = append(ops, bulkOperation{action: bulkDelete, index: "topic", docID: docID})
			continue
		}

		ops = append(ops, bulkOperation{
			action: bulkIndex,
			index:  "topic",
			docID:  docID,
			doc: model.ESTopic{
				ID:        id,
				Name:      h.getString(row, "name"),
				QuoteNum:  h.getInt64(row, "quote_num"),
				CreatedAt: h.getTimeUnix(row, "created_at"),
			},
		})
	}
	postIDs, err := h.builder.FindPostIDsByTopicIDs(h.ctx, topicIDs)
	return ops, postIDs, err
}

func (h *SyncHandler) syncPostTopic(msg *model.CanalMessage) ([]bulkOperation, []uint64, error) {
	postIDs := make([]uint64, 0, len(msg.Data))
	for _, data := range msg.Data {
		row := map[string]interface{}{}
		if err := json.Unmarshal(data, &row); err != nil {
			return nil, nil, err
		}
		postIDs = append(postIDs, h.getUint64(row, "post_id"))
	}
	return nil, postIDs, nil
}

func (h *SyncHandler) bulkWriteToES(ops []bulkOperation) error {
	var body bytes.Buffer

	for _, op := range ops {
		meta := map[string]map[string]string{
			string(op.action): {
				"_index": op.index,
				"_id":    op.docID,
			},
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		body.Write(metaBytes)
		body.WriteByte('\n')

		if op.action == bulkDelete {
			continue
		}

		var payload []byte
		if op.action == bulkUpdate {
			payload, err = json.Marshal(map[string]interface{}{"doc": op.doc, "doc_as_upsert": true})
		} else {
			payload, err = json.Marshal(op.doc)
		}
		if err != nil {
			return err
		}
		body.Write(payload)
		body.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    bytes.NewReader(body.Bytes()),
		Refresh: "false",
	}
	res, err := req.Do(h.ctx, h.svcCtx.ESClient)
	if err != nil {
		logx.Errorf("es bulk request failed: op_count=%d err=%v", len(ops), err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		logx.Errorf("es bulk response error: op_count=%d status=%s", len(ops), res.String())
		return fmt.Errorf("es bulk error: %s", res.String())
	}

	return nil
}

func (h *SyncHandler) getString(row map[string]interface{}, key string) string {
	if v, ok := row[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (h *SyncHandler) getUint64(row map[string]interface{}, key string) uint64 {
	if v, ok := row[key]; ok {
		switch val := v.(type) {
		case float64:
			return uint64(val)
		case string:
			if id, err := strconv.ParseUint(val, 10, 64); err == nil {
				return id
			}
		}
	}
	return 0
}

func (h *SyncHandler) getInt64(row map[string]interface{}, key string) int64 {
	if v, ok := row[key]; ok {
		switch val := v.(type) {
		case float64:
			return int64(val)
		case string:
			if id, err := strconv.ParseInt(val, 10, 64); err == nil {
				return id
			}
		}
	}
	return 0
}

func (h *SyncHandler) getInt32(row map[string]interface{}, key string) int32 {
	return int32(h.getInt64(row, key))
}

func (h *SyncHandler) getInt8(row map[string]interface{}, key string) int8 {
	return int8(h.getInt64(row, key))
}

func (h *SyncHandler) getBool(row map[string]interface{}, key string) bool {
	if v, ok := row[key]; ok {
		switch val := v.(type) {
		case float64:
			return val == 1
		case string:
			return val == "1" || val == "true"
		case bool:
			return val
		}
	}
	return false
}

func (h *SyncHandler) getTimeUnix(row map[string]interface{}, key string) int64 {
	if v, ok := row[key]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.UnixMilli()
		}
		if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func normalizeLegacyTags(tags string) []string {
	if tags == "" {
		return nil
	}
	parts := strings.Split(tags, ",")
	return uniqueStrings(parts)
}
