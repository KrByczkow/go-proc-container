package net_tools

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
)

type Veth struct{}

func NewVeth() *Veth {
	return &Veth{}
}

func (v *Veth) Create(namePrefix string) (*net.Interface, *net.Interface, error) {
	hVethName := fmt.Sprintf("%s0", namePrefix)
	cVethName := fmt.Sprintf("%s1", namePrefix)

	if NetIntfExists(hVethName) {
		return vethInterfaces(hVethName, cVethName)
	}

	vethLinkAttrs := netlink.NewLinkAttrs()
	vethLinkAttrs.Name = hVethName

	veth := &netlink.Veth{
		LinkAttrs: vethLinkAttrs,
		PeerName:  cVethName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, nil, err
	}

	if err := netlink.LinkSetUp(veth); err != nil {
		return nil, nil, err
	}

	return vethInterfaces(hVethName, cVethName)
}

func vethInterfaces(hVethName, cVethName string) (*net.Interface, *net.Interface, error) {
	hostVeth, err := net.InterfaceByName(hVethName)
	if err != nil {
		return nil, nil, err
	}

	containerVeth, err := net.InterfaceByName(cVethName)
	if err != nil {
		return nil, nil, err
	}

	return hostVeth, containerVeth, nil
}
