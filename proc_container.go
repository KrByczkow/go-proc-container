package main

import (
	"ProcContainer/container"
	. "ProcContainer/container/cgroups"
	"ProcContainer/container/net_tools"
	. "ProcContainer/utils"
	"code.cloudfoundry.org/guardian/kawasaki/netns"
	"fmt"
	"net"
	"os"
	"syscall"
)

func handleCgError(err error) {
	if err != nil {
		if err.Error() != "the subtree control for the given control group name already exists" {
			PanicOnError(err)
		}
	}
}

func initNetwork(pid int, namePrefix, bridgeAddress, containerAddress string) error {
	cBridge := net_tools.NewBridge()
	cVeth := net_tools.NewVeth()
	nsExec := &netns.Execer{}

	hostConfig := net_tools.NewHostConfiguration(cBridge, cVeth)
	containerConfig := net_tools.NewContainerConfiguration(nsExec)
	netSet := net_tools.New(hostConfig, containerConfig)

	bridgeIP, bridgeSubnet, err := net.ParseCIDR(bridgeAddress)
	PanicOnError(err)

	containerIP, _, err := net.ParseCIDR(containerAddress)
	PanicOnError(err)

	netConfig := net_tools.NetworkConfig{
		BridgeName:     "brdg0",
		BridgeIP:       bridgeIP,
		ContainerIP:    containerIP,
		Subnet:         bridgeSubnet,
		VethNamePrefix: namePrefix,
	}

	PanicOnError(netSet.ConfigureHost(netConfig, pid))
	PanicOnError(netSet.ConfigureContainer(netConfig, pid))

	return nil
}

func main() {
	if len(os.Args) < 1 || len(os.Args) == 1 {
		fmt.Printf("No arguments given, currently only has the program executable (\"%s\")\n", os.Args[0])
		os.Exit(1)
	}

	if os.Getuid() != 0 && os.Getgid() != 0 {
		fmt.Println("Cannot run as normal user.")
		fmt.Println("Make sure to run this process as root.")
		os.Exit(1)
		return
	}

	executor := os.Args[1]
	args := os.Args[2:]

	switch executor {
	case "run":
		if len(args) == 0 {
			fmt.Println("Usage: run <command> [args]")
			os.Exit(1)
		}

		comm, pid := container.RestartSelf(append([]string{"child"}, args...))

		err := initNetwork(pid, "veth", "10.15.21.1/24", "10.15.21.2/24")
		if err != nil {
			fmt.Printf("An error occurred initializing Network\n%v\n", err)

			err = comm.Process.Signal(syscall.SIGINT)
			if err != nil {
				fmt.Printf("An error occurred while sending a signal to %d: %v\n", pid, err)
			}

			return
		}

		if err := comm.Wait(); err != nil {
			fmt.Printf("ErrorHandle for Process %d: %v\n", pid, err)
		}
	case "child":
		if len(args) == 0 {
			fmt.Println("Usage: child <command> [args]")
			os.Exit(1)
		}

		// Initialize cgroups, if not yet made
		PanicOnError(CGInit("proc-container"))

		// Enable certain cgroup modules (CPU, Memory)
		handleCgError(CGEnable("proc-container", "cpu"))
		handleCgError(CGEnable("proc-container", "cpuset"))
		handleCgError(CGEnable("proc-container", "memory"))

		// Set specific cgroup values
		handleCgError(CGSet("proc-container", "cpu.max", "200000 1000000", true))
		handleCgError(CGSet("proc-container", "memory.max", "50M", false))
		handleCgError(CGSet("proc-container", "memory.min", "1M", false))
		handleCgError(CGSet("proc-container", "memory.low", "3M", false))
		handleCgError(CGSet("proc-container", "memory.high", "48M", false))

		// Start the cgroup containerization
		handleCgError(CGRun("proc-container"))

		// Finally, run the container
		container.InitContainer("proc-container", args[0], args[1:])
	default:
		fmt.Printf("Unknown Argument \"%s\"\n", os.Args[1])
		os.Exit(1)
	}
}
