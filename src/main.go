package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"frachi/src/utils"
)

type InstallConfig struct {
	Disk       string
	Hostname   string
	Username   string
	Password   string
	Locale     string
	Timezone   string
	Bootloader string
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

	cfg := parseArgs()
	confirmConfig(cfg)

	parts, err := utils.ParsePartitions(cfg.Disk)
	if err != nil || len(parts) < 2 {
		fmt.Println("Failed to parse partitions or not enough partitions on disk.")
		os.Exit(1)
	}
	fmt.Println("Partitions detected:")
	for _, p := range parts {
		fmt.Printf("%s	%dMB	%s	%s\n", p.Name, p.SizeMB, p.FSType, p.Mount)
	}
	efi, root := utils.SelectPartitions(parts)
	fmt.Printf("EFI candidate: %s (%dMB)\n", efi.Name, efi.SizeMB)
	fmt.Printf("Root candidate: %s (%dMB)\n", root.Name, root.SizeMB)
	utils.ConfirmAndFormat(efi, "vfat")
	utils.ConfirmAndFormat(root, "ext4")
	utils.MountDiskWithEfi(root, efi)

	utils.InstallBase()
	utils.DetectAndInstallDrivers()
	utils.ConfigureSystem(cfg)
	utils.InstallBootloader(root.Name)
	finalMessage(cfg)
}

func parseArgs() InstallConfig {
	disk := flag.String("disk", "", "Target disk (e.g. /dev/sda)")
	hostname := flag.String("hostname", "archlinux", "Hostname")
	username := flag.String("username", "user", "Username")
	password := flag.String("password", "", "Password")
	locale := flag.String("locale", "en_US.UTF-8", "Locale")
	timezone := flag.String("timezone", "UTC", "Timezone")
	bootloader := flag.String("bootloader", "grub", "Bootloader (grub)")
	flag.Parse()

	if *disk == "" || *password == "" {
		fmt.Println("--disk and --password are required")
		os.Exit(1)
	}

	return InstallConfig{
		Disk:       *disk,
		Hostname:   *hostname,
		Username:   *username,
		Password:   *password,
		Locale:     *locale,
		Timezone:   *timezone,
		Bootloader: *bootloader,
	}
}

func confirmConfig(cfg InstallConfig) {
	fmt.Println("Installation config:")
	fmt.Printf("Disk: %s\nHostname: %s\nUsername: %s\nLocale: %s\nTimezone: %s\nBootloader: %s\n", cfg.Disk, cfg.Hostname, cfg.Username, cfg.Locale, cfg.Timezone, cfg.Bootloader)
	fmt.Print("Continue? (y/N): ")
	var resp string
	fmt.Scanln(&resp)
	if strings.ToLower(resp) != "y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}
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

func finalMessage(cfg InstallConfig) {
	fmt.Println("Installation complete! You can reboot now.")
}
