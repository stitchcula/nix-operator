package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"go.xbrother.com/nix-operator/pkg/config"
	"gopkg.in/yaml.v3"

	"github.com/fsnotify/fsnotify"
)

type OSInfo struct {
	ID         string // 发行版ID，如 "ubuntu", "centos"
	VersionID  string // 发行版版本，如 "20.04", "7"
	KernelName string // 内核名称，如 "Linux"
	KernelVer  string // 内核版本
}

type Controller struct {
	configPath string
	handlers   map[string]Handler // key 是处理器类型
	osInfo     OSInfo
}

type Handler interface {
	// Match 检查是否支持该操作系统
	Match(osInfo OSInfo) bool
	// Reconcile 处理配置
	Reconcile(ctx context.Context, config *config.SystemConfiguration) error
}

var handlerFactories = make(map[string][]Handler)

// RegisterHandler 注册处理器工厂
func RegisterHandler(typeName string, handler Handler) {
	handlerFactories[typeName] = append(handlerFactories[typeName], handler)
}

func NewController(configPath string) (*Controller, error) {
	osInfo, err := getOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %v", err)
	}

	handlers := make(map[string]Handler)

	// 为每种类型选择合适的处理器
	requiredTypes := []string{
		"network",
		"dns",
		"hosts",
		"firewall",
		"system",
		"ntp",
		"serial",
		"udev",
	}
	for _, typeName := range requiredTypes {
		typedHandlers := handlerFactories[typeName]
		if len(typedHandlers) == 0 {
			return nil, fmt.Errorf("no handler registered for type: %s", typeName)
		}

		// 查找匹配的处理器
		var matched bool
		for _, handler := range typedHandlers {
			if handler.Match(osInfo) {
				handlers[typeName] = handler
				matched = true
				break
			}
		}

		if !matched {
			return nil, fmt.Errorf("no compatible handler found for type %s on OS %s %s",
				typeName, osInfo.ID, osInfo.VersionID)
		}
	}

	return &Controller{
		configPath: configPath,
		handlers:   handlers,
		osInfo:     osInfo,
	}, nil
}

func getOSInfo() (OSInfo, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return OSInfo{}, err
	}

	info := OSInfo{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.Trim(parts[1], "\"")
		switch parts[0] {
		case "ID":
			info.ID = value
		case "VERSION_ID":
			info.VersionID = value
		}
	}

	// 获取内核信息
	kernel, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return OSInfo{}, err
	}
	info.KernelVer = strings.TrimSpace(string(kernel))
	info.KernelName = "Linux" // 可以根据需要扩展

	return info, nil
}

func (c *Controller) Run() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					c.reconcile()
				}
			case err := <-watcher.Errors:
				log.Printf("Error: %v", err)
			}
		}
	}()

	err = watcher.Add(c.configPath)
	if err != nil {
		return err
	}

	// 初始调谐
	c.reconcile()

	// 保持运行
	select {}
}

func (c *Controller) reconcile() {
	ctx := context.Background()
	config, err := loadConfig(c.configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}

	for _, handler := range c.handlers {
		if err := handler.Reconcile(ctx, config); err != nil {
			log.Printf("Reconciliation error: %v", err)
		}
	}
}

func loadConfig(path string) (*config.SystemConfiguration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.SystemConfiguration
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
