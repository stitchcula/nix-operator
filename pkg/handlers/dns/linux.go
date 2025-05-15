package dns

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
	"go.xbrother.com/nix-operator/pkg/utils"
)

func init() {
	controller.RegisterHandler("dns", &LinuxDNSHandler{})
}

type LinuxDNSHandler struct{}

func (f *LinuxDNSHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (h *LinuxDNSHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	// 读取现有的 resolv.conf
	currentServers, err := h.getCurrentNameservers()
	if err != nil {
		return fmt.Errorf("failed to read current nameservers: %v", err)
	}

	// 比较现有配置和期望配置
	desiredServers := cfg.Spec.Network.DNS.Nameservers
	if h.areNameserversEqual(currentServers, desiredServers) {
		return nil // 配置一致，无需更新
	}

	// 生成新的 resolv.conf 内容
	var content strings.Builder
	content.WriteString(config.CommentHeader)
	for _, ns := range desiredServers {
		content.WriteString(fmt.Sprintf("nameserver %s\n", ns))
	}

	// 使用工具函数原子性写入文件
	return utils.AtomicWriteFile([]byte(content.String()), "/etc/resolv.conf", 0644)
}

func (h *LinuxDNSHandler) getCurrentNameservers() ([]string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var servers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				servers = append(servers, fields[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func (h *LinuxDNSHandler) areNameserversEqual(current, desired []string) bool {
	// 排除空条目
	current = slices.DeleteFunc(current, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	desired = slices.DeleteFunc(desired, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})

	if len(current) != len(desired) {
		return false
	}

	// 复制切片以避免修改原始数据
	currentCopy := make([]string, len(current))
	desiredCopy := make([]string, len(desired))
	copy(currentCopy, current)
	copy(desiredCopy, desired)

	// 排序后比较，忽略顺序
	slices.Sort(currentCopy)
	slices.Sort(desiredCopy)

	return slices.Equal(currentCopy, desiredCopy)
}
