module aur-cache-service

go 1.24.1

require (
	github.com/dgraph-io/ristretto v0.2.0
	github.com/dustin/go-humanize v1.0.1
	github.com/linxGnu/grocksdb v1.10.1
	github.com/prometheus/client_golang v1.22.0
	github.com/redis/go-redis/v9 v9.8.0
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v3 v3.0.1
	telegram-alerts-go v0.0.0-20250616092414-fb9fa9ae520e
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

require (
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-chi/chi/v5 v5.2.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace telegram-alerts-go => github.com/NikolayNN/telegram-alerts-go v0.0.0-20250616092414-fb9fa9ae520e
