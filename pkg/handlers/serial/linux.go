package serial

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"go.xbrother.com/nix-operator/pkg/config"
	"go.xbrother.com/nix-operator/pkg/controller"
)

type Config struct {
	Device      string             `json:"device"`
	DisplayName string             `json:"displayName"`
	BaudRate    int                `json:"baudRate"`
	DataBits    int                `json:"dataBits"`
	StopBits    int                `json:"stopBits"`
	Parity      string             `json:"parity"`
	Mode        string             `json:"mode"`                  // "rs232" 或 "rs485"
	RS485       *RS485Config       `json:"rs485,omitempty"`       // RS485 特定配置
	Transparent *TransparentConfig `json:"transparent,omitempty"` // 透传配置
}

type RS485Config struct {
	Enabled            bool `json:"enabled"`            // 启用 RS485 模式
	RTSOnSend          bool `json:"rtsOnSend"`          // 发送时 RTS 信号状态
	RTSAfterSend       bool `json:"rtsAfterSend"`       // 发送后 RTS 信号状态
	RTSDelay           int  `json:"rtsDelay"`           // RTS 延迟时间（微秒）
	DelayRTSBeforeSend int  `json:"delayRTSBeforeSend"` // 发送前 RTS 延迟（微秒）
	DelayRTSAfterSend  int  `json:"delayRTSAfterSend"`  // 发送后 RTS 延迟（微秒）
	ReceiveTimeout     int  `json:"receiveTimeout"`     // 接收超时（毫秒）
}

type TransparentConfig struct {
	Enabled    bool   `json:"enabled"`    // 启用透传功能
	Protocol   string `json:"protocol"`   // 透传协议 "tcp" 或 "udp"
	ListenAddr string `json:"listenAddr"` // 监听地址，如 "0.0.0.0:8080"
	BufferSize int    `json:"bufferSize"` // 缓冲区大小（字节）
	Timeout    int    `json:"timeout"`    // 连接超时（秒）
}

func init() {
	controller.RegisterHandler("SerialConfiguration", &LinuxSerialHandler{modeSwitcher: &LightingAModeSwitcher{}})
	controller.RegisterHandler("SerialConfiguration", &LinuxSerialHandler{modeSwitcher: &LightingBModeSwitcher{}})
	controller.RegisterHandler("SerialConfiguration", &LinuxSerialHandler{modeSwitcher: &RainbowBModeSwitcher{}})
}

type LinuxSerialHandler struct {
	modeSwitcher       ModeSwitcher
	transparentServers map[string]*TransparentServer
	mu                 *sync.RWMutex
}

func (f *LinuxSerialHandler) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux" && f.modeSwitcher.Match(osInfo)
}

func (h *LinuxSerialHandler) Reconcile(ctx context.Context, cfg *config.SystemConfiguration) error {
	if h.transparentServers == nil {
		h.transparentServers = make(map[string]*TransparentServer)
	}

	for _, serial := range cfg.Spec.Serials {
		// 配置基本串口参数
		if err := h.configureSerialParams(ctx, serial); err != nil {
			return err
		}

		// 配置 RS232/RS485 模式
		if err := h.configureSerialMode(ctx, serial); err != nil {
			return err
		}

		// 配置透传功能
		if err := h.configureTransparent(ctx, serial); err != nil {
			return err
		}
	}
	return nil
}

func (h *LinuxSerialHandler) configureSerialParams(ctx context.Context, serial config.SerialConfig) error {
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
	return nil
}

func (h *LinuxSerialHandler) configureSerialMode(ctx context.Context, serial config.SerialConfig) error {
	if serial.Mode == "" {
		return nil // 如果没有指定模式，跳过
	}

	if err := h.modeSwitcher.Switch(ctx, serial.Device, serial.Mode); err != nil {
		return fmt.Errorf("failed to switch serial mode: %v", err)
	}

	if serial.Mode == "rs485" {
		return configureRS485(ctx, serial)
	}

	return nil
}

func (h *LinuxSerialHandler) configureTransparent(ctx context.Context, serial config.SerialConfig) error {
	// 如果没有透传配置或未启用，跳过
	if serial.Transparent == nil || !serial.Transparent.Enabled {
		return nil
	}

	// TODO: 实现串口透传功能
	// 这里需要实现 TCP/UDP 服务器和串口数据转发逻辑
	return nil
}
