package system

import (
	"context"
	"fmt"
	"os/exec"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

type LinuxSystemHandler struct{}

func (h *LinuxSystemHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	cmd := exec.CommandContext(ctx, "timedatectl", "set-timezone", cfg.Spec.System.Timezone)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set timezone: %v, output: %s", err, output)
	}
	return nil
}

type LinuxSystemFactory struct{}

func (f *LinuxSystemFactory) Create() controller.Handler {
	return &LinuxSystemHandler{}
}

func (f *LinuxSystemFactory) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func init() {
	controller.RegisterHandler("system", &LinuxSystemFactory{})
}
