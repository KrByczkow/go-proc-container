package net_tools

import (
	"github.com/vishvananda/netlink"
	"net"
)

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
