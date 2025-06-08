package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	root := "/tmp/sandbox_root"

	// Chroot to isolate
	if err := syscall.Chroot(root); err != nil {
		log.Fatalf("Chroot failed: %v", err)
	}

	// Change working directory to /home
	if err := syscall.Chdir("/home"); err != nil {
		log.Fatalf("Chdir failed: %v", err)
	}

	// Set clean environment variables
	env := []string{
		"PS1=sandboxuser@sandboxed:\\w$ ",
		"HOME=/home",
		"PATH=/bin:/usr/bin:/usr/local/bin",
		"TERM=xterm",
		"USER=sandboxuser",
	}

	// Launch /bin/bash interactively
	cmd := exec.Command("/bin/bash")
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Starting sandboxed environment...")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run shell: %v", err)
	}
}
