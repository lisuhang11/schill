package es

import (
	"context"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Client struct {
	*elasticsearch.Client
}

type Config struct {
	Hosts    []string
	Username string
	Password string
}

func NewClient(cfg Config) (*Client, error) {
	esConfig := elasticsearch.Config{
		Addresses: cfg.Hosts,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Printf("Elasticsearch connected: %s", res.Status())

	return &Client{Client: client}, nil
}

func (c *Client) Search(ctx context.Context, index string, query string) (*esapi.Response, error) {
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}

	return req.Do(ctx, c)
}
