package script_executer

import (
	"context"
	"fmt"
	"github.com/melbahja/goph"
	"os/exec"
	"time"
)

func ExecuteLocalScript(ctx context.Context, scriptType string, scriptContent string) ([]byte, error) {
	var cmd *exec.Cmd

	switch scriptType {
	case "bash":
		cmd = exec.CommandContext(ctx, "sh", "-c", scriptContent)
	case "python":
		cmd = exec.CommandContext(ctx, "python", "-c", scriptContent)
	default:
		return nil, fmt.Errorf("unsupported script type")
	}

	return cmd.Output()
}

func ExecuteRemoteScript(ctx context.Context, sshUser, host, sshKeyPath,
	sshPassword, scriptContent string, timeout time.Duration) ([]byte, error) {
	var auth goph.Auth
	var err error

	if sshKeyPath != "" {
		// 使用私钥进行认证
		auth, err = goph.Key(sshKeyPath, "")
		if err != nil {
			return nil, err
		}
	} else {
		// 使用密码进行认证
		auth = goph.Password(sshPassword)
	}

	// 创建 SSH 客户端
	client, err := goph.New(sshUser, host, auth)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// 使用上下文来控制指令执行的超时
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行远程脚本
	out, err := client.RunContext(ctx, scriptContent)
	if err != nil {
		return nil, err
	}

	return out, nil
}
