package scriptmetricreceiver

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/melbahja/goph"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

func newScriptMetricReceiver(params receiver.Settings, cfg *Config, nextConsumer consumer.Metrics) *scriptMetricReceiver {
	return &scriptMetricReceiver{
		params:       params,
		config:       cfg,
		nextConsumer: nextConsumer,
	}
}

type scriptMetricReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	params       receiver.Settings

	cancel context.CancelFunc
}

func (smr *scriptMetricReceiver) Start(ctx context.Context, host component.Host) error {
	ctx, cancel := context.WithCancel(ctx)
	smr.cancel = cancel

	go smr.start(ctx)
	return nil
}

func (smr *scriptMetricReceiver) start(ctx context.Context) {
	ticker := time.NewTicker(smr.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			smr.collectMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}
}
func (smr *scriptMetricReceiver) collectMetrics(ctx context.Context) {
	var output []byte
	var err error

	scriptCtx, cancel := context.WithTimeout(ctx, smr.config.Timeout)
	defer cancel()

	switch smr.config.ExecutionMode {
	case "local":
		output, err = smr.executeLocalScript(scriptCtx)
	case "remote":
		output, err = smr.executeRemoteScript(scriptCtx)
	default:
		smr.params.Logger.Error("unsupported execution mode", zap.String("mode", smr.config.ExecutionMode))
		return
	}

	if err != nil {
		smr.params.Logger.Error("error executing script", zap.Error(err))
		return
	}

	smr.params.Logger.Debug("script output", zap.String("output", string(output)))

	metrics := pmetric.NewMetrics()
	rms := metrics.ResourceMetrics().AppendEmpty()
	ils := rms.ScopeMetrics().AppendEmpty()
	now := pcommon.NewTimestampFromTime(time.Now())

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()

		// Split the string by colon first
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			smr.params.Logger.Error("invalid metric line", zap.String("line", line))
			continue
		}

		// Metric Name and Other Values
		metricName := strings.TrimSpace(parts[0])
		metricVals := strings.Fields(strings.TrimSpace(parts[1]))
		if len(metricVals) < 1 {
			smr.params.Logger.Error("invalid metric values", zap.String("line", line))
			continue
		}

		metricValue, err := strconv.ParseFloat(metricVals[0], 64)
		if err != nil {
			smr.params.Logger.Error("invalid metric value", zap.String("value", metricVals[0]), zap.Error(err))
			continue
		}

		metric := ils.Metrics().AppendEmpty()
		metric.SetName(metricName)
		metric.SetUnit("")

		gauge := metric.SetEmptyGauge()
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetDoubleValue(metricValue)
		dp.SetTimestamp(now)

		for i := 1; i < len(metricVals); i++ {
			labelParts := strings.SplitN(metricVals[i], "=", 2)
			if len(labelParts) == 2 {
				dp.Attributes().PutStr(labelParts[0], labelParts[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		smr.params.Logger.Error("error reading script output", zap.Error(err))
		return
	}

	err = smr.nextConsumer.ConsumeMetrics(ctx, metrics)
	if err != nil {
		smr.params.Logger.Error("failed to consume metrics", zap.Error(err))
	}
}

func (smr *scriptMetricReceiver) executeLocalScript(ctx context.Context) ([]byte, error) {
	var cmd *exec.Cmd

	switch smr.config.ScriptType {
	case "bash":
		cmd = exec.CommandContext(ctx, "sh", "-c", smr.config.ScriptContent)
	case "python":
		cmd = exec.CommandContext(ctx, smr.config.PythonInterpreter, "-c", smr.config.ScriptContent)
	default:
		return nil, fmt.Errorf("unsupported script type")
	}

	return cmd.Output()
}

func (smr *scriptMetricReceiver) executeRemoteScript(ctx context.Context) ([]byte, error) {
	var auth goph.Auth
	var err error

	if smr.config.SSHKeyPath != "" {
		auth, err = goph.Key(smr.config.SSHKeyPath, "")
		if err != nil {
			return nil, err
		}
	} else {
		auth = goph.Password(smr.config.SSHPassword)
	}

	client, err := goph.New(smr.config.SSHUser, smr.config.Host, auth)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	out, err := client.RunContext(ctx, smr.config.ScriptContent)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (smr *scriptMetricReceiver) Shutdown(ctx context.Context) error {
	if smr.cancel != nil {
		smr.cancel()
	}
	return nil
}
