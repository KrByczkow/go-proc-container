package utils

import (
	"os"
	"os/exec"
)

func PanicOnError(e error) {
	if e != nil {
		panic(e)
	}
}

func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)

	if err != nil {
		return "", err
	}

	return string(data), err
}

func RunCommand(cmd string, args ...string) error {
	comm := exec.Command(cmd, args...)
	comm.Stdin = os.Stdin
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stderr

	return comm.Run()
}
