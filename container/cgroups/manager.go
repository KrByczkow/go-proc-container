package cgroups

import (
	"os"
	"path/filepath"
	"strconv"
)

func CGGet(cgName, cgLabel string, asTasks bool) (string, error) {
	cgPath, err := CGMountPath()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(cgPath, cgName)

	if asTasks {
		filePath = filepath.Join(filePath, "tasks")
	}

	filePath = filepath.Join(filePath, cgLabel)

	val, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(val), err
}

func CGSet(cgName, cgLabel, cgValue string, asTasks bool) error {
	cgPath, err := CGMountPath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(cgPath, cgName)

	if asTasks {
		filePath = filepath.Join(filePath, "tasks")
	}

	filePath = filepath.Join(filePath, cgLabel)

	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(cgValue)
	return err
}

func CGRun(cgName string) error {
	cgPath, err := CGMountPath()
	if err != nil {
		return err
	}

	location := filepath.Join(cgPath, cgName, "tasks", "cgroup.procs")

	file, err := os.OpenFile(location, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(os.Getpid()))
	return err
}
