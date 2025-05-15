package network

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/utils"
)

type Ifupdown struct{}

func (ifd *Ifupdown) IsInstall(ctx context.Context) bool {
	_, err := os.Stat("/sbin/ifup")
	return err == nil
}

func (ifd *Ifupdown) findConfig(iface config.Interface) (string, error) {
	// 检查主配置文件
	mainConfig, err := os.ReadFile("/etc/network/interfaces")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(mainConfig)))
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), fmt.Sprintf("iface %s ", iface.Name)) {
				return "/etc/network/interfaces", nil
			}
		}
	}

	// 检查 interfaces.d 目录
	files, err := os.ReadDir("/etc/network/interfaces.d")
	if err != nil {
		return "", fmt.Errorf("failed to read interfaces.d directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join("/etc/network/interfaces.d", file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if strings.Contains(string(data), fmt.Sprintf("iface %s ", iface.Name)) {
			return path, nil
		}
	}

	return fmt.Sprintf("/etc/network/interfaces.d/%s", iface.Name), nil
}

func (ifd *Ifupdown) Configure(ctx context.Context, iface config.Interface) error {
	configPath, err := ifd.findConfig(iface)
	if err != nil {
		return err
	}

	var desired strings.Builder
	desired.WriteString(config.CommentHeader)
	desired.WriteString(fmt.Sprintf("auto %s\n", iface.Name))

	// IPv4 配置
	if iface.IPAddress != "" {
		desired.WriteString(fmt.Sprintf("iface %s inet static\n", iface.Name))
		desired.WriteString(fmt.Sprintf("    address %s\n", iface.IPAddress))
		if iface.Gateway != "" {
			desired.WriteString(fmt.Sprintf("    gateway %s\n", iface.Gateway))
		}
	}

	// IPv6 配置
	if iface.IPv6Address != "" {
		desired.WriteString(fmt.Sprintf("iface %s inet6 static\n", iface.Name))
		desired.WriteString(fmt.Sprintf("    address %s\n", iface.IPv6Address))
		if iface.IPv6Gateway != "" {
			desired.WriteString(fmt.Sprintf("    gateway %s\n", iface.IPv6Gateway))
		}
	}

	desired.WriteString(fmt.Sprintf("    mtu %d\n", iface.MTU))

	// 读取现有配置
	current, err := os.ReadFile(configPath)
	if err == nil && string(current) == desired.String() {
		return nil // 配置相同，无需更新
	}

	return utils.AtomicWriteFile([]byte(desired.String()), configPath, 0644)
}

func (ifd *Ifupdown) ReloadIfy(ctx context.Context) error {
	if !isServiceActive(ctx, "networking") {
		return nil
	}
	cmd := exec.CommandContext(ctx, "systemctl", "restart", "networking")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart networking: %v, output: %s", err, output)
	}
	return nil
} 