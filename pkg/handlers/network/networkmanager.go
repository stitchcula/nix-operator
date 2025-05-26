package network

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type NetworkManager struct{}

func (nm *NetworkManager) IsInstall(ctx context.Context) bool {
	_, err := os.Stat("/usr/sbin/NetworkManager")
	return err == nil
}

func (nm *NetworkManager) Configure(ctx context.Context, iface config.Interface) error {
	configPath := fmt.Sprintf("/etc/NetworkManager/system-connections/%s.nmconnection", iface.Name)

	var desired strings.Builder
	desired.WriteString(config.CommentHeader)
	desired.WriteString("[connection]\n")
	desired.WriteString(fmt.Sprintf("id=%s\n", iface.Name))
	desired.WriteString("type=ethernet\n")
	desired.WriteString(fmt.Sprintf("interface-name=%s\n", iface.Name))

	// IPv4 配置
	desired.WriteString("\n[ipv4]\n")
	if iface.IPAddress != "" {
		desired.WriteString(fmt.Sprintf("address1=%s\n", iface.IPAddress))
		desired.WriteString("method=manual\n")
		if iface.Gateway != "" {
			desired.WriteString(fmt.Sprintf("gateway=%s\n", iface.Gateway))
		}
	} else {
		desired.WriteString("method=disabled\n")
	}

	// IPv6 配置
	desired.WriteString("\n[ipv6]\n")
	if iface.IPv6Address != "" {
		desired.WriteString(fmt.Sprintf("address1=%s\n", iface.IPv6Address))
		desired.WriteString("method=manual\n")
		if iface.IPv6Gateway != "" {
			desired.WriteString(fmt.Sprintf("gateway=%s\n", iface.IPv6Gateway))
		}
	} else {
		desired.WriteString("method=disabled\n")
	}

	// 读取现有配置
	current, err := os.ReadFile(configPath)
	if err == nil && string(current) == desired.String() {
		return nil // 配置相同，无需更新
	}

	// 写入新配置
	return utils.AtomicWriteFile([]byte(desired.String()), configPath, 0600)
}

func (nm *NetworkManager) ReloadIfy(ctx context.Context) error {
	if !isServiceActive(ctx, "NetworkManager") {
		return nil
	}
	cmd := exec.CommandContext(ctx, "nmcli", "connection", "reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload NetworkManager: %v, output: %s", err, output)
	}
	return nil
}
