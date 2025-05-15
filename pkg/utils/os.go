package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type NodeSelector struct {
	MACAddress string `yaml:"macAddress"`
	Hostname   string `yaml:"hostname"`
}

func MatchNodeSelector(selector NodeSelector) (bool, error) {
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
