package utils

import (
	"fmt"
	"io"
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

type UserSpec struct {
	Name string
	Pass string
}

func ConfigureSystemExt(cfg InstallConfig, users []UserSpec, addSudo, addDoas []string, useSudo, useDoas, useNetworkManager bool) {
	fmt.Println("Configuring system...")

	copyResolvConf()
	archChroot("/mnt", "update-ca-trust")

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

	// fuck your systemd-{resolved,networkd}. connman is better.
	archChroot("/mnt", "systemctl", "enable", "systemd-networkd")
	archChroot("/mnt", "systemctl", "enable", "systemd-resolved")

	if useNetworkManager {
		archChroot("/mnt", "pacman", "-S", "--noconfirm", "networkmanager")
		archChroot("/mnt", "systemctl", "enable", "NetworkManager")
	}

	archChroot("/mnt", "bash", "-c", "echo 'root:"+cfg.Password+"' | chpasswd")

	for _, u := range users {
		archChroot("/mnt", "useradd", "-m", u.Name)
		archChroot("/mnt", "bash", "-c", "echo '"+u.Name+":"+u.Pass+"' | chpasswd")
	}

	if useSudo {
		archChroot("/mnt", "pacman", "-S", "--noconfirm", "sudo")
		for _, u := range addSudo {
			archChroot("/mnt", "usermod", "-aG", "wheel", u)
		}
		archChroot("/mnt", "sed", "-i", "s/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) ALL/", "/etc/sudoers")
	}

	if useDoas {
		archChroot("/mnt", "pacman", "-S", "--noconfirm", "opendoas")
		doasConf := "/mnt/etc/doas.conf"
		f, _ := os.Create(doasConf)
		for _, u := range addDoas {
			f.WriteString("permit persist " + u + " as root\n")
		}
		f.Close()
	}
}

func copyResolvConf() {
	src := "/etc/resolv.conf"
	dst := "/mnt/etc/resolv.conf"
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, in)
}

func archChroot(root string, args ...string) {
	cmdArgs := append([]string{"arch-chroot", root}, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func InstallYay(username string) {
	fmt.Println("Installing yay AUR helper...")
	archChroot("/mnt", "pacman", "-S", "--noconfirm", "base-devel", "git")
	archChroot("/mnt", "mkdir", "-p", "/tmp/yay")
	archChroot("/mnt", "bash", "-c", "cd /tmp && git clone https://aur.archlinux.org/yay.git")
	if username != "" {
		archChroot("/mnt", "bash", "-c", "cd /tmp/yay && sudo -u "+username+" makepkg -si --noconfirm")
	} else {
		archChroot("/mnt", "bash", "-c", "cd /tmp/yay && makepkg -si --noconfirm")
	}
}
