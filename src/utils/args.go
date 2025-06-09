package utils

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type multiFlag []string

func (m *multiFlag) String() string       { return strings.Join(*m, ",") }
func (m *multiFlag) Set(val string) error { *m = append(*m, val); return nil }

func ParseArgs() (InstallConfig, []UserSpec, []string, []string, bool, bool, bool) {
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

	cfg := InstallConfig{
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
	var userSpecs []UserSpec
	for _, u := range users {
		parts := strings.SplitN(u, ":", 2)
		if len(parts) == 2 {
			userSpecs = append(userSpecs, UserSpec{Name: parts[0], Pass: parts[1]})
		}
	}
	return cfg, userSpecs, addSudo, addDoas, useSudo, useDoas, afterBase
}
