package udev

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type Config struct {
	Rules []UdevRule `json:"rules"`
}

type UdevRule struct {
	Name      string            `json:"name"`
	Subsystem string            `json:"subsystem"`
	Attrs     map[string]string `json:"attrs"`
	Symlink   string            `json:"symlink"`
}

func init() {
	controller.RegisterHandler("UdevConfiguration", &LinuxUdevHandler{})
}

func (f *LinuxUdevHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

type LinuxUdevHandler struct{}

func (h *LinuxUdevHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	// 生成期望的 udev 规则内容
	desiredContent := h.generateUdevRules(cfg.Spec.Udev.Rules)

	// 读取现有的 udev 规则文件
	currentContent, err := h.getCurrentUdevRules()
	if err != nil {
		return fmt.Errorf("failed to read current udev rules: %v", err)
	}

	// 比较现有配置和期望配置
	if currentContent == desiredContent {
		return nil // 配置一致，无需更新
	}

	// 使用工具函数原子性写入文件
	if err := utils.AtomicWriteFile([]byte(desiredContent), "/etc/udev/rules.d/99-nix-operator.rules", 0644); err != nil {
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

func (h *LinuxUdevHandler) generateUdevRules(rules []config.UdevRule) string {
	var content strings.Builder
	content.WriteString(config.CommentHeader)

	for _, rule := range rules {
		content.WriteString(fmt.Sprintf("SUBSYSTEM==\"%s\", ", rule.Subsystem))
		for key, value := range rule.Attrs {
			content.WriteString(fmt.Sprintf("ATTRS{%s}==\"%s\", ", key, value))
		}
		content.WriteString(fmt.Sprintf("SYMLINK+=\"%s\"\n", rule.Symlink))
	}

	return content.String()
}

func (h *LinuxUdevHandler) getCurrentUdevRules() (string, error) {
	data, err := os.ReadFile("/etc/udev/rules.d/99-nix-operator.rules")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // 文件不存在，返回空字符串
		}
		return "", err
	}
	return string(data), nil
}
