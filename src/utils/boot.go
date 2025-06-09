package utils

import (
	"fmt"
	"os/exec"
)

func InstallBootloader(disk string) {
	fmt.Println("Installing bootloader...")
	cmd1 := exec.Command("arch-chroot", "/mnt", "pacman", "-S", "--noconfirm", "grub")
	cmd1.Run()
	cmd2 := exec.Command("arch-chroot", "/mnt", "grub-install", "--target=i386-pc", disk)
	cmd2.Run()
	cmd3 := exec.Command("arch-chroot", "/mnt", "grub-mkconfig", "-o", "/boot/grub/grub.cfg")
	cmd3.Run()
}
