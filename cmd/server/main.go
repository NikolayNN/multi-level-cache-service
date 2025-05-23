package main

import (
	"aur-cache-service/internal/clients/redisClient"
	"aur-cache-service/internal/clients/ristrettoClient"
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
	ristrettoConfig := ristrettoClient.Config{
		NumCounters: 10000,
		BufferItems: 64,
		MaxCost:     100 * 1000,
	}

	client, err := ristrettoClient.New(ristrettoConfig)
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
	redisConfig := redisClient.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
		Timeout:  5 * time.Second,
	}
	client, err := redisClient.New(redisConfig)
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
