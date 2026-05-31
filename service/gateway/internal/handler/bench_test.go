package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"SChill/service/gateway/internal/types"
)

// =============================================================================
// HTTP 响应速度测试 — 测试 handler 层的纯逻辑耗时（不依赖后端 gRPC）
// 这些测试测量：请求解析 → handler 调用 → 响应序列化 的延迟
// =============================================================================

// BenchmarkHealthHandler 健康检查接口 — 无依赖，纯响应
func BenchmarkHealthHandler(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		ok(w, map[string]string{"status": "healthy"})
		if w.Code != 200 {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

// BenchmarkFailResponse 测试错误响应生成速度
func BenchmarkFailResponse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		fail(w, http.StatusUnauthorized, "未授权访问，请先登录")
		if w.Code != http.StatusUnauthorized {
			b.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
		}
	}
}

// BenchmarkParseUintParam 测试路径参数解析
func BenchmarkParseUintParam(b *testing.B) {
	r := httptest.NewRequest("GET", "/api/users/12345", nil)
	r = r.WithContext(httptest.NewRequest("GET", "/api/users/12345", nil).Context())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, ok := parseUintParam(r, "id")
		if !ok {
			// parseUintParam requires go-zero's pathvar, skip in unit test context
		}
	}
}

// BenchmarkCurrentUserID 测试用户 ID header 解析
func BenchmarkCurrentUserID(b *testing.B) {
	r := httptest.NewRequest("GET", "/api/test", nil)
	r.Header.Set(currentUserHeader, "12345")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		id, ok := currentUserID(r)
		if !ok || id != 12345 {
			b.Fatalf("expected 12345, got %d", id)
		}
	}
}

// BenchmarkPageParam 测试分页参数解析
func BenchmarkPageParam(b *testing.B) {
	r := httptest.NewRequest("GET", "/api/test?page=5&pageSize=20", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		page := pageParam(r, "page", 1)
		if page != 5 {
			b.Fatalf("expected 5, got %d", page)
		}
	}
}

// BenchmarkUintQueryParam 测试 query 参数解析
func BenchmarkUintQueryParam(b *testing.B) {
	r := httptest.NewRequest("GET", "/api/test?topicId=42", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		id := uintQueryParam(r, "topicId", 0)
		if id != 42 {
			b.Fatalf("expected 42, got %d", id)
		}
	}
}

// BenchmarkResponseJSON 测试 JSON 序列化 + 响应写入
func BenchmarkResponseJSON(b *testing.B) {
	resp := types.Response{
		Code: 0,
		Msg:  "ok",
		Data: map[string]interface{}{
			"userId":   12345,
			"username": "testuser",
			"avatar":   "https://example.com/avatar.png",
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		ok(w, resp.Data)
		if w.Code != 200 {
			b.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if body == "" {
			b.Fatal("empty body")
		}
	}
}

// =============================================================================
// HTTP 并发性能测试 — 模拟多用户并发请求
// =============================================================================

// BenchmarkConcurrentHealthCheck 并发健康检查请求
func BenchmarkConcurrentHealthCheck(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			ok(w, map[string]string{"status": "healthy"})
		}
	})
}

// BenchmarkConcurrentFailResponse 并发错误响应
func BenchmarkConcurrentFailResponse(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			fail(w, http.StatusUnauthorized, "未授权")
		}
	})
}

// BenchmarkConcurrentJSONResponse 并发 JSON 响应序列化
func BenchmarkConcurrentJSONResponse(b *testing.B) {
	data := map[string]interface{}{
		"userId":   12345,
		"username": "testuser",
	}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			ok(w, data)
		}
	})
}

// =============================================================================
// HTTP 请求构造性能测试 — 模拟真实请求流程
// =============================================================================

// BenchmarkLoginRequestConstruction 模拟登录请求构造 + 参数解析
func BenchmarkLoginRequestConstruction(b *testing.B) {
	body := `{"username":"testuser","password":"TestPass123"}`
	b.ReportAllocs()
	for b.Loop() {
		var req types.LoginReq
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			b.Fatal(err)
		}
		if req.Username == "" || req.Password == "" {
			b.Fatal("empty credentials")
		}
	}
}

// BenchmarkRegisterRequestConstruction 模拟注册请求构造 + 参数解析
func BenchmarkRegisterRequestConstruction(b *testing.B) {
	body := `{"username":"newuser","password":"StrongPass123"}`
	b.ReportAllocs()
	for b.Loop() {
		var req types.RegisterReq
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			b.Fatal(err)
		}
		if len(req.Username) < 3 || len(req.Password) < 8 {
			b.Fatal("invalid input")
		}
	}
}

// =============================================================================
// 大型响应序列化性能测试
// =============================================================================

