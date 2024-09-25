module github.com/open-telemetry/opentelemetry-collector-contrib/extension/tokenauthextension

go 1.22.0

require (
	github.com/jellydator/ttlcache/v3 v3.3.0
	github.com/redis/go-redis/v9 v9.6.1
	go.opentelemetry.io/collector/component v0.109.0
	go.opentelemetry.io/collector/extension v0.109.0
	go.opentelemetry.io/collector/extension/auth v0.109.0
	go.uber.org/zap v1.27.0
)