package contNet

import (
	"errors"
	"fmt"
	"net"
)

func interfaceExists(interfaceName string) (bool, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return false, err
	}

	fmt.Printf("Given length of %d\n", len(inters))

	for _, inter := range inters {
		fmt.Printf("Processing %s\n", inter.Name)

		if inter.Name == interfaceName {
			return true, nil
		}
	}

	return false, errors.New("no such interface")
}
