package scriptlogreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"time"
)

var defaultScriptType = "bash"
var defaultCollectionInterval = 10 * time.Second
var defaultTimeout = 5 * time.Second
var defaultExecutionMode = "local"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("scriptlog"),
		createDefaultConfig,
		receiver.WithLogs(createLogReceiver, component.StabilityLevelBeta))
}

func createDefaultConfig() component.Config {
	return &Config{
		ScriptType:         defaultScriptType,
		CollectionInterval: defaultCollectionInterval,
		Timeout:            defaultTimeout,
		ExecutionMode:      defaultExecutionMode,
	}
}

func createLogReceiver(
	_ context.Context,
	params receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (receiver.Logs, error) {
	rCfg := cfg.(*Config)
	return newScriptLogReciever(params, rCfg, nextConsumer), nil
}
