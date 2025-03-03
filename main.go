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

	data, err := ReadYML(ymlfile) // Function is now accessible
	if err != nil {
		log.Fatal("failed to read yml file:", err)
	}

	fmt.Println("Mount Paths:", data.Mount)
	fmt.Println(" Starting sandboxed environment...")

	cmd := exec.Command("/proc/self/exe", "child")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | // New PID namespace
			syscall.CLONE_NEWNS | // New Mount namespace
			syscall.CLONE_NEWUTS | // New Hostname namespace
			syscall.CLONE_NEWIPC | // New IPC namespace
			syscall.CLONE_NEWUSER, // New User namespace
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1}, // Map user as root inside
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1}, // Map group as root inside
		},
	}

	if err := cmd.Run(); err != nil {
		fmt.Println(" Failed to start sandbox:", err)
	}
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "child" {
		fmt.Println("🔒 Inside the isolated process...")

		// Set a new hostname
		_ = syscall.Sethostname([]byte("sandboxed"))

		// Drop network access
		_ = syscall.Unshare(syscall.CLONE_NEWNET)

		// Create an isolated filesystem
		newRoot := "/tmp/sandbox_root"
		os.MkdirAll(newRoot, 0755)

		// Mount a tmpfs (RAM-based) filesystem to isolate changes
		err := syscall.Mount("tmpfs", newRoot, "tmpfs", 0, "")
		if err != nil {
			fmt.Println(" Failed to mount tmpfs:", err)
			os.Exit(1)
		}

		// Create necessary directories inside the new root
		os.MkdirAll(newRoot+"/bin", 0755)
		os.MkdirAll(newRoot+"/lib", 0755)
		os.MkdirAll(newRoot+"/lib64", 0755)
		os.MkdirAll(newRoot+"/usr/bin", 0755)
		os.MkdirAll(newRoot+"/usr/lib", 0755)
		os.MkdirAll(newRoot+"/usr/lib64", 0755)
		os.MkdirAll(newRoot+"/proc", 0755)
		os.MkdirAll(newRoot+"/home", 0755) // For abc.py

		// Bind-mount system directories
		syscall.Mount("/bin", newRoot+"/bin", "", syscall.MS_BIND, "")
		syscall.Mount("/lib", newRoot+"/lib", "", syscall.MS_BIND, "")
		syscall.Mount("/lib64", newRoot+"/lib64", "", syscall.MS_BIND, "")
		syscall.Mount("/usr/bin", newRoot+"/usr/bin", "", syscall.MS_BIND, "")
		syscall.Mount("/usr/lib", newRoot+"/usr/lib", "", syscall.MS_BIND, "")
		syscall.Mount("/usr/lib64", newRoot+"/usr/lib64", "", syscall.MS_BIND, "")
		syscall.Mount("/proc", newRoot+"/proc", "proc", 0, "")

		// Bind-mount abc.py from the current directory
		// currentDir, _ := os.Getwd() + "/abc.py"
		//	syscall.Mount(currentDir+"", newRoot+"/home/abc.py", "", syscall.MS_BIND, "")
		currentDir, _ := os.Getwd()
		abcPyPath := currentDir + "/abc.py"
		fmt.Println(abcPyPath)
		err = syscall.Mount("/usr/bin/python3", newRoot+"/usr/bin/python3", "", syscall.MS_BIND, "")
		if err != nil {
			fmt.Println("Failed to mount Python:", err)
			os.Exit(1)
		}
		// err = syscall.Mount(abcPyPath, newRoot+"/abc.py", "", syscall.MS_BIND, "")
		// if err != nil {
		// 	fmt.Println("Failed to mount file:", err)
		// 	os.Exit(1)
		// }

		// Change root to the new isolated filesystem
		if err := syscall.Chroot(newRoot); err != nil {
			fmt.Println("Failed to change root:", err)
			os.Exit(1)
		}
		os.Chdir("/") // Ensure we are inside the new root

		// Run a shell in the sandbox
		err = syscall.Exec("/bin/sh", []string{"sh"}, os.Environ())
		if err != nil {
			fmt.Println(" Failed to start shell:", err)
			os.Exit(1)
		}
	}
}
