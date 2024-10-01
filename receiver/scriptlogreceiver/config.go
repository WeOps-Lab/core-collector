package scriptlogreceiver

import (
	"fmt"
	"go.opentelemetry.io/collector/confmap"
	"time"
)

type Config struct {
	ScriptType         string        `mapstructure:"script_type"`
	ScriptContent      string        `mapstructure:"script_content"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
	ExecutionMode      string        `mapstructure:"execution_mode"` // local or remote
	Protocol           string        `mapstructure:"protocol"`       // ssh
	SSHUser            string        `mapstructure:"ssh_user"`
	SSHPassword        string        `mapstructure:"ssh_password"`
	SSHKeyPath         string        `mapstructure:"ssh_key_path"`
	Host               string        `mapstructure:"host"`
	Timeout            time.Duration `mapstructure:"timeout"` // script execution timeout
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	return componentParser.Unmarshal(cfg)
}

func (cfg *Config) Validate() error {
	// Ensure ScriptType is valid
	switch cfg.ScriptType {
	case "bash", "python":
		// valid script types
	default:
		return fmt.Errorf("invalid script type: %s", cfg.ScriptType)
	}

	// Ensure ScriptContent is provided
	if cfg.ScriptContent == "" {
		return fmt.Errorf("script content must not be empty")
	}

	// Ensure ExecutionMode is valid
	switch cfg.ExecutionMode {
	case "local", "remote":
		// valid execution modes
	default:
		return fmt.Errorf("invalid execution mode: %s", cfg.ExecutionMode)
	}

	// If remote execution, ensure Protocol, Host, and authentication are provided
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

	// Ensure CollectionInterval and Timeout are valid
	if cfg.CollectionInterval <= 0 {
		return fmt.Errorf("collection interval must be greater than 0")
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	if cfg.Timeout >= cfg.CollectionInterval {
		return fmt.Errorf("timeout must be smaller than collection interval")
	}
	return nil
}
