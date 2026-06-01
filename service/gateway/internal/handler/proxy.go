package handler

import (
	"net/http"

	"SChill/service/gateway/internal/svc"
)

func HealthHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		ok(w, map[string]string{"service": "gateway"})
	}
}
