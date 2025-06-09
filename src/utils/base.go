package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func InstallBase() {
	fmt.Println("Installing base system...")
	cmd := exec.Command("pacstrap", "/mnt", "base", "linux", "linux-firmware")
	if err := cmd.Run(); err != nil {
		fmt.Println("pacstrap failed:", err)
		os.Exit(1)
	}
}