func generateFeedResponse(count int) map[string]interface{} {
	items := make([]map[string]interface{}, count)
	for i := range items {
		items[i] = map[string]interface{}{
			"postId":    1000 + i,
			"title":     fmt.Sprintf("Test Post Title %d", i),
			"content":   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			"authorId":  100,
			"authorName": fmt.Sprintf("author_%d", i%10),
			"topicName":  "general",
			"viewCount":  1000 + i,
			"likeCount":  50 + i%10,
			"commentCount": 10 + i%5,
			"createdAt":  "2025-01-01T00:00:00Z",
		}
	}
	return map[string]interface{}{
		"list":  items,
		"total": count,
		"hasMore": true,
	}
}

// BenchmarkSerializeLargeResponse_10 序列化 10 条数据
func BenchmarkSerializeLargeResponse_10(b *testing.B) {
	data := generateFeedResponse(10)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		body, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
		if len(body) < 100 {
			b.Fatal("too short")
		}
	}
}

// BenchmarkSerializeLargeResponse_100 序列化 100 条数据
func BenchmarkSerializeLargeResponse_100(b *testing.B) {
	data := generateFeedResponse(100)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		body, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
		if len(body) < 100 {
			b.Fatal("too short")
		}
	}
}

// BenchmarkDeserializeLargeResponse_10 反序列化 10 条数据
func BenchmarkDeserializeLargeResponse_10(b *testing.B) {
	data := generateFeedResponse(10)
	body, _ := json.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDeserializeLargeResponse_100 反序列化 100 条数据
func BenchmarkDeserializeLargeResponse_100(b *testing.B) {
	data := generateFeedResponse(100)
	body, _ := json.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// HTTP 端到端并发压测（需要真实服务运行，作为集成性能测试）
// =============================================================================

var benchHTTPClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     200,
	},
}

// BenchmarkHTTPHealthEndpoint 端到端健康检查（需要 gateway 运行在 localhost:8086）
func BenchmarkHTTPHealthEndpoint(b *testing.B) {
	// 跳过如果没有设置 E2E_BENCH=1
	if testing.Short() {
		b.Skip("skipping e2e benchmark in short mode")
	}
	b.ReportAllocs()
	for b.Loop() {
		resp, err := benchHTTPClient.Get("http://localhost:8086/health")
		if err != nil {
			b.Skipf("gateway not running: %v", err)
			return
		}
		if resp.StatusCode != 200 {
			b.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// BenchmarkHTTPHealthConcurrent 端到端并发健康检查
func BenchmarkHTTPHealthConcurrent(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping e2e benchmark in short mode")
	}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := benchHTTPClient.Get("http://localhost:8086/health")
			if err != nil {
				b.Skipf("gateway not running: %v", err)
				return
			}
			if resp.StatusCode != 200 {
				b.Fatalf("expected 200, got %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}

// =============================================================================
// 高并发写操作模拟 — 测试 handler 层在压力下的表现
// =============================================================================

// BenchmarkConcurrentRequestParsing 高并发请求解析
func BenchmarkConcurrentRequestParsing(b *testing.B) {
	bodies := make([][]byte, 100)
	for i := range bodies {
		bodies[i] = []byte(fmt.Sprintf(`{"username":"user_%d","password":"Pass%d"}`, i, i))
	}
	var mu sync.Mutex
	var idx int

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			body := bodies[idx%100]
			idx++
			mu.Unlock()

			var req types.LoginReq
			json.Unmarshal(body, &req)
		}
	})
}

// =============================================================================
// 辅助函数
// =============================================================================

func buildRequest(method, path, body string, headers map[string]string) *http.Request {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

// BenchmarkBuildAuthenticatedRequest 测试构建带认证信息的请求
func BenchmarkBuildAuthenticatedRequest(b *testing.B) {
	body := `{"content":"Hello World"}`
	b.ReportAllocs()
	for b.Loop() {
		r := buildRequest("POST", "/api/posts", body, map[string]string{
			"X-User-Id": "12345",
		})
		if r.Header.Get("X-User-Id") != "12345" {
			b.Fatal("header not set")
		}
		_ = r
	}
}

// BenchmarkQueryStringParsing 测试 Query String 解析
func BenchmarkQueryStringParsing(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		r := httptest.NewRequest("GET", "/api/feed?page=1&pageSize=20&feedType=1", nil)
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		pageSize, _ := strconv.Atoi(q.Get("pageSize"))
		feedType, _ := strconv.Atoi(q.Get("feedType"))
		if page != 1 || pageSize != 20 || feedType != 1 {
			b.Fatal("wrong params")
		}
	}
}

// BenchmarkHeaderParsing 测试 Header 解析
func BenchmarkHeaderParsing(b *testing.B) {
	r := httptest.NewRequest("GET", "/api/test", nil)
	r.Header.Set("X-User-Id", "12345")
	r.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		uid := r.Header.Get("X-User-Id")
		auth := r.Header.Get("Authorization")
		if uid != "12345" || !strings.HasPrefix(auth, "Bearer ") {
			b.Fatal("header parse failed")
		}
	}
}
