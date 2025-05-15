package network

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

func init() {
	controller.RegisterHandler("network", &LinuxNetworkHandler{})
}

type LinuxNetworkHandler struct{}

func (h *LinuxNetworkHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (h *LinuxNetworkHandler) matchNodeSelector(selector config.NodeSelector) (bool, error) {
	if selector.Hostname != "" {
		hostname, err := os.Hostname()
		if err != nil {
			return false, fmt.Errorf("failed to get hostname: %v", err)
		}
		if hostname != selector.Hostname {
			return false, nil
		}
	}

	if selector.MACAddress != "" {
		interfaces, err := net.Interfaces()
		if err != nil {
			return false, fmt.Errorf("failed to get network interfaces: %v", err)
		}

		macFound := false
		for _, iface := range interfaces {
			if strings.EqualFold(iface.HardwareAddr.String(), selector.MACAddress) {
				macFound = true
				break
			}
		}
		if !macFound {
			return false, nil
		}
	}

	return true, nil
}

func (h *LinuxNetworkHandler) configureInterface(ctx context.Context, iface config.Interface) error {
	// 设置IP地址
	cmd := exec.CommandContext(ctx, "ip", "addr", "add", iface.IPAddress, "dev", iface.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set IP address: %v, output: %s", err, output)
	}

	// 设置MTU
	cmd = exec.CommandContext(ctx, "ip", "link", "set", iface.Name, "mtu", fmt.Sprintf("%d", iface.MTU))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set MTU: %v, output: %s", err, output)
	}

	// 设置MAC地址
	cmd = exec.CommandContext(ctx, "ip", "link", "set", iface.Name, "address", iface.MACAddress)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set MAC address: %v, output: %s", err, output)
	}

	// 设置网关
	if iface.Gateway != "" {
		cmd = exec.CommandContext(ctx, "ip", "route", "add", "default", "via", iface.Gateway, "dev", iface.Name)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to set gateway: %v, output: %s", err, output)
		}
	}

	return nil
}

func (h *LinuxNetworkHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	for _, iface := range cfg.Spec.Network.Interfaces {
		match, err := h.matchNodeSelector(iface.NodeSelector)
		if err != nil {
			return fmt.Errorf("failed to check node selector: %v", err)
		}

		if !match {
			continue
		}

		if err := h.configureInterface(ctx, iface); err != nil {
			return err
		}
	}
	return nil
}
