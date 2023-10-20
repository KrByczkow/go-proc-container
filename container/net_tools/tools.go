package net_tools

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"time"
)

func HasNewNetworkIntf(maxIntervalTimeout time.Duration) error {
	checkInterval := time.Second
	timeStarted := time.Now()

	initInterfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	initCount := len(initInterfaces)

	for {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}

		if len(interfaces) > initCount {
			return nil
		}

		if time.Since(timeStarted) > maxIntervalTimeout {
			return fmt.Errorf("timeout after waiting for network")
		}

		time.Sleep(checkInterval)
	}
}

func NetIntfExists(name string) bool {
	_, err := net.InterfaceByName(name)
	return err == nil
}

func (v *Veth) MoveNetToNS(veth *net.Interface, pid int) error {
	vethLink, err := netlink.LinkByName(veth.Name)
	if err != nil {
		return err
	}

	return netlink.LinkSetNsPid(vethLink, pid)
}
