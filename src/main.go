package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"frachi/src/utils"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Root privileges required")
		os.Exit(1)
	}

	if !isArchLinux() {
		fmt.Println("This installer is for Arch Linux only")
		os.Exit(1)
	}

	cfg, users, addSudo, addDoas, useSudo, useDoas, afterBase := parseArgs()
	confirmConfig(cfg)

	parts, err := utils.ParsePartitions(cfg.Disk)
	if err != nil || len(parts) < 1 {
		fmt.Println("Failed to parse partitions or not enough partitions on disk.")
		os.Exit(1)
	}

	var realParts []utils.Partition
	for _, p := range parts {
		if isRealPartition(p.Name, cfg.Disk) {
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
		efi = findPartition(realParts, cfg.EFI)
		root = findPartition(realParts, cfg.Root)
		if cfg.Swap != "" {
			swap = findPartition(realParts, cfg.Swap)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter EFI partition (device path, empty to skip): ")
		name := readDevice(reader, realParts)
		if name != "" {
			efi = findPartition(realParts, name)
		}
		fmt.Print("Enter root partition (device path): ")
		name = readDevice(reader, realParts)
		if name != "" {
			root = findPartition(realParts, name)
		}
		fmt.Print("Enter swap partition (device path, empty to skip): ")
		name = readDevice(reader, realParts)
		if name != "" {
			swap = findPartition(realParts, name)
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
	utils.ConfigureSystemExt(cfg, users, addSudo, addDoas, useSudo, useDoas)
	utils.InstallBootloader(sel.Root.Name)
	finalMessage(cfg)
}

func isRealPartition(name, disk string) bool {
	return strings.HasPrefix(name, disk) && len(name) > len(disk)
}

func findPartition(parts []utils.Partition, name string) *utils.Partition {
	for i := range parts {
		if parts[i].Name == name {
			return &parts[i]
		}
	}
	return nil
}

func readDevice(reader *bufio.Reader, parts []utils.Partition) string {
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

func parseArgs() (utils.InstallConfig, []utils.UserSpec, []string, []string, bool, bool, bool) {
	disk := flag.String("disk", "", "Target disk (e.g. /dev/sda)")
	efi := flag.String("efi", "", "EFI partition (e.g. /dev/sda1)")
	root := flag.String("root", "", "Root partition (e.g. /dev/sda2)")
	swap := flag.String("swap", "", "Swap partition (e.g. /dev/sda3)")
	hostname := flag.String("hostname", "archlinux", "Hostname")
	password := flag.String("password", "", "Password")
	locale := flag.String("locale", "en_US.UTF-8", "Locale")
	timezone := flag.String("timezone", "UTC", "Timezone")
	bootloader := flag.String("bootloader", "grub", "Bootloader (grub)")

	var users multiFlag
	var addSudo multiFlag
	var addDoas multiFlag
	var useSudo bool
	var useDoas bool
	var afterBase bool
	flag.Var(&users, "user", "User in format user:pass (can be repeated)")
	flag.Var(&addSudo, "addsudo", "Add user to sudoers (can be repeated)")
	flag.Var(&addDoas, "adddoas", "Add user to doas.conf (can be repeated)")
	flag.BoolVar(&useSudo, "sudo", false, "Install and configure sudo")
	flag.BoolVar(&useDoas, "doas", false, "Install and configure doas")
	flag.BoolVar(&afterBase, "afterbase", false, "Skip base install and drivers (for re-config)")
	flag.Parse()

	if *disk == "" || *password == "" {
		fmt.Println("--disk and --password are required")
		os.Exit(1)
	}

	cfg := utils.InstallConfig{
		Disk:       *disk,
		EFI:        *efi,
		Root:       *root,
		Swap:       *swap,
		Hostname:   *hostname,
		Password:   *password,
		Locale:     *locale,
		Timezone:   *timezone,
		Bootloader: *bootloader,
	}
	var userSpecs []utils.UserSpec
	for _, u := range users {
		parts := strings.SplitN(u, ":", 2)
		if len(parts) == 2 {
			userSpecs = append(userSpecs, utils.UserSpec{Name: parts[0], Pass: parts[1]})
		}
	}
	return cfg, userSpecs, addSudo, addDoas, useSudo, useDoas, afterBase
}

type multiFlag []string

func (m *multiFlag) String() string       { return strings.Join(*m, ",") }
func (m *multiFlag) Set(val string) error { *m = append(*m, val); return nil }

func confirmConfig(cfg utils.InstallConfig) {
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

func finalMessage(cfg utils.InstallConfig) {
	fmt.Println("Installation complete! You can reboot now.")
}
