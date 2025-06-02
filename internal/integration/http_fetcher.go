// Package integration предоставляет компоненты для объединения результатов
// из внешнего HTTP API, поддерживает таймауты и пакетные запросы по ключам.
package integration

import (
	"aur-cache-service/internal/cache/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// httpBatchFetcher определяет интерфейс для HTTP-клиента,
// который выполняет пакетные запросы к внешнему API, получая
// данные по списку ключей и возвращая JSON-объекты по ключу.
type httpBatchFetcher interface {

	// GetAll отправляет POST-запрос с массивом ключей и возвращает
	// результат в виде отображения ключей на JSON-объекты.
	GetAll(ctx context.Context, keys []string, cfg *config.ApiBatchConfig) (map[string]*json.RawMessage, error)
}

// httpBatchFetcherImpl — реализация httpBatchFetcher,
// предназначенная для получения данных из внешнего API по ключам.
//
// Описание работы:
//   - Формирует тело запроса вида {prop: [key1, key2, ...]}.
//   - Поддерживает ключи типа string и number.
//   - Добавляет кастомные заголовки (если заданы в конфигурации).
//   - Отправляет POST-запрос по указанному URL.
//   - Возвращает результат в виде map[string]json.RawMessage, где
//     ключи соответствуют входным значениям.
//
// Пример тела запроса:
//
//	{ "id": ["123", "456"] }  // при keyType: string
//	{ "id": [123, 456] }      // при keyType: number
type httpBatchFetcherImpl struct {
	client *http.Client
}

func newHttpBatchFetcher(c *http.Client) *httpBatchFetcherImpl {
	return &httpBatchFetcherImpl{client: c}
}

const defaultTimeout = 15 * time.Second

func (f *httpBatchFetcherImpl) GetAll(ctx context.Context, keys []string, cfg *config.ApiBatchConfig) (map[string]*json.RawMessage, error) {

	if len(keys) == 0 {
		return map[string]*json.RawMessage{}, nil
	}

	bodyBytes, err := f.prepareBody(keys, cfg)
	if err != nil {
		return nil, err
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if cfg.URL == "" {
		return nil, fmt.Errorf("API URL is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response (%d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]*json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// prepareBody формирует JSON-тело запроса с учётом типа ключей (строковые или числовые).
func (f *httpBatchFetcherImpl) prepareBody(keys []string, cfg *config.ApiBatchConfig) (bodyBytes []byte, err error) {

	var payload map[string]interface{}

	switch cfg.KeyType {
	case config.KeyTypeNumber:
		converted := make([]json.Number, 0, len(keys))
		for _, k := range keys {
			if _, err := strconv.ParseFloat(k, 64); err != nil {
				return nil, fmt.Errorf("invalid numeric key %q: %w", k, err)
			}
			converted = append(converted, json.Number(k))
		}
		payload = map[string]interface{}{cfg.Prop: converted}

	case config.KeyTypeString:
		payload = map[string]interface{}{cfg.Prop: keys}

	default:
		return nil, fmt.Errorf("unsupported key type: %s", cfg.KeyType)
	}

	bodyBytes, err = json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}
	return
}
