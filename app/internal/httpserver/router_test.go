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
	"strings"
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

func TestHandleBatchGet(t *testing.T) {
	adapter := &mockAdapter{}
	adapter.getAllResults = []*dto.CacheEntryHit{
		{CacheEntry: &dto.CacheEntry{CacheId: &dto.CacheId{CacheName: "c", Key: "1"}}, Found: true},
		nil,
	}
	router := NewRouter(adapter)
	body := bytes.NewBufferString(`{"requests":[{"c":"c","k":"1"},{"c":"c","k":"2"}]}`)
	req := httptest.NewRequest(http.MethodPost, getAllPath, body)
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
	req := httptest.NewRequest(http.MethodPost, putAllPath, body)
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
	req := httptest.NewRequest(http.MethodPost, evictAllPath, body)
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
	req := httptest.NewRequest(http.MethodPost, putAllPath, bytes.NewReader(big))
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
	req := httptest.NewRequest(http.MethodPost, putAllPath, &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", contentTypeJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestMetricsNoGzip(t *testing.T) {
	adapter := &mockAdapter{}
	router := NewRouter(adapter)
	req := httptest.NewRequest(http.MethodGet, metricsPath, nil)
	req.Header.Set("Accept-Encoding", encodingGzip)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Header().Get(headerContentEncoding) == encodingGzip {
		t.Fatalf("/metrics response should not be gzipped")
	}
}

func TestMetricsHealth(t *testing.T) {
	router := NewMetricRouter()
	req := httptest.NewRequest(http.MethodGet, metricsHealthPath, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != `{"status":"UP"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
