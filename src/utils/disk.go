package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Partition struct {
	Name   string
	SizeMB int
	FSType string
	Mount  string
}

func ParsePartitions(disk string) ([]Partition, error) {
	cmd := exec.Command("lsblk", "-b", "-o", "NAME,SIZE,FSTYPE,MOUNTPOINT", disk)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var parts []Partition
	s := bufio.NewScanner(strings.NewReader(string(out)))
	s.Scan() // skip header
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		size, _ := strconv.Atoi(fields[1])
		fs := ""
		mnt := ""
		if len(fields) > 2 {
			fs = fields[2]
		}
		if len(fields) > 3 {
			mnt = fields[3]
		}
		parts = append(parts, Partition{
			Name:   "/dev/" + name,
			SizeMB: size / (1024 * 1024),
			FSType: fs,
			Mount:  mnt,
		})
	}
	return parts, nil
}

func SelectPartitions(parts []Partition) (efi, root Partition) {
	var min *Partition
	var max *Partition
	for i, p := range parts {
		if min == nil || p.SizeMB < min.SizeMB {
			min = &parts[i]
		}
		if max == nil || p.SizeMB > max.SizeMB {
			max = &parts[i]
		}
	}
	return *min, *max
}

func ConfirmAndFormat(part Partition, fstype string) {
	if part.FSType != fstype {
		fmt.Printf("Partition %s is not %s, format? (y/N): ", part.Name, fstype)
		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(resp) == "y" {
			var cmd *exec.Cmd
			if fstype == "vfat" {
				cmd = exec.Command("mkfs.fat", "-F32", part.Name)
			} else if fstype == "ext4" {
				cmd = exec.Command("mkfs.ext4", part.Name)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
}

func MountDiskWithEfi(root, efi Partition) {
	fmt.Printf("Mounting %s to /mnt...\n", root.Name)
	cmd := exec.Command("mount", root.Name, "/mnt")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to mount root partition:", err)
		os.Exit(1)
	}
	os.MkdirAll("/mnt/boot/efi", 0755)
	fmt.Printf("Mounting %s to /mnt/boot/efi...\n", efi.Name)
	cmd = exec.Command("mount", efi.Name, "/mnt/boot/efi")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to mount EFI partition:", err)
		os.Exit(1)
	}
}
