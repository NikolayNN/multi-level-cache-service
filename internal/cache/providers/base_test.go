package providers

import (
	"fmt"
	"testing"
)

func TestCalcChunkSize(t *testing.T) {
	// Простые базовые тесты
	if result := calcChunkSize(300, 100, 500); result != 300 {
		t.Errorf("Expected 300, got %d", result)
	}

	if result := calcChunkSize(1006, 100, 500); result != 336 {
		t.Errorf("Expected 336, got %d", result)
	}

	if result := calcChunkSize(1000, 100, 500); result != 500 {
		t.Errorf("Expected 500, got %d", result)
	}

	if result := calcChunkSize(0, 100, 500); result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

func TestSplitToChunks(t *testing.T) {
	// Тест с пустой map
	empty := make(map[string]string)
	result := splitToChunks(empty, 10, 100)
	if len(result) != 0 {
		t.Errorf("Expected 0 chunks for empty map, got %d", len(result))
	}

	// Тест с одним элементом
	single := map[string]string{"key1": "value1"}
	result = splitToChunks(single, 1, 10)
	if len(result) != 1 || len(*result[0]) != 1 {
		t.Errorf("Expected 1 chunk with 1 element")
	}

	// Тест с точным делением
	data := createTestMap(10)
	result = splitToChunks(data, 2, 5)
	if len(result) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(result))
	}

	// Проверяем общее количество элементов
	total := 0
	for _, chunk := range result {
		total += len(*chunk)
	}
	if total != 10 {
		t.Errorf("Expected total 10 elements, got %d", total)
	}
}

func TestChunkIntegrity(t *testing.T) {
	// Проверяем, что все данные сохраняются
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
	}

	result := splitToChunks(input, 1, 2)

	// Собираем все обратно
	reconstructed := make(map[string]string)
	for _, chunk := range result {
		for k, v := range *chunk {
			reconstructed[k] = v
		}
	}

	// Проверяем совпадение
	if len(reconstructed) != len(input) {
		t.Errorf("Lost data: input %d, reconstructed %d", len(input), len(reconstructed))
	}

	for k, v := range input {
		if reconstructed[k] != v {
			t.Errorf("Data mismatch for key %s", k)
		}
	}
}

// Вспомогательная функция для создания тестовых данных
func createTestMap(size int) map[string]string {
	result := make(map[string]string)
	for i := 0; i < size; i++ {
		result[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	return result
}
