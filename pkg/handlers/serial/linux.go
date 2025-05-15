package serial

import (
	"context"
	"fmt"
	"os/exec"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

type LinuxSerialHandler struct{}

func (h *LinuxSerialHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	for _, serial := range cfg.Spec.Serials {
		cmd := exec.CommandContext(ctx, "stty",
			"-F", serial.Device,
			fmt.Sprintf("%d", serial.BaudRate),
			fmt.Sprintf("cs%d", serial.DataBits),
			fmt.Sprintf("-%s", serial.Parity),
			fmt.Sprintf("-%sstopb", map[int]string{1: "", 2: "-"}[serial.StopBits]))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to configure serial port %s: %v, output: %s", serial.Device, err, output)
		}
	}
	return nil
}

type LinuxSerialFactory struct{}

func (f *LinuxSerialFactory) Create() controller.Handler {
	return &LinuxSerialHandler{}
}

func (f *LinuxSerialFactory) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func init() {
	controller.RegisterHandler("serial", &LinuxSerialFactory{})
}
