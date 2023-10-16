package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	CGMountPathFindError   = errors.New("cgroup mount path not found")
	CGMountPathMountsError = errors.New("cannot open '/proc/mounts' for reading")
	CGEnableControlExists  = errors.New("the subtree control for the given control group name already exists")
	CGControlTypeUndefined = errors.New("undefined control type")
	CGControlTypeNotGiven  = errors.New("no control types were given")
)

type ControlGroup struct {
	cgPath  string
	Name    string
	Modules []string
}

func CGMountPath() (string, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner == nil {
		return "", CGMountPathMountsError
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) <= 1 {
			continue
		}

		if len(fields) >= 2 && strings.Contains(fields[0], "cgroup") {
			return fields[1], nil
		}
	}

	return "", CGMountPathFindError
}

func CGEnable(cgName, cgControlType string) error {
	if len(cgControlType) == 0 {
		return CGControlTypeNotGiven
	}

	cgControlType = strings.TrimSpace(cgControlType)
	if len(cgControlType) == 0 {
		return CGControlTypeUndefined
	}

	if strings.Contains(cgControlType, "+") {
		cgControlType = strings.ReplaceAll(cgControlType, "+", "")
	}

	cgMount, err := CGMountPath()
	if err != nil {
		return err
	}

	path := filepath.Join(cgMount, cgName)

	reader, err := os.OpenFile(filepath.Join(path, "cgroup.subtree_control"), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer reader.Close()

	bData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	data := string(bData)
	if strings.Contains(data, cgControlType) {
		return CGEnableControlExists
	}

	writeData := ""
	if strings.HasSuffix(data, "\n") {
		writeData += "\n"
	}

	writeData += "+" + cgControlType

	writer, err := os.OpenFile(filepath.Join(path, "cgroup.subtree_control"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer writer.Close()

	if _, err := writer.WriteString(writeData); err != nil {
		return err
	}

	return nil
}

func CGInit(cgName string) error {
	path, err := CGMountPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(path, cgName), 0744)
	if err != nil {
		return err
	}

	return os.MkdirAll(filepath.Join(path, cgName, "tasks"), 0744)
}

func CGVerifyProcess(paths []string) (bool, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", os.Getpid()))
	if err != nil {
		return false, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	if scanner == nil {
		return false, errors.New("")
	}

	foundPaths := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 {
			continue
		}

		opts := strings.SplitN(line, "::", 2)

		for _, path := range paths {
			if path == opts[1] {
				foundPaths++
			}
		}
	}

	return foundPaths == len(paths), nil
}
