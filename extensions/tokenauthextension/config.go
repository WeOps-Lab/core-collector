package tokenauthextension

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	RedisAddress  string        `mapstructure:"redis_address,omitempty"`
	RedisPassword string        `mapstructure:"redis_password,omitempty"`
	RedisDB       int           `mapstructure:"redis_db,omitempty"`
	CacheTTL      time.Duration `mapstructure:"cache_ttl,omitempty"`
}

var _ component.Config = (*Config)(nil)

var errNoRedisAddressProvided = errors.New("no redis address provided")

func (cfg *Config) Validate() error {
	if cfg.RedisAddress == "" {
		return errNoRedisAddressProvided
	}
	return nil
}
