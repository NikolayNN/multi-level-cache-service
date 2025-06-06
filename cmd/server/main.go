package main

import (
	"aur-cache-service/internal/cache/config"
	"fmt"
)

func main() {

}

func testLoadCacheConfig() {
	configs, err := config.LoadAppConfig("configs/cache.yml")
	if err != nil {
		fmt.Println("Error read config")
	}
	fmt.Printf("%+v\n", configs)
}
