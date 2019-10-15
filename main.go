package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker         run image <cmd> <args>
// go run main.go run       <cmd> <args>
func main()  {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cmd.Run()
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cg()
	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("./ubuntu-1804-base/"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	must(syscall.Unmount("/proc", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids")
	memory := filepath.Join(cgroups, "memory")
	panicOnCreateDirError(os.Mkdir(filepath.Join(pids, "prakash"), 0755))
	panicOnCreateDirError(os.Mkdir(filepath.Join(memory, "prakash"), 0755))


	must(ioutil.WriteFile(filepath.Join(pids, "prakash/pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "prakash/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "prakash/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

	must(ioutil.WriteFile(filepath.Join(memory, "prakash/memory.limit_in_bytes"), []byte("41943040"), 0700))
	must(ioutil.WriteFile(filepath.Join(memory, "prakash/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(memory, "prakash/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

}

func panicOnCreateDirError(err error) {
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func must(err error)  {
	if err != nil {
		panic(err)
	}
}