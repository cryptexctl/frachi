package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

type Partition struct {
	Name   string
	SizeMB int
	FSType string
	Mount  string
}

type PartitionSelection struct {
	EFI  *Partition
	Root *Partition
	Swap *Partition
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
		name := cleanName(fields[0])
		if !hasDigit(name) {
			continue
		}
		name = "/dev/" + name
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
			Name:   name,
			SizeMB: size / (1024 * 1024),
			FSType: fs,
			Mount:  mnt,
		})
	}
	return parts, nil
}

func cleanName(s string) string {
	return strings.TrimLeftFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func ConfirmAndFormat(part *Partition, fstype string) {
	if part == nil {
		return
	}
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
			} else if fstype == "swap" {
				cmd = exec.Command("mkswap", part.Name)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}
}

func MountDiskWithEfiAndSwap(sel PartitionSelection) {
	if sel.Root != nil {
		fmt.Printf("Mounting %s to /mnt...\n", sel.Root.Name)
		cmd := exec.Command("mount", sel.Root.Name, "/mnt")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Failed to mount root partition:", err)
			os.Exit(1)
		}
	}
	if sel.EFI != nil {
		os.MkdirAll("/mnt/boot/efi", 0755)
		fmt.Printf("Mounting %s to /mnt/boot/efi...\n", sel.EFI.Name)
		cmd := exec.Command("mount", sel.EFI.Name, "/mnt/boot/efi")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Failed to mount EFI partition:", err)
			os.Exit(1)
		}
	}
	if sel.Swap != nil {
		fmt.Printf("Enabling swap on %s...\n", sel.Swap.Name)
		cmd := exec.Command("swapon", sel.Swap.Name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Failed to enable swap:", err)
		}
	}
}
