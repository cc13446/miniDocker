package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

// RunContainerInitProcess 启动容器进程
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("Command is %s", command)

	// NOEXEC 不允许运行其他程序
	// NOSUID 不允许 set user id 或者 set group id
	// NODEV  不允许访问设备
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载/proc
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Fail to mount /proc, error is %v", err)
	}
	argv := []string{command}
	// 启动进程
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
