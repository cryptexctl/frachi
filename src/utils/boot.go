package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func InstallBootloader(disk string) {
	fmt.Println("Installing bootloader...")
	archChrootLog("/mnt", "pacman", "-S", "--noconfirm", "grub", "efibootmgr")
	if isUEFI() {
		fmt.Println("Detected UEFI system. Installing grub-efi...")
		archChrootLog("/mnt", "grub-install", "--target=x86_64-efi", "--efi-directory=/boot/efi", "--bootloader-id=GRUB")
	} else {
		fmt.Println("Detected BIOS/MBR system. Installing grub-pc...")
		archChrootLog("/mnt", "grub-install", "--target=i386-pc", disk)
	}
	archChrootLog("/mnt", "grub-mkconfig", "-o", "/boot/grub/grub.cfg")
}

func isUEFI() bool {
	_, err := os.Stat("/sys/firmware/efi/efivars")
	return err == nil
}

func archChrootLog(root string, args ...string) {
	fmt.Printf("==> Running: arch-chroot %s %s\n", root, strings.Join(args, " "))
	cmdArgs := append([]string{"arch-chroot", root}, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
