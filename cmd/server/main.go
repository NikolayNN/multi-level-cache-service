package main

import (
	config2 "aur-cache-service/internal/cache/config"
	providers2 "aur-cache-service/internal/cache/providers"
	"aur-cache-service/internal/config"
	"fmt"
	"log"
	"time"
)

func main() {
	testLoadCacheConfig()
	//testRocksDb()
	// testRedis()
	//testRistretto()
}

func testLoadCacheConfig() {
	configs, err := config.LoadCacheStorage("configs/cache.yml")
	if err != nil {
		fmt.Println("Error read config")
	}
	fmt.Printf("%+v\n", configs)

}

func testRistretto() {
	ristrettoConfig := config2.Ristretto{
		NumCounters: 10000,
		BufferItems: 64,
		MaxCost:     "64MiB",
	}

	client, err := providers2.NewRistretto(ristrettoConfig)
	if err != nil {
		log.Fatalf("error init Ristretto: %v", err)
	}
	err = client.Put("rst:123", "ristrettoValue", 65000)
	if err != nil {
		log.Fatalf("error put value %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	v, _, _ := client.Get("rst:123")
	fmt.Println(v)
}

func testRedis() {
	redisConfig := config2.Redis{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  5 * time.Second,
	}
	client, err := providers2.NewRedis(redisConfig)
	if err != nil {
		log.Fatalf("error init Redis: %v", err)
	}

	err = client.Put("u:123", "value123", 600)
	if err != nil {
		log.Fatalf("error put value %v", err)
	}

	v, _, _ := client.Get("u:123")
	fmt.Println(v)
}
