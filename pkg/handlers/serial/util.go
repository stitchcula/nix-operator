package serial

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"unsafe"

	"golang.org/x/sys/unix"

	"go.xbrother.com/nix-operator/pkg/config"
)

// RS485 ioctl 常量
const (
	TIOCGRS485 = 0x542E // 获取 RS485 配置
	TIOCSRS485 = 0x542F // 设置 RS485 配置
)

// RS485 配置标志
const (
	SER_RS485_ENABLED        = 1 << 0 // 启用 RS485 模式
	SER_RS485_RTS_ON_SEND    = 1 << 1 // 发送时 RTS 高电平
	SER_RS485_RTS_AFTER_SEND = 1 << 2 // 发送后 RTS 高电平
	SER_RS485_RX_DURING_TX   = 1 << 4 // 发送时接收
)

// RS485 配置结构体
type rs485Config struct {
	Flags              uint32    // 配置标志
	DelayRTSBeforeSend uint32    // 发送前 RTS 延迟（微秒）
	DelayRTSAfterSend  uint32    // 发送后 RTS 延迟（微秒）
	Padding            [5]uint32 // 填充字段
}

func configureRS485(ctx context.Context, serial config.SerialConfig) error {
	// 如果没有 RS485 配置，使用默认配置
	if serial.RS485 == nil {
		serial.RS485 = &config.RS485Config{
			Enabled:            true,
			RTSOnSend:          true,
			RTSAfterSend:       false,
			DelayRTSBeforeSend: 0,
			DelayRTSAfterSend:  0,
		}
	}

	// 打开串口设备
	file, err := os.OpenFile(serial.Device, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open serial device %s: %v", serial.Device, err)
	}
	defer file.Close()

	// 准备 RS485 配置
	rs485Cfg := rs485Config{
		DelayRTSBeforeSend: uint32(serial.RS485.DelayRTSBeforeSend),
		DelayRTSAfterSend:  uint32(serial.RS485.DelayRTSAfterSend),
	}

	// 设置配置标志
	if serial.RS485.Enabled {
		rs485Cfg.Flags |= SER_RS485_ENABLED
	}
	if serial.RS485.RTSOnSend {
		rs485Cfg.Flags |= SER_RS485_RTS_ON_SEND
	}
	if serial.RS485.RTSAfterSend {
		rs485Cfg.Flags |= SER_RS485_RTS_AFTER_SEND
	}

	// 使用 ioctl 设置 RS485 配置
	_, _, errno := unix.Syscall6(
		unix.SYS_IOCTL,
		file.Fd(),
		uintptr(TIOCSRS485),
		uintptr(unsafe.Pointer(&rs485Cfg)),
		0, 0, 0,
	)

	if errno != 0 {
		// 如果 ioctl 失败，尝试使用 setserial 命令作为备选方案
		return configureRS485WithSetserial(ctx, serial)
	}

	return nil
}

func configureRS485WithSetserial(ctx context.Context, serial config.SerialConfig) error {
	// 使用 setserial 命令配置 RS485 模式（备选方案）
	args := []string{serial.Device}

	// 设置 UART 类型
	args = append(args, "uart", "16550A")

	// 如果有 RTS 延迟配置，添加相应参数
	if serial.RS485 != nil && serial.RS485.DelayRTSBeforeSend > 0 {
		args = append(args, "rts_delay", strconv.Itoa(serial.RS485.DelayRTSBeforeSend))
	}

	cmd := exec.CommandContext(ctx, "setserial", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure RS485 with setserial for %s: %v, output: %s", serial.Device, err, output)
	}

	return nil
}
