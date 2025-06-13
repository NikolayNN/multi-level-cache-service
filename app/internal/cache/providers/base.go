package providers

import (
	"context"
	"time"
)

// CacheProvider — интерфейс для многоключевого кэш-хранилища с поддержкой TTL и отмены через контекст.
type CacheProvider interface {

	// BatchGet возвращает значения по ключам. Не найденые ключи игнорируются
	BatchGet(ctx context.Context, keys []string) (map[string]string, error)

	// BatchPut сохраняет ключи со значениями и соответствующими TTL.
	BatchPut(ctx context.Context, items map[string]string, ttls map[string]time.Duration) error

	// BatchDelete удаляет указанные ключи.
	BatchDelete(ctx context.Context, keys []string) error

	// Close освобождает ресурсы.
	Close() error
}

// calcChunkSize вычисляет оптимальный размер chunk'а для равномерного распределения элементов.
//
// Параметры:
//   - givenSize: общее количество элементов для разбиения
//   - minSize: минимально допустимый размер chunk'а
//   - maxSize: максимально допустимый размер chunk'а
//
// Возвращает оптимальный размер chunk'а в пределах [minSize, maxSize].
func calcChunkSize(givenSize int, minSize int, maxSize int) int {
	if givenSize <= maxSize {
		return givenSize
	}

	optimalChunks := (givenSize + maxSize - 1) / maxSize

	optimalSize := (givenSize + optimalChunks - 1) / optimalChunks

	if optimalSize < minSize {
		return minSize
	}
	if optimalSize > maxSize {
		return maxSize
	}

	return optimalSize
}

// splitToChunks разбивает map на несколько chunk'ов оптимального размера.
//
// Параметры:
//   - keyValues: исходная map для разбиения
//   - minSize: минимально допустимый размер chunk'а
//   - maxSize: максимально допустимый размер chunk'а
//
// Возвращает slice указателей на map'ы, представляющие chunk'и.
// Каждый chunk содержит примерно одинаковое количество элементов
// в пределах заданных ограничений.
//
// Особенности:
//   - Порядок элементов в chunk'ах не гарантирован (особенность Go map)
//   - Последний chunk может быть меньше остальных
//   - Для пустой map возвращается пустой slice
func splitKeyValueToChunks(keyValues map[string]string, minSize int, maxSize int) []map[string]string {
	if len(keyValues) == 0 {
		return []map[string]string{}
	}

	chunkSize := calcChunkSize(len(keyValues), minSize, maxSize)
	expectedChunks := (len(keyValues) + chunkSize - 1) / chunkSize

	chunks := make([]map[string]string, 0, expectedChunks)
	currentChunk := make(map[string]string)
	count := 0

	for key, value := range keyValues {
		currentChunk[key] = value
		count++

		if count == chunkSize {
			chunks = append(chunks, currentChunk)
			currentChunk = make(map[string]string)
			count = 0
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// splitKeysToChunks разбивает slice строк на несколько chunk'ов оптимального размера.
//
// Параметры:
//   - keys: исходный slice строк для разбиения
//   - minSize: минимально допустимый размер chunk'а
//   - maxSize: максимально допустимый размер chunk'а
//
// Возвращает slice из slice'ов строк, представляющих chunk'и.
// Каждый chunk содержит примерно одинаковое количество элементов
// в пределах заданных ограничений.
func splitKeysToChunks(keys []string, minSize int, maxSize int) [][]string {
	if len(keys) == 0 {
		return [][]string{}
	}

	chunkSize := calcChunkSize(len(keys), minSize, maxSize)
	expectedChunks := (len(keys) + chunkSize - 1) / chunkSize

	chunks := make([][]string, 0, expectedChunks)

	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := keys[i:end]
		chunks = append(chunks, chunk)
	}

	return chunks
}
