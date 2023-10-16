package container

import (
	"ProcContainer/utils"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Hostname(name string) {
	utils.PanicOnError(syscall.Sethostname([]byte(name)))
}

func InitContainer(hostname string, exe string, cmd []string) {
	utils.PanicOnError(syscall.Chdir("/"))
	utils.PanicOnError(syscall.Mount("proc", "proc", "proc", 0, ""))

	Hostname(hostname)

	utils.PanicOnError(syscall.Unshare(syscall.CLONE_NEWPID | syscall.CLONE_NEWNS))

	RunProcess(exe, cmd)
}

func RestartSelf(cmd []string) {
	comm := exec.Command("/proc/self/exe", cmd...)
	comm.Stdin = os.Stdin
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stderr
	comm.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	err := comm.Run()
	if err != nil {
		if !strings.Contains(err.Error(), "exit status") {
			utils.PanicOnError(err)
		}
	}
}

func RunProcess(exe string, cmd []string) {
	comm := exec.Command(exe, cmd...)
	comm.Stdin = os.Stdin
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stderr

	err := comm.Run()
	if err != nil {
		if !strings.Contains(err.Error(), "exit status") {
			utils.PanicOnError(err)
		}
	}
}
