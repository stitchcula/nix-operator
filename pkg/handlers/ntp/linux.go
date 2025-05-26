package ntp

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

//go:embed chrony.conf.tpl
var chronyConfigTemplate string

func init() {
	controller.RegisterHandler("ntp", &LinuxNTPHandler{})
}

type LinuxNTPHandler struct{}

func (h *LinuxNTPHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

type chronyConfig struct {
	Servers []string
}

func (h *LinuxNTPHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	if !cfg.Spec.System.NTP.Enabled {
		return nil
	}

	// 准备模板数据
	templateData := chronyConfig{
		Servers: cfg.Spec.System.NTP.Servers,
	}

	// 解析模板
	tmpl, err := template.New("chrony").Parse(chronyConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// 渲染配置
	var content strings.Builder
	if err := tmpl.Execute(&content, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	desiredContent := content.String()

	// 读取现有配置
	currentContent, err := os.ReadFile("/etc/chrony.conf")
	if err == nil {
		// 配置文件存在，比较内容
		if string(currentContent) == desiredContent {
			return nil // 配置相同，无需更新
		}
	}
	// 如果文件不存在或读取失败，继续写入新配置

	// 使用工具函数原子性写入文件
	if err := utils.AtomicWriteFile([]byte(desiredContent), "/etc/chrony.conf", 0644); err != nil {
		return fmt.Errorf("failed to write chrony.conf: %v", err)
	}

	// 重新加载配置（不需要 root 权限）
	cmd := exec.CommandContext(ctx, "chronyc", "reload", "sources")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload chronyd config: %v, output: %s", err, output)
	}

	return nil
}
