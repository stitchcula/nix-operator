package udev

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type LinuxUdevHandler struct{}

func (h *LinuxUdevHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	var content strings.Builder
	content.WriteString(config.CommentHeader)

	for _, rule := range cfg.Spec.Udev.Rules {
		content.WriteString(fmt.Sprintf("SUBSYSTEM==\"%s\", ", rule.Subsystem))
		for key, value := range rule.Attrs {
			content.WriteString(fmt.Sprintf("ATTRS{%s}==\"%s\", ", key, value))
		}
		content.WriteString(fmt.Sprintf("SYMLINK+=\"%s\"\n", rule.Symlink))
	}

	// 使用工具函数原子性写入文件
	if err := utils.AtomicWriteFile([]byte(content.String()), "/etc/udev/rules.d/99-custom.rules", 0644); err != nil {
		return fmt.Errorf("failed to write udev rules: %v", err)
	}

	// 重新加载 udev 规则
	cmd := exec.CommandContext(ctx, "udevadm", "control", "--reload-rules")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload udev rules: %v, output: %s", err, output)
	}

	// 触发 udev 规则
	cmd = exec.CommandContext(ctx, "udevadm", "trigger")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to trigger udev rules: %v, output: %s", err, output)
	}

	return nil
}

type LinuxUdevFactory struct{}

func (f *LinuxUdevFactory) Create() controller.Handler {
	return &LinuxUdevHandler{}
}

func (f *LinuxUdevFactory) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func init() {
	controller.RegisterHandler("udev", &LinuxUdevFactory{})
}
