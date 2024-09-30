package scriptlogreceiver

import (
	"go.opentelemetry.io/collector/confmap"
	"time"
)

type Config struct {
	ScriptType         string        `mapstructure:"script_type"`
	ScriptContent      string        `mapstructure:"script_content"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	return componentParser.Unmarshal(cfg)
}
