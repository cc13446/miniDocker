package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

// RunContainerInitProcess 启动容器进程
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("Command is %s", command)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 挂载 proc
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Fail to mount /proc, error is %v", err)
	}

	// 启动进程
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}

	// 卸载 proc
	if err := syscall.Unmount("proc", defaultMountFlags); err != nil {
		log.Errorf("Fail to unmount /proc, error is %v", err)
	}

	return nil
}
