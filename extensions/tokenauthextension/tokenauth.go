package tokenauthextension

import (
	"context"
	"errors"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/auth"
	"go.uber.org/zap"
)

var (
	_ auth.Server = (*TokenAuth)(nil)
)

type TokenAuth struct {
	scheme                   string
	authorizationValueAtomic atomic.Value
	shutdownCH               chan struct{}
	logger                   *zap.Logger
	redisClient              *redis.Client
	cache                    *ttlcache.Cache[string, string]
}

func NewTokenAuth(cfg *Config, logger *zap.Logger) *TokenAuth {
	if cfg.RedisAddress == "" {
		logger.Warn("no redis address provided")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](cfg.CacheTTL),
	)

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		logger.Error("failed to connect to redis", zap.Error(err))
	}

	auth := &TokenAuth{
		logger:      logger,
		redisClient: redisClient,
		cache:       cache,
		shutdownCH:  make(chan struct{}),
	}

	go cache.Start()
	return auth
}

func (t *TokenAuth) Start(ctx context.Context, host component.Host) error {
	t.logger.Info("starting token auth extension...")
	return nil
}

func (t *TokenAuth) Shutdown(ctx context.Context) error {
	t.logger.Info("shutting down token auth extension, closing redis client...")
	t.redisClient.Close()
	close(t.shutdownCH)
	return nil
}

func (t *TokenAuth) Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	authHeader, ok := headers["authorization"]
	if !ok {
		authHeader, ok = headers["Authorization"]
	}
	if !ok || len(authHeader) == 0 {
		return ctx, errors.New("missing or empty authorization header")
	}

	token := extractToken(authHeader[0])
	if token == "" {
		return ctx, errors.New("invalid authorization header format")
	}

	if t.cache.Has(token) {
		t.logger.Debug("token found in memory cache", zap.String("token", token))
		return ctx, nil
	}

	redisToken, err := t.redisClient.Get(context.Background(), token).Result()
	if err != nil {
		return ctx, fmt.Errorf("token not found: %s", token)
	}

	t.logger.Debug("token found in redis cache, setting to local cache with default TTL", zap.String("token", token))
	t.cache.Set(token, redisToken, ttlcache.DefaultTTL)
	return ctx, nil
}

func extractToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
