package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ConfirmConfig(cfg InstallConfig) {
	fmt.Println("Installation config:")
	fmt.Printf("Disk: %s\nHostname: %s\nLocale: %s\nTimezone: %s\nBootloader: %s\n", cfg.Disk, cfg.Hostname, cfg.Locale, cfg.Timezone, cfg.Bootloader)
	fmt.Print("Continue? (y/N): ")
	var resp string
	fmt.Scanln(&resp)
	if strings.ToLower(resp) != "y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}
}

func FinalMessage(cfg InstallConfig) {
	fmt.Println("Installation complete! You can reboot now.")
}

func IsArchLinux() bool {
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

func FindPartition(parts []Partition, name string) *Partition {
	for i := range parts {
		if parts[i].Name == name {
			return &parts[i]
		}
	}
	return nil
}

func ReadDevice(reader *bufio.Reader, parts []Partition) string {
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			return ""
		}
		for _, p := range parts {
			if p.Name == input {
				return input
			}
		}
		fmt.Print("Enter valid device path: ")
	}
}

func IsRealPartition(name, disk string) bool {
	return strings.HasPrefix(name, disk) && len(name) > len(disk)
}
