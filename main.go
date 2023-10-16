package main

import (
	"ProcContainer/container"
	"ProcContainer/container/cgroups"
	"ProcContainer/container/contNet"
	"ProcContainer/utils"
	"fmt"
	"os"
)

func handleCgError(err error) {
	if err != nil {
		if err.Error() != "the subtree control for the given control group name already exists" {
			utils.PanicOnError(err)
		}
	}
}

func main() {
	if len(os.Args) < 1 || len(os.Args) == 1 {
		fmt.Printf("No arguments given, currently only has the program executable (\"%s\")\n", os.Args[0])
		os.Exit(1)
	}

	uid, gid := os.Getuid(), os.Getgid()
	if uid != 0 && gid != 0 {
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

		contNet.MkNetInterface("procnet", "veth0", "veth1", "10.32.10.1")

		// Logic before the first container initialization

		container.RestartSelf(append([]string{"child"}, args...))
	case "child":
		if len(args) == 0 {
			fmt.Println("Usage: child <command> [args]")
			os.Exit(1)
		}

		// Logic after the first container initialization

		// Initialize cgroups, if not yet made
		utils.PanicOnError(cgroups.CGInit("proc-container"))

		// Enable certain cgroup modules (CPU, Memory)
		handleCgError(cgroups.CGEnable("proc-container", "cpu"))
		handleCgError(cgroups.CGEnable("proc-container", "cpuset"))
		handleCgError(cgroups.CGEnable("proc-container", "memory"))

		// Set specific cgroup values
		handleCgError(cgroups.CGSet("proc-container", "cpu.max", "200000 1000000", true))
		handleCgError(cgroups.CGSet("proc-container", "memory.max", "10M", false))
		handleCgError(cgroups.CGSet("proc-container", "memory.min", "1M", false))
		handleCgError(cgroups.CGSet("proc-container", "memory.low", "3M", false))
		handleCgError(cgroups.CGSet("proc-container", "memory.high", "8M", false))

		// Start the cgroup containerization
		handleCgError(cgroups.CGRun("proc-container"))

		// Finally, run the container
		container.InitContainer("proc-container", args[0], args[1:])

		// Handle tasks after the main process (presumably a shell process) stops
	default:
		fmt.Printf("Unknown Argument \"%s\"\n", os.Args[1])
		os.Exit(1)
	}
}
