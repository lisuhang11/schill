package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"SChill/service/canal/internal/config"
	"SChill/service/canal/internal/logic"
	"SChill/service/canal/internal/model"
	"SChill/service/canal/internal/svc"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/conf"
)

var (
	configFile = flag.String("f", "service/canal/etc/canal.yaml", "config file")
	indexName  = flag.String("index", "", "target index name")
	aliasName  = flag.String("alias", "post", "target alias")
	pageSize   = flag.Int("page-size", 200, "page size")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	svcCtx := svc.NewServiceContext(c)
	builder := logic.NewPostDocumentBuilder(svcCtx)
	ctx := context.Background()

	targetIndex := *indexName
	if targetIndex == "" {
		targetIndex = fmt.Sprintf("post_v2_%s", time.Now().Format("20060102150405"))
	}

	mappingPath := filepath.Join("es_index", "post_index.json")
	mappingBytes, err := os.ReadFile(mappingPath)
	if err != nil {
		panic(err)
	}

	if res, err := (esapi.IndicesCreateRequest{
		Index: targetIndex,
		Body:  bytes.NewReader(mappingBytes),
	}).Do(ctx, svcCtx.ESClient); err != nil {
		panic(err)
	} else {
		res.Body.Close()
	}

	var maxID uint64
	if err := svcCtx.DB.WithContext(ctx).Table("post").Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error; err != nil {
		panic(err)
	}

	for start := uint64(1); start <= maxID; start += uint64(*pageSize) {
		end := start + uint64(*pageSize) - 1
		var postIDs []uint64
		if err := svcCtx.DB.WithContext(ctx).
			Table("post").
			Where("id BETWEEN ? AND ? AND deleted_at IS NULL", start, end).
			Order("id ASC").
			Pluck("id", &postIDs).Error; err != nil {
			panic(err)
		}
		if len(postIDs) == 0 {
			continue
		}

		docs, err := builder.Build(ctx, postIDs)
		if err != nil {
			panic(err)
		}
		if err := bulkIndex(ctx, svcCtx, targetIndex, docs); err != nil {
			panic(err)
		}
		fmt.Printf("indexed posts: %d-%d count=%d\n", start, end, len(docs))
	}

	if err := switchAlias(ctx, svcCtx, *aliasName, targetIndex); err != nil {
		panic(err)
	}
}

func bulkIndex(ctx context.Context, svcCtx *svc.ServiceContext, index string, docs []model.ESPost) error {
	var body bytes.Buffer
	for _, doc := range docs {
		metaBytes, _ := json.Marshal(map[string]map[string]string{
			"index": {"_index": index, "_id": fmt.Sprintf("%d", doc.ID)},
		})
		body.Write(metaBytes)
		body.WriteByte('\n')
		docBytes, _ := json.Marshal(doc)
		body.Write(docBytes)
		body.WriteByte('\n')
	}

	res, err := esapi.BulkRequest{Body: bytes.NewReader(body.Bytes()), Refresh: "false"}.Do(ctx, svcCtx.ESClient)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.String())
	}
	return nil
}

func switchAlias(ctx context.Context, svcCtx *svc.ServiceContext, alias, index string) error {
	reqBody := map[string]interface{}{
		"actions": []map[string]interface{}{
			{"remove": map[string]interface{}{"index": "*", "alias": alias}},
			{"add": map[string]interface{}{"index": index, "alias": alias}},
		},
	}
	data, _ := json.Marshal(reqBody)
	res, err := esapi.IndicesUpdateAliasesRequest{Body: bytes.NewReader(data)}.Do(ctx, svcCtx.ESClient)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("switch alias failed: %s", res.String())
	}
	return nil
}
