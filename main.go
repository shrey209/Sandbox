package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	ymlfile := "./config.yml"

	data, err := ReadYML(ymlfile)
	if err != nil {
		log.Fatal("failed to read yml file:", err)
	}

	fmt.Println("Mount Paths:", data.Mount)
	fmt.Println("Starting sandboxed environment...")

	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to start sandbox:", err)
	}
}
func init() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("Inside the isolated process...")

		_ = syscall.Sethostname([]byte("sandboxed"))
		_ = syscall.Unshare(syscall.CLONE_NEWNET)

		newRoot := "/tmp/sandbox_root"
		os.MkdirAll(newRoot, 0755)

		// Mount tmpfs (ephemeral in-memory root)
		if err := syscall.Mount("tmpfs", newRoot, "tmpfs", 0, ""); err != nil {
			log.Fatal("Failed to mount tmpfs:", err)
		}

		// Create necessary folders
		dirs := []string{"/bin", "/lib", "/lib64", "/usr/bin", "/usr/lib", "/usr/lib64", "/proc", "/home"}
		for _, d := range dirs {
			if err := os.MkdirAll(newRoot+d, 0755); err != nil {
				log.Fatal("Failed to mkdir:", err)
			}
		}

		// Bind-mount required system folders
		binds := []string{"/bin", "/lib", "/lib64", "/usr/bin", "/usr/lib", "/usr/lib64"}
		for _, src := range binds {
			if err := syscall.Mount(src, newRoot+src, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
				log.Fatal("Bind mount failed for", src, ":", err)
			}
		}

		// Mount /proc
		if err := syscall.Mount("proc", newRoot+"/proc", "proc", 0, ""); err != nil {
			log.Fatal("Failed to mount /proc:", err)
		}

		// 🔄 COPY project folder into the sandbox (instead of bind-mounting)
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Failed to get cwd:", err)
		}
		sandboxHome := newRoot + "/home"

		// Copy folder using 'cp -a' (simple, reliable)
		copyCmd := exec.Command("cp", "-a", cwd+"/.", sandboxHome)
		copyCmd.Stderr = os.Stderr
		copyCmd.Stdout = os.Stdout
		if err := copyCmd.Run(); err != nil {
			log.Fatal("Failed to copy project into sandbox:", err)
		}

		// Chroot
		if err := syscall.Chroot(newRoot); err != nil {
			log.Fatal("Failed to chroot:", err)
		}
		if err := syscall.Chdir("/home"); err != nil {
			log.Fatal("Failed to chdir:", err)
		}

		// Launch bash
		if err := syscall.Exec("/bin/bash", []string{"bash"}, os.Environ()); err != nil {
			log.Fatal("Failed to exec bash:", err)
		}
	}
}
