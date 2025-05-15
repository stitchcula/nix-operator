package network

import (
	"context"
	"fmt"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type LinuxNetworkHandler struct {
	managers []INetworkManager
}

func init() {
	handler := &LinuxNetworkHandler{
		managers: []INetworkManager{
			&NetworkManager{},
			&Netplan{},
			&Ifupdown{},
		},
	}
	controller.RegisterHandler("network", handler)
}

func (h *LinuxNetworkHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (h *LinuxNetworkHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	for _, iface := range cfg.Spec.Network.Interfaces {
		match, err := utils.MatchNodeSelector(iface.NodeSelector)
		if err != nil {
			return fmt.Errorf("failed to check node selector: %v", err)
		}
		if !match {
			continue
		}

		// 为每个已安装的网络管理器生成配置
		for _, manager := range h.managers {
			if !manager.IsInstall(ctx) {
				continue
			}
			if err := manager.Configure(ctx, iface); err != nil {
				return err
			}
			if err := manager.ReloadIfy(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
