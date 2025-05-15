package main

import (
	"flag"
	"log"

	"go.xbrother.com/nix-operator/pkg/controller"

	// 注册所有处理器
	_ "go.xbrother.com/nix-operator/pkg/handlers/dns"
	_ "go.xbrother.com/nix-operator/pkg/handlers/firewall"
	_ "go.xbrother.com/nix-operator/pkg/handlers/hosts"
	_ "go.xbrother.com/nix-operator/pkg/handlers/network"
	_ "go.xbrother.com/nix-operator/pkg/handlers/ntp"
	_ "go.xbrother.com/nix-operator/pkg/handlers/serial"
	_ "go.xbrother.com/nix-operator/pkg/handlers/system"
	_ "go.xbrother.com/nix-operator/pkg/handlers/udev"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	controller, err := controller.NewController(*configPath)
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	if err := controller.Run(); err != nil {
		log.Fatalf("Error running controller: %v", err)
	}
}
