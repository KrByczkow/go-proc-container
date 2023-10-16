package contNet

/*
Sources that I used for Network Namespace Setup:
https://www.cyberciti.biz/faq/how-to-configuring-bridging-in-debian-linux/
https://blog.scottlowe.org/2013/09/04/introducing-linux-network-namespaces/
https://medium.com/@teddyking/namespaces-in-go-network-fdcf63e76100
*/

import (
	"fmt"
	"os/exec"
	"strings"
)

func handleCommandError(exe string, args ...string) {
	comm := exec.Command(exe, args...)
	err := comm.Run()

	if err != nil {
		if strings.Contains(err.Error(), "exit status") {
			b, bErr := comm.Output()
			fmt.Printf("Length of the byte array: %d\n", len(b))
			if bErr == nil {
				panic(err)
			}

			strErr := string(b)
			fmt.Printf("why you aint working\n%s\n", strErr)
			if !(strings.Contains(strErr, "Cannot create namespace") && strings.Contains(strErr, "File exists")) {
				panic(strErr)
			}
		}
	}
}

func Run(netNs string) error {
	return nil
}

func Destroy(netNs string) error {
	return nil
}

func MkNetInterface(namespaceName, linkName, deviceName, address string) {
	handleCommandError("ip", "netns", "add", namespaceName)
	handleCommandError("ip", "netns", "list")
	handleCommandError("ip", "link", "add", linkName, "type", "veth", "peer", "name", deviceName)
	handleCommandError("ip", "link", "set", deviceName, "netns", namespaceName)
	handleCommandError("ip", "netns", "exec", namespaceName, "ip", "addr", "add", address+"/24", "dev", deviceName)
	handleCommandError("ip", "netns", "exec", namespaceName, "ip", "link", "set", "dev", deviceName, "up")
}

func MkNetBridge(namespaceName string) {
}
