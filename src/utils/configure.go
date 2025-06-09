package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func ConfigureSystem(cfg interface{}) {
	c := cfg.(struct {
		Disk       string
		Hostname   string
		Username   string
		Password   string
		Locale     string
		Timezone   string
		Bootloader string
	})
	fmt.Println("Configuring system...")
	cmd := exec.Command("genfstab", "-U", "/mnt")
	fstab, _ := cmd.Output()
	f, _ := os.Create("/mnt/etc/fstab")
	f.Write(fstab)
	f.Close()

	archChroot("/mnt", "ln", "-sf", "/usr/share/zoneinfo/"+c.Timezone, "/etc/localtime")
	archChroot("/mnt", "hwclock", "--systohc")
	archChroot("/mnt", "sed", "-i", "s/^#"+c.Locale+"/"+c.Locale+"/", "/etc/locale.gen")
	archChroot("/mnt", "locale-gen")
	archChroot("/mnt", "bash", "-c", "echo LANG="+c.Locale+">/etc/locale.conf")
	archChroot("/mnt", "bash", "-c", "echo "+c.Hostname+">/etc/hostname")
	archChroot("/mnt", "useradd", "-m", "-G", "wheel", c.Username)
	archChroot("/mnt", "bash", "-c", "echo '"+c.Username+":"+c.Password+"' | chpasswd")
	archChroot("/mnt", "bash", "-c", "echo 'root:"+c.Password+"' | chpasswd")
	archChroot("/mnt", "sed", "-i", "s/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) ALL/", "/etc/sudoers")
}

func archChroot(root string, args ...string) {
	cmdArgs := append([]string{"arch-chroot", root}, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
