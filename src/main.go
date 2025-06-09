package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type SystemInfo struct {
	CPU  string
	GPU  string
	RAM  string
	Disk string
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Root privileges required")
		os.Exit(1)
	}

	if !isArchLinux() {
		fmt.Println("This installer is for Arch Linux only")
		os.Exit(1)
	}

	sysInfo := detectHardware()
	installBase()
	installDrivers(sysInfo)
	configureSystem()
}

func isArchLinux() bool {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "ID=arch") {
			return true
		}
	}
	return false
}

func detectHardware() SystemInfo {
	var info SystemInfo

	// CPU detection
	cpuCmd := exec.Command("lscpu")
	cpuOut, _ := cpuCmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(cpuOut)))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Model name") {
			info.CPU = strings.Split(scanner.Text(), ":")[1]
			break
		}
	}

	// GPU detection
	gpuCmd := exec.Command("lspci")
	gpuOut, _ := gpuCmd.Output()
	scanner = bufio.NewScanner(strings.NewReader(string(gpuOut)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D") {
			info.GPU = line
			break
		}
	}

	// RAM detection
	memCmd := exec.Command("free", "-h")
	memOut, _ := memCmd.Output()
	scanner = bufio.NewScanner(strings.NewReader(string(memOut)))
	scanner.Scan() // Skip header
	scanner.Scan() // Get Mem line
	info.RAM = strings.Fields(scanner.Text())[1]

	// Disk detection
	diskCmd := exec.Command("lsblk", "-d", "-o", "SIZE,MODEL")
	diskOut, _ := diskCmd.Output()
	info.Disk = string(diskOut)

	return info
}

func installBase() {
	cmd := exec.Command("pacstrap", "/mnt", "base", "linux", "linux-firmware")
	cmd.Run()
}

func installDrivers(sysInfo SystemInfo) {
	var drivers []string

	if strings.Contains(strings.ToLower(sysInfo.GPU), "nvidia") {
		drivers = append(drivers, "nvidia")
	} else if strings.Contains(strings.ToLower(sysInfo.GPU), "amd") {
		drivers = append(drivers, "xf86-video-amdgpu")
	} else if strings.Contains(strings.ToLower(sysInfo.GPU), "intel") {
		drivers = append(drivers, "xf86-video-intel")
	}

	if len(drivers) > 0 {
		cmd := exec.Command("pacstrap", append([]string{"/mnt"}, drivers...)...)
		cmd.Run()
	}
}

func configureSystem() {
	cmd := exec.Command("genfstab", "-U", "/mnt", ">>", "/mnt/etc/fstab")
	cmd.Run()
}
