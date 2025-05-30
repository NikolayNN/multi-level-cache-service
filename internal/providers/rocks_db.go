package providers

import (
	"aur-cache-service/internal/config"
	"fmt"
	"strconv"
	"time"

	"github.com/linxGnu/grocksdb"
)

type RocksDb struct {
	db        *grocksdb.DB
	readOpts  *grocksdb.ReadOptions
	writeOpts *grocksdb.WriteOptions
	ttlCache  map[string]time.Time // Простой кэш TTL в памяти
}

// NewRocksDb создает новый клиент RocksDB
func NewRocksDb(cfg config.RocksDB) (*RocksDb, error) {
	// Настройка опций базы данных
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	if cfg.BlockSizeBytes() > 0 {
		bbto.SetBlockSize(int(cfg.BlockSizeBytes()))
	}
	if cfg.BlockCacheBytes() > 0 {
		bbto.SetBlockCache(grocksdb.NewLRUCache(cfg.BlockCacheBytes()))
	}

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(cfg.CreateIfMissing)

	if cfg.MaxOpenFiles > 0 {
		opts.SetMaxOpenFiles(cfg.MaxOpenFiles)
	}
	if cfg.WriteBufferSizeBytes() > 0 {
		opts.SetWriteBufferSize(cfg.WriteBufferSizeBytes())
	}

	// Открытие базы данных
	db, err := grocksdb.OpenDb(opts, cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть RocksDB: %w", err)
	}

	// Создание объекта клиента
	return &RocksDb{
		db:        db,
		readOpts:  grocksdb.NewDefaultReadOptions(),
		writeOpts: grocksdb.NewDefaultWriteOptions(),
		ttlCache:  make(map[string]time.Time),
	}, nil
}

// BatchGet получает несколько значений за один запрос
func (c *RocksDb) BatchGet(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	result := make(map[string]string)

	for _, key := range keys {
		// Проверка TTL
		if c.isExpired(key) {
			c.Delete(key)
			continue
		}

		// Получение значения
		slice, err := c.db.Get(c.readOpts, []byte(key))
		if err != nil {
			return nil, fmt.Errorf("ошибка получения значения из RocksDB: %w", err)
		}

		if slice.Exists() {
			result[key] = string(slice.Data())
		}
		slice.Free()
	}

	return result, nil
}

// BatchPut сохраняет несколько значений за один запрос
func (c *RocksDb) BatchPut(items map[string]string, ttls map[string]int64) error {
	if len(items) == 0 {
		return nil
	}

	// Создаем пакет операций
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	// Для всех ключей с TTL
	ttlBatch := grocksdb.NewWriteBatch()
	defer ttlBatch.Destroy()

	for key, value := range items {
		// Добавление операции записи в пакет
		batch.Put([]byte(key), []byte(value))

		// Обработка TTL, если он указан
		if ttl, exists := ttls[key]; exists && ttl > 0 {
			expiration := time.Now().Add(time.Duration(ttl) * time.Second)
			c.ttlCache[key] = expiration

			// Сохраняем TTL в базу данных
			ttlKey := "ttl:" + key
			ttlValue := strconv.FormatInt(expiration.UnixNano(), 10)
			ttlBatch.Put([]byte(ttlKey), []byte(ttlValue))
		}
	}

	// Выполнение всех операций записи за один вызов
	err := c.db.Write(c.writeOpts, batch)
	if err != nil {
		return fmt.Errorf("ошибка пакетного сохранения в RocksDB: %w", err)
	}

	// Выполнение записи TTL
	if ttlBatch.Count() > 0 {
		err = c.db.Write(c.writeOpts, ttlBatch)
		if err != nil {
			return fmt.Errorf("ошибка сохранения TTL в RocksDB: %w", err)
		}
	}

	return nil
}

// BatchDelete удаляет несколько значений за один запрос
func (c *RocksDb) BatchDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Создаем пакет операций
	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	// Удаление TTL для всех ключей
	ttlBatch := grocksdb.NewWriteBatch()
	defer ttlBatch.Destroy()

	for _, key := range keys {
		// Добавление операции удаления в пакет
		batch.Delete([]byte(key))

		// Удаление TTL, если он был
		delete(c.ttlCache, key)
		ttlKey := "ttl:" + key
		ttlBatch.Delete([]byte(ttlKey))
	}

	// Выполнение всех операций удаления за один вызов
	err := c.db.Write(c.writeOpts, batch)
	if err != nil {
		return fmt.Errorf("ошибка пакетного удаления из RocksDB: %w", err)
	}

	// Выполнение удаления TTL
	if ttlBatch.Count() > 0 {
		err = c.db.Write(c.writeOpts, ttlBatch)
		if err != nil {
			return fmt.Errorf("ошибка удаления TTL из RocksDB: %w", err)
		}
	}

	return nil
}

// Close закрывает соединение с базой данных
func (c *RocksDb) Close() error {
	c.readOpts.Destroy()
	c.writeOpts.Destroy()
	c.db.Close()
	return nil
}

// isExpired проверяет, истек ли срок действия ключа
func (c *RocksDb) isExpired(key string) bool {
	// Проверка в кэше в памяти
	if expTime, exists := c.ttlCache[key]; exists {
		if time.Now().After(expTime) {
			return true
		}
		return false
	}

	// Если нет в кэше, проверяем в базе данных
	ttlKey := "ttl:" + key
	slice, err := c.db.Get(c.readOpts, []byte(ttlKey))
	if err != nil || !slice.Exists() {
		if slice != nil {
			slice.Free()
		}
		return false
	}
	defer slice.Free()

	// Преобразуем строку в время
	expNano, err := strconv.ParseInt(string(slice.Data()), 10, 64)
	if err != nil {
		return false
	}

	expTime := time.Unix(0, expNano)

	// Кэшируем для будущих запросов
	c.ttlCache[key] = expTime

	// Проверяем истечение срока
	return time.Now().After(expTime)
}
