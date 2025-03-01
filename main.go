package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	fmt.Println("Starting an isolated environment...")

	// Create a new process with namespace isolation
	cmd := exec.Command("/proc/self/exe", "child") // Re-run itself as child
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set namespace isolation flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | // New PID namespace
			syscall.CLONE_NEWNS | // New Mount namespace
			syscall.CLONE_NEWUTS, // New UTS namespace (hostname)
	}

	// Run the isolated process
	if err := cmd.Run(); err != nil {
		fmt.Println("Failed to start isolated process:", err)
	}
}

// Child process function
func init() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("Inside the isolated process...")

		// Change hostname (UTS namespace)
		err := syscall.Sethostname([]byte("sandbox"))
		if err != nil {
			fmt.Println("Failed to set hostname:", err)
			os.Exit(1)
		}

		// Replace the current process with Bash
		err = syscall.Exec("/bin/bash", []string{"/bin/bash"}, os.Environ())
		if err != nil {
			fmt.Println("Failed to start shell:", err)
			os.Exit(1)
		}
	}
}
