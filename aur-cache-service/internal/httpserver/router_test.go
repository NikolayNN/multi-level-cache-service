package httpserver

import (
	"aur-cache-service/api/dto"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAdapter struct {
	getCalled      []*dto.CacheId
	putCalled      []*dto.CacheEntry
	evictCalled    []*dto.CacheId
	getAllCalled   [][]*dto.CacheId
	putAllCalled   [][]*dto.CacheEntry
	evictAllCalled [][]*dto.CacheId

	getResult     *dto.CacheEntryHit
	getAllResults []*dto.CacheEntryHit
}

func (m *mockAdapter) Get(_ context.Context, id *dto.CacheId) *dto.CacheEntryHit {
	m.getCalled = append(m.getCalled, id)
	return m.getResult
}

func (m *mockAdapter) Put(_ context.Context, e *dto.CacheEntry) {
	m.putCalled = append(m.putCalled, e)
}

func (m *mockAdapter) Evict(_ context.Context, id *dto.CacheId) {
	m.evictCalled = append(m.evictCalled, id)
}

func (m *mockAdapter) GetAll(_ context.Context, ids []*dto.CacheId) []*dto.CacheEntryHit {
	m.getAllCalled = append(m.getAllCalled, ids)
	return m.getAllResults
}

func (m *mockAdapter) PutAll(_ context.Context, entries []*dto.CacheEntry) {
	m.putAllCalled = append(m.putAllCalled, entries)
}

func (m *mockAdapter) EvictAll(_ context.Context, ids []*dto.CacheId) {
	m.evictAllCalled = append(m.evictAllCalled, ids)
}

func TestParsePath(t *testing.T) {
	cases := []struct {
		path  string
		cache string
		key   string
		ok    bool
	}{
		{"/api/cache/a/b", "a", "b", true},
		{"/api/cache/a/a%2Fb", "a", "a/b", true},
		{"/wrong", "", "", false},
	}
	for _, c := range cases {
		cache, key, ok := parsePath(c.path)
		if ok != c.ok || cache != c.cache || key != c.key {
			t.Errorf("parsePath(%q)=%q,%q,%v want %q,%q,%v", c.path, cache, key, ok, c.cache, c.key, c.ok)
		}
	}
}

func TestHandleSingleGet(t *testing.T) {
	hitVal := json.RawMessage(`"v"`)
	adapter := &mockAdapter{getResult: &dto.CacheEntryHit{CacheEntry: &dto.CacheEntry{CacheId: &dto.CacheId{CacheName: "c", Key: "k"}, Value: &hitVal}, Found: true}}
	router := NewRouter(adapter)
	req := httptest.NewRequest(http.MethodGet, "/api/cache/c/k", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Fatalf("content-type=%s", ct)
	}
	var res dto.CacheEntryHit
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !res.Found {
		t.Fatalf("expected found")
	}
}

func TestHandleSingleGetNotFound(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	req := httptest.NewRequest(http.MethodGet, "/api/cache/c/miss", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestHandleSinglePut(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	body := bytes.NewBufferString(`{"x":1}`)
	req := httptest.NewRequest(http.MethodPut, "/api/cache/c/k", body)
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if len(adapter.putCalled) != 1 {
		t.Fatalf("put not called")
	}
	if adapter.putCalled[0].CacheId.CacheName != "c" || adapter.putCalled[0].CacheId.Key != "k" {
		t.Fatalf("wrong id")
	}
}

func TestHandleSingleDelete(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	req := httptest.NewRequest(http.MethodDelete, "/api/cache/c/k", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if len(adapter.evictCalled) != 1 {
		t.Fatalf("evict not called")
	}
}

func TestHandleBatchGet(t *testing.T) {
	adapter := &mockAdapter{}
	adapter.getAllResults = []*dto.CacheEntryHit{
		{CacheEntry: &dto.CacheEntry{CacheId: &dto.CacheId{CacheName: "c", Key: "1"}}, Found: true},
		nil,
	}
	router := NewRouter(adapter)
	body := bytes.NewBufferString(`{"requests":[{"c":"c","k":"1"},{"c":"c","k":"2"}]}`)
	req := httptest.NewRequest(http.MethodPost, batchGetPath, body)
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if len(adapter.getAllCalled) != 1 {
		t.Fatalf("getAll not called")
	}
	var resp struct {
		Results []dto.CacheEntryHit `json:"results"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) != 2 || !resp.Results[0].Found || resp.Results[1].Found {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
}

func TestHandleBatchPut(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	body := bytes.NewBufferString(`{"requests":[{"c":"c","k":"1","v":1}]}`)
	req := httptest.NewRequest(http.MethodPost, batchPutPath, body)
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if len(adapter.putAllCalled) != 1 {
		t.Fatalf("putAll not called")
	}
}

func TestHandleBatchDelete(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	body := bytes.NewBufferString(`{"requests":[{"c":"c","k":"1"}]}`)
	req := httptest.NewRequest(http.MethodPost, batchDeletePath, body)
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if len(adapter.evictAllCalled) != 1 {
		t.Fatalf("evictAll not called")
	}
}

func TestBodyLimit(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	big := bytes.Repeat([]byte("a"), maxBodySize+1)
	req := httptest.NewRequest(http.MethodPost, batchPutPath, bytes.NewReader(big))
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestGzipDecompress(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	io.WriteString(gz, `{"requests":[]}`)
	gz.Close()
	req := httptest.NewRequest(http.MethodPost, batchPutPath, &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rr.Code)
	}
}
