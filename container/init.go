package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const defaultProcMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
const defaultTmpfsMountFlags = syscall.MS_NOSUID | syscall.MS_STRICTATIME

// RunContainerInitProcess 启动容器进程
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("get user command error when init container, cmdArray is nil")
	}

	log.Infof("User command is %s", strings.Join(cmdArray, " "))

	setUpMount()

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}

	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}

	setUpUnmount()
	return nil
}

// readUserCommand 读取用户命令
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("Init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// setUpMount 挂载
func setUpMount() {
	// 修改容器 root 目录
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		log.Errorf("Fail to pivot root, error is %v", err)
	}

	// 挂载 proc
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultProcMountFlags), ""); err != nil {
		log.Errorf("Fail to mount proc to /proc, error is %v", err)
	}
	// 挂载 tmpfs
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", defaultTmpfsMountFlags, "mode=755"); err != nil {
		log.Errorf("Fail to mount tmpfs to /dev, error is %v", err)
	}
}

// setUpUnmount 卸载挂载
func setUpUnmount() {
	// 卸载 proc
	if err := syscall.Unmount("/proc", defaultProcMountFlags); err != nil {
		log.Errorf("Fail to unmount /proc, error is %v", err)
	}
	// 卸载 proc
	if err := syscall.Unmount("/dev", defaultTmpfsMountFlags); err != nil {
		log.Errorf("Fail to unmount /dev, error is %v", err)
	}
}

// pivotRoot 修改 root 目录
func pivotRoot(root string) error {

	// pivotRoot 调用要求新旧两个的 root 文件夹不能在同一个文件系统下，所以这里把 old_root 重新挂载一下
	// bind mount 就是把相同的内容换了一个挂载点
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root 到新的rootfs, 现在 old_root 是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root error: %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")

	// umount rootfs/.pivot_root
	// 也就是把 old_root 卸载
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
