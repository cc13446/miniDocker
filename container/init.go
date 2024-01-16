package container

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 新建容器父进程
func NewParentProcess(tty bool, command string) *exec.Cmd {
	// 传入参数，执行：miniDocker init [command]
	args := []string{"init", command}
	// 在/proc/self/目录下路径是进程自己的环境
	// 其中的exe为进程自己的可执行文件
	cmd := exec.Command("/proc/self/exe", args...)
	// 命名空间隔离参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
