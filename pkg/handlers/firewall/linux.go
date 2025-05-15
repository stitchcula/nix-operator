package firewall

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

type LinuxFirewallHandler struct{}

func (h *LinuxFirewallHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	// 清除现有规则
	cmd := exec.CommandContext(ctx, "iptables", "-F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to flush iptables rules: %v, output: %s", err, output)
	}

	// 设置默认策略
	cmd = exec.CommandContext(ctx, "iptables", "-P", "INPUT", "DROP")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set default policy: %v, output: %s", err, output)
	}

	// 允许已建立的连接
	cmd = exec.CommandContext(ctx, "iptables", "-A", "INPUT", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to allow established connections: %v, output: %s", err, output)
	}

	// 应用新规则
	for _, rule := range cfg.Spec.Network.Firewall.Rules {
		cmd = exec.CommandContext(ctx, "iptables", "-A", "INPUT",
			"-p", rule.Protocol,
			"--dport", fmt.Sprintf("%d", rule.Port),
			"-j", strings.ToUpper(rule.Action))
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add rule for port %d: %v, output: %s", rule.Port, err, output)
		}
	}

	return nil
}

type LinuxFirewallFactory struct{}

func (f *LinuxFirewallFactory) Create() controller.Handler {
	return &LinuxFirewallHandler{}
}

func (f *LinuxFirewallFactory) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func init() {
	controller.RegisterHandler("firewall", &LinuxFirewallFactory{})
}
