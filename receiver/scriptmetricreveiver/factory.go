package scriptmetricreceiver

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
var defaultPythonInterpreter = "python"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("scriptmetric"),
		createDefaultConfig,
		receiver.WithMetrics(createMetricReceiver, component.StabilityLevelBeta))
}

func createDefaultConfig() component.Config {
	return &Config{
		ScriptType:         defaultScriptType,
		CollectionInterval: defaultCollectionInterval,
		Timeout:            defaultTimeout,
		ExecutionMode:      defaultExecutionMode,
		PythonInterpreter:  defaultPythonInterpreter,
	}
}

func createMetricReceiver(
	_ context.Context,
	params receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	mCfg := cfg.(*Config)
	return newScriptMetricReceiver(params, mCfg, nextConsumer), nil
}
