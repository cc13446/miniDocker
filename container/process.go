package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const RootPath = "./"
const LowerPath = RootPath + "busybox"
const UpperPath = RootPath + "upper"
const WorkPath = RootPath + "work"
const MergedPath = RootPath + "merged"

// NewParentProcess 新建容器父进程
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	// 管道
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}

	// 传入参数，执行：miniDocker init [command]
	// 在/proc/self/目录下路径是进程自己的环境
	// 其中的exe为进程自己的可执行文件
	cmd := exec.Command("/proc/self/exe", "init")
	// 命名空间隔离参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	NewWorkSpace(RootPath)
	cmd.Dir = MergedPath
	return cmd, writePipe
}

// NewPipe 生成管道
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

// NewWorkSpace Create an Overlay2 filesystem as container root workspace
func NewWorkSpace(rootPath string) {
	CreateLower(rootPath)
	CreateDirs()
	MountOverlayFS()
}

func CreateLower(rootPath string) {
	busyboxTarURL := path.Join(rootPath, "busybox.tar")
	exist, err := PathExists(LowerPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", LowerPath, err)
	}
	// 解压tar包
	if exist == false {
		if err := os.Mkdir(LowerPath, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", LowerPath, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", LowerPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", LowerPath, err)
		}
	}
}

func CreateDirs() {
	dirs := []string{
		MergedPath, UpperPath, WorkPath,
	}

	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", dir, err)
		}
	}
}

func MountOverlayFS() {
	// 拼接参数
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		LowerPath,
		UpperPath,
		WorkPath)

	// 拼接完整命令
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, MergedPath)
	log.Infof("Mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount overlayfs error: %v", err)
	}
}

// DeleteWorkSpace Delete the Overlay2 filesystem while container exit
func DeleteWorkSpace() {
	UmountOverlayFS()
	DeleteDirs()
}

func UmountOverlayFS() {
	cmd := exec.Command("umount", MergedPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Run unmount %s error: %v", MergedPath, err)
	}
	if err := os.RemoveAll(MergedPath); err != nil {
		log.Errorf("Remove dir %s error %v", MergedPath, err)
	}
}

func DeleteDirs() {
	dirs := []string{
		MergedPath, UpperPath, WorkPath,
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}

// PathExists 路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
