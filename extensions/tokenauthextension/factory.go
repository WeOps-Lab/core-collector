package tokenauthextension

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

const (
	defaultRedisPassword = ""
	defaultCacheTTL      = 30 * time.Minute
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType("tokenauth"),
		createDefaultConfig,
		createExtension,
		component.StabilityLevelBeta,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		RedisPassword: defaultRedisPassword,
		CacheTTL:      defaultCacheTTL,
		RedisDB:       0,
	}
}

func createExtension(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	config := cfg.(*Config)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return NewTokenAuth(config, set.Logger), nil
}
