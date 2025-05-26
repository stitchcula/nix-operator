package serial

import (
	"context"

	"go.xbrother.com/nix-operator/pkg/controller"
)

type ModeSwitcher interface {
	Match(osInfo controller.OSInfo) bool
	Switch(ctx context.Context, dev string, mode string) error
}

type LightingAModeSwitcher struct {
	Mode string
}

func (m *LightingAModeSwitcher) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (m *LightingAModeSwitcher) Switch(ctx context.Context, dev string, mode string) error {
	return nil
}

type LightingBModeSwitcher struct {
	Mode string
}

func (m *LightingBModeSwitcher) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (m *LightingBModeSwitcher) Switch(ctx context.Context, dev string, mode string) error {
	return nil
}

type RainbowBModeSwitcher struct {
	Mode string
}

func (m *RainbowBModeSwitcher) Match(osInfo controller.OSInfo) bool {
	return osInfo.KernelName == "Linux"
}

func (m *RainbowBModeSwitcher) Switch(ctx context.Context, dev string, mode string) error {
	return nil
}
