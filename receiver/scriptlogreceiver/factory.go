package scriptlogreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"time"
)

var defaultScriptType = "shell"
var defaultCollectionInterval = 10 * time.Second

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("scriptlog"),
		createDefaultConfig,
		receiver.WithMetrics(createLogReceiver, component.StabilityLevelBeta))
}

func createDefaultConfig() component.Config {
	return &Config{
		ScriptType:         defaultScriptType,
		CollectionInterval: defaultCollectionInterval,
	}
}

func createLogReceiver(
	_ context.Context,
	params receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	rCfg := cfg.(*Config)
	return newScriptLogReciever(params, rCfg, nextConsumer), nil
}
