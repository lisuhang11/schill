package handler

import (
	"net/http"
	"strings"

	"SChill/service/gateway/internal/svc"
)

func SearchProxyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		svcCtx.SearchProxy.ServeHTTP(w, r)
	}
}

func HealthHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		ok(w, map[string]string{"service": "gateway"})
	}
}
