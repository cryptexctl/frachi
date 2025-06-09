package utils

import (
	"io"
	"os"
	"os/exec"
)

func CopyResolvConf() {
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

func UpdateCATrust() {
	cmd := exec.Command("arch-chroot", "/mnt", "update-ca-trust")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
