package observability

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func HTTPAccessLog(service string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next(recorder, r)

			latency := time.Since(start)
			fields := []logx.LogField{
				logx.Field("service", service),
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("status", recorder.status),
				logx.Field("latency_ms", latency.Milliseconds()),
			}

			switch {
			case recorder.status >= http.StatusInternalServerError:
				logx.WithContext(r.Context()).Errorw("api request failed", fields...)
			case latency > 200*time.Millisecond:
				logx.WithContext(r.Context()).Sloww("api request slow", fields...)
			default:
				logx.WithContext(r.Context()).Infow("api request", fields...)
			}
		}
	}
}
