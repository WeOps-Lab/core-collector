package scriptmetricreceiver

import (
	"fmt"
	"go.opentelemetry.io/collector/confmap"
	"time"
)

type Config struct {
	ScriptType         string        `mapstructure:"script_type"`
	ScriptContent      string        `mapstructure:"script_content"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
	ExecutionMode      string        `mapstructure:"execution_mode"`
	Protocol           string        `mapstructure:"protocol"`
	SSHUser            string        `mapstructure:"ssh_user"`
	SSHPassword        string        `mapstructure:"ssh_password"`
	SSHKeyPath         string        `mapstructure:"ssh_key_path"`
	Host               string        `mapstructure:"host"`
	Timeout            time.Duration `mapstructure:"timeout"`
	PythonInterpreter  string        `mapstructure:"python_interpreter"`
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	return componentParser.Unmarshal(cfg)
}

func (cfg *Config) Validate() error {
	switch cfg.ScriptType {
	case "shell", "bash", "python":
	default:
		return fmt.Errorf("invalid script type: %s", cfg.ScriptType)
	}

	if cfg.ScriptContent == "" {
		return fmt.Errorf("script content must not be empty")
	}

	switch cfg.ExecutionMode {
	case "local", "remote":
	default:
		return fmt.Errorf("invalid execution mode: %s", cfg.ExecutionMode)
	}

	if cfg.ExecutionMode == "remote" {
		if cfg.Protocol == "ssh" {
			if cfg.Host == "" {
				return fmt.Errorf("host must be specified for remote execution")
			}
			if cfg.SSHUser == "" {
				return fmt.Errorf("SSH user must be specified for remote execution")
			}
			if cfg.SSHPassword == "" && cfg.SSHKeyPath == "" {
				return fmt.Errorf("either SSH password or SSH key path must be provided for remote execution")
			}
		} else {
			return fmt.Errorf("invalid protocol: %s", cfg.Protocol)
		}
	}

	if cfg.CollectionInterval <= 0 {
		return fmt.Errorf("collection interval must be greater than 0")
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	return nil
}
