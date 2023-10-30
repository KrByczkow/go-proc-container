package net_tools

import (
	"code.cloudfoundry.org/guardian/kawasaki/netns"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
)

type NetworkConfig struct {
	BridgeName     string
	BridgeIP       net.IP
	ContainerIP    net.IP
	Subnet         *net.IPNet
	VethNamePrefix string
}

type Config interface {
	Apply(netConfig NetworkConfig, pid int) error
}

type Host struct {
	BridgeCreator BridgeCreator
	VethCreator   VethCreator
}

type Container struct {
	NsExecutor *netns.Execer
}

type BridgeCreator interface {
	Create(name string, ip net.IP, subnet *net.IPNet) (*net.Interface, error)
	Attach(bridge, hostVeth *net.Interface) error
}

type VethCreator interface {
	Create(vethNamePrefix string) (*net.Interface, *net.Interface, error)
	MoveNetToNS(containerVeth *net.Interface, pid int) error
}

type NetConfig struct {
	HostConfig      Config
	ContainerConfig Config
}

func New(hostConfiguration, containerConfiguration Config) *NetConfig {
	return &NetConfig{
		HostConfig:      hostConfiguration,
		ContainerConfig: containerConfiguration,
	}
}

func NewHostConfig(bridgeCreator BridgeCreator, vethCreator VethCreator) *Host {
	return &Host{
		BridgeCreator: bridgeCreator,
		VethCreator:   vethCreator,
	}
}

func NewContainerConfig(NsExecutor *netns.Execer) *Container {
	return &Container{
		NsExecutor: NsExecutor,
	}
}

func (conf *NetConfig) ConfigureHost(netConfig NetworkConfig, pid int) error {
	return conf.HostConfig.Apply(netConfig, pid)
}

func (conf *NetConfig) ConfigureContainer(netConfig NetworkConfig, pid int) error {
	return conf.ContainerConfig.Apply(netConfig, pid)
}

func (h *Host) Apply(netConfig NetworkConfig, pid int) error {
	bridge, err := h.BridgeCreator.Create(netConfig.BridgeName, netConfig.BridgeIP, netConfig.Subnet)
	if err != nil {
		return err
	}

	hostVeth, containerVeth, err := h.VethCreator.Create(netConfig.VethNamePrefix)
	if err != nil {
		return err
	}

	err = h.BridgeCreator.Attach(bridge, hostVeth)
	if err != nil {
		return err
	}

	err = h.VethCreator.MoveNetToNS(containerVeth, pid)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) Apply(netConfig NetworkConfig, pid int) error {
	nsFile, err := os.Open(fmt.Sprintf("/proc/%d/ns/net", pid))
	defer nsFile.Close()
	if err != nil {
		return fmt.Errorf("unable to find network namespace under the process id %d", pid)
	}

	cbFunc := func() error {
		containerVethName := fmt.Sprintf("%s1", netConfig.VethNamePrefix)
		link, err := netlink.LinkByName(containerVethName)
		if err != nil {
			return fmt.Errorf("container veth '%s' not found", containerVethName)
		}

		addr := &netlink.Addr{IPNet: &net.IPNet{IP: netConfig.ContainerIP, Mask: netConfig.Subnet.Mask}}
		err = netlink.AddrAdd(link, addr)
		if err != nil {
			return fmt.Errorf("unable to assign IP address '%s' to %s: %v", netConfig.ContainerIP, containerVethName, err)
		}

		if err := netlink.LinkSetUp(link); err != nil {
			return err
		}

		route := &netlink.Route{
			Scope:     netlink.SCOPE_UNIVERSE,
			LinkIndex: link.Attrs().Index,
			Gw:        netConfig.BridgeIP,
		}

		return netlink.RouteAdd(route)
	}

	return c.NsExecutor.Exec(nsFile, cbFunc)
}
