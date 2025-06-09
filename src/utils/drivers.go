package utils

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func DetectAndInstallDrivers() {
	fmt.Println("Detecting and installing drivers...")
	gpu := detectGPU()
	var drivers []string
	if strings.Contains(strings.ToLower(gpu), "nvidia") {
		drivers = append(drivers, "nvidia")
	} else if strings.Contains(strings.ToLower(gpu), "amd") {
		drivers = append(drivers, "xf86-video-amdgpu")
	} else if strings.Contains(strings.ToLower(gpu), "intel") {
		drivers = append(drivers, "xf86-video-intel")
	}
	if len(drivers) > 0 {
		cmd := exec.Command("pacstrap", append([]string{"/mnt"}, drivers...)...)
		cmd.Run()
	}
}

func detectGPU() string {
	cmd := exec.Command("lspci")
	out, _ := cmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D") {
			return line
		}
	}
	return ""
}
