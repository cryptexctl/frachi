package utils

import (
	"fmt"
	"os"
	"os/exec"
)

type InstallConfig struct {
	Disk       string
	EFI        string
	Root       string
	Swap       string
	Hostname   string
	Username   string
	Password   string
	Locale     string
	Timezone   string
	Bootloader string
}

func ConfigureSystem(cfg InstallConfig) {
	fmt.Println("Configuring system...")
	cmd := exec.Command("genfstab", "-U", "/mnt")
	fstab, _ := cmd.Output()
	f, _ := os.Create("/mnt/etc/fstab")
	f.Write(fstab)
	f.Close()

	archChroot("/mnt", "ln", "-sf", "/usr/share/zoneinfo/"+cfg.Timezone, "/etc/localtime")
	archChroot("/mnt", "hwclock", "--systohc")
	archChroot("/mnt", "sed", "-i", "s/^#"+cfg.Locale+"/"+cfg.Locale+"/", "/etc/locale.gen")
	archChroot("/mnt", "locale-gen")
	archChroot("/mnt", "bash", "-c", "echo LANG="+cfg.Locale+">/etc/locale.conf")
	archChroot("/mnt", "bash", "-c", "echo "+cfg.Hostname+">/etc/hostname")
	archChroot("/mnt", "useradd", "-m", "-G", "wheel", cfg.Username)
	archChroot("/mnt", "bash", "-c", "echo '"+cfg.Username+":"+cfg.Password+"' | chpasswd")
	archChroot("/mnt", "bash", "-c", "echo 'root:"+cfg.Password+"' | chpasswd")
	archChroot("/mnt", "sed", "-i", "s/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) ALL/", "/etc/sudoers")
}

func archChroot(root string, args ...string) {
	cmdArgs := append([]string{"arch-chroot", root}, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
