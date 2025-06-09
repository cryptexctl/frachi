package main

import (
	"bufio"
	"fmt"
	"os"

	"frachi/src/utils"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Root privileges required")
		os.Exit(1)
	}

	if !utils.IsArchLinux() {
		fmt.Println("This installer is for Arch Linux only")
		os.Exit(1)
	}

	cfg, users, addSudo, addDoas, useSudo, useDoas, afterBase, addYay, useNetworkManager := utils.ParseArgs()
	utils.ConfirmConfig(cfg)

	parts, err := utils.ParsePartitions(cfg.Disk)
	if err != nil || len(parts) < 1 {
		fmt.Println("Failed to parse partitions or not enough partitions on disk.")
		os.Exit(1)
	}

	var realParts []utils.Partition
	for _, p := range parts {
		if utils.IsRealPartition(p.Name, cfg.Disk) {
			realParts = append(realParts, p)
		}
	}
	if len(realParts) == 0 {
		fmt.Println("No partitions found on disk.")
		os.Exit(1)
	}

	fmt.Println("Partitions detected:")
	for _, p := range realParts {
		fmt.Printf("%s\t%dMB\t%s\t%s\n", p.Name, p.SizeMB, p.FSType, p.Mount)
	}

	var efi, root, swap *utils.Partition
	if cfg.EFI != "" && cfg.Root != "" {
		efi = utils.FindPartition(realParts, cfg.EFI)
		root = utils.FindPartition(realParts, cfg.Root)
		if cfg.Swap != "" {
			swap = utils.FindPartition(realParts, cfg.Swap)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter EFI partition (device path, empty to skip): ")
		name := utils.ReadDevice(reader, realParts)
		if name != "" {
			efi = utils.FindPartition(realParts, name)
		}
		fmt.Print("Enter root partition (device path): ")
		name = utils.ReadDevice(reader, realParts)
		if name != "" {
			root = utils.FindPartition(realParts, name)
		}
		fmt.Print("Enter swap partition (device path, empty to skip): ")
		name = utils.ReadDevice(reader, realParts)
		if name != "" {
			swap = utils.FindPartition(realParts, name)
		}
	}

	if efi != nil {
		fmt.Printf("EFI: %s (%dMB)\n", efi.Name, efi.SizeMB)
		utils.ConfirmAndFormat(efi, "vfat")
	}
	if root != nil {
		fmt.Printf("Root: %s (%dMB)\n", root.Name, root.SizeMB)
		utils.ConfirmAndFormat(root, "ext4")
	}
	if swap != nil {
		fmt.Printf("Swap: %s (%dMB)\n", swap.Name, swap.SizeMB)
		utils.ConfirmAndFormat(swap, "swap")
	}

	sel := utils.PartitionSelection{EFI: efi, Root: root, Swap: swap}
	utils.MountDiskWithEfiAndSwap(sel)

	if !afterBase {
		utils.InstallBase()
		utils.DetectAndInstallDrivers()
	}
	utils.ConfigureSystemExt(cfg, users, addSudo, addDoas, useSudo, useDoas, useNetworkManager)
	utils.InstallBootloader(sel.Root.Name)
	if addYay {
		username := ""
		if len(users) > 0 {
			username = users[0].Name
		}
		utils.InstallYay(username)
	}
	utils.FinalMessage(cfg)
}
