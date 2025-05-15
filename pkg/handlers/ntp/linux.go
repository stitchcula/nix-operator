package ntp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type LinuxNTPHandler struct{}

func (h *LinuxNTPHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	if !cfg.Spec.System.NTP.Enabled {
		cmd := exec.CommandContext(ctx, "systemctl", "stop", "chronyd")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to stop chronyd: %v, output: %s", err, output)
		}
		return nil
	}

	// 生成 chrony.conf
	var content strings.Builder
	content.WriteString(config.CommentHeader)
	for _, server := range cfg.Spec.System.NTP.Servers {
		content.WriteString(fmt.Sprintf("server %s iburst\n", server))
	}

	// 使用工具函数原子性写入文件
	if err := utils.AtomicWriteFile([]byte(content.String()), "/etc/chrony.conf", 0644); err != nil {
		return fmt.Errorf("failed to write chrony.conf: %v", err)
	}

	// 重启服务
	cmd := exec.CommandContext(ctx, "systemctl", "restart", "chronyd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart chronyd: %v, output: %s", err, output)
	}

	return nil
}

type LinuxNTPFactory struct{}

func (f *LinuxNTPFactory) Create() controller.Handler {
	return &LinuxNTPHandler{}
}

func (f *LinuxNTPFactory) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func init() {
	controller.RegisterHandler("ntp", &LinuxNTPFactory{})
}
