package scriptlogreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func newScriptLogReciever(params receiver.Settings, cfg *Config, nextConsumer consumer.Logs) *scriptLogReceiver {
	return &scriptLogReceiver{
		params:       params,
		config:       cfg,
		nextConsumer: nextConsumer,
	}
}

type scriptLogReceiver struct {
	config       *Config
	nextConsumer consumer.Logs
	params       receiver.Settings

	cancel context.CancelFunc
}

func (slr *scriptLogReceiver) Start(ctx context.Context, host component.Host) error {
	ctx, cancel := context.WithCancel(ctx)
	slr.cancel = cancel

	go slr.start(ctx)

	return nil
}

func (slr *scriptLogReceiver) start(ctx context.Context) {
	ticker := time.NewTicker(slr.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slr.executeScript()
		case <-ctx.Done():
			return
		}
	}
}

func (slr *scriptLogReceiver) executeScript() {
	var cmd *exec.Cmd

	switch slr.config.ScriptType {
	case "shell":
		cmd = exec.Command("sh", "-c", slr.config.ScriptContent)
	case "python":
		cmd = exec.Command("python", "-c", slr.config.ScriptContent)
	default:
		slr.params.Logger.Error("unsupported script type", zap.String("type", slr.config.ScriptType))
		return
	}

	output, err := cmd.Output()
	if err != nil {
		slr.params.Logger.Error("error executing script", zap.Error(err))
		return
	}

	// Placeholder for processing output and sending to nextConsumer
	slr.params.Logger.Info("script output", zap.String("output", string(output)))
}

func (slr *scriptLogReceiver) Shutdown(ctx context.Context) error {
	if slr.cancel != nil {
		slr.cancel()
	}
	return nil
}
