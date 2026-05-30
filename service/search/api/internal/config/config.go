package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	Elasticsearch struct {
		Hosts     []string
		Username  string
		Password  string
		PostIndex string
		UserIndex string
		TopicIndex string
	}

	Redis struct {
		Host     string
		Port     string
		Password string
		DB       int
	}
}
