package firewall

import (
	"context"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

func init() {
	controller.RegisterHandler("firewall", &LinuxFirewallHandler{})
}

type LinuxFirewallHandler struct{}

func (h *LinuxFirewallHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (h *LinuxFirewallHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	panic("unimplemented")
}
