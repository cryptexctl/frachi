package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func MountDisk(disk string) {
	fmt.Printf("Mounting %s to /mnt...\n", disk)
	cmd := exec.Command("mount", disk, "/mnt")
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to mount disk:", err)
		os.Exit(1)
	}
}
