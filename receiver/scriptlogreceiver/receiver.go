package scriptlogreceiver

import (
	"context"
	"fmt"
	"github.com/melbahja/goph"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
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
	var output []byte
	var err error

	// 设置上下文，用于脚本执行超时
	ctx, cancel := context.WithTimeout(context.Background(), slr.config.Timeout)
	defer cancel()

	switch slr.config.ExecutionMode {
	case "local":
		output, err = slr.executeLocalScript(ctx)
	case "remote":
		output, err = slr.executeRemoteScript(ctx)
	default:
		slr.params.Logger.Error("unsupported execution mode", zap.String("mode", slr.config.ExecutionMode))
		return
	}

	if err != nil {
		slr.params.Logger.Error("error executing script", zap.Error(err))
		return
	}

	// Placeholder for processing output and sending to nextConsumer
	slr.params.Logger.Debug("script output", zap.String("output", string(output)))

	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	ils := rls.ScopeLogs().AppendEmpty()
	logRecord := ils.LogRecords().AppendEmpty()
	logRecord.Body().SetStr(string(output))
	slr.nextConsumer.ConsumeLogs(context.Background(), logs)
}

func (slr *scriptLogReceiver) executeLocalScript(ctx context.Context) ([]byte, error) {
	var cmd *exec.Cmd

	switch slr.config.ScriptType {
	case "bash":
		cmd = exec.CommandContext(ctx, "sh", "-c", slr.config.ScriptContent)
	case "python":
		cmd = exec.CommandContext(ctx, "python", "-c", slr.config.ScriptContent)
	default:
		return nil, fmt.Errorf("unsupported script type")
	}

	return cmd.Output()
}

func (slr *scriptLogReceiver) executeRemoteScript(ctx context.Context) ([]byte, error) {
	var auth goph.Auth
	var err error

	if slr.config.SSHKeyPath != "" {
		// 使用私钥进行认证
		auth, err = goph.Key(slr.config.SSHKeyPath, "")
		if err != nil {
			return nil, err
		}
	} else {
		// 使用密码进行认证
		auth = goph.Password(slr.config.SSHPassword)
	}

	// 创建 SSH 客户端
	client, err := goph.New(slr.config.SSHUser, slr.config.Host, auth)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// 使用上下文来控制指令执行的超时
	ctx, cancel := context.WithTimeout(context.Background(), slr.config.Timeout)
	defer cancel()

	// 执行远程脚本
	out, err := client.RunContext(ctx, slr.config.ScriptContent)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (slr *scriptLogReceiver) Shutdown(ctx context.Context) error {
	if slr.cancel != nil {
		slr.cancel()
	}
	return nil
}
