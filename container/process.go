package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"syscall"
)

const RootPath = "/var/run/miniDocker/root/"
const LowerPath = RootPath + "%s/lower/"
const UpperPath = RootPath + "%s/upper/"
const WorkPath = RootPath + "%s/work/"
const MergedPath = RootPath + "%s/merged/"
const ImagePath = "/var/run/miniDocker/image/"

// NewParentProcess 新建容器父进程
func NewParentProcess(tty bool, volume string, containerName, imageName string) (*exec.Cmd, *os.File) {
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
	} else {
		infoPath := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(infoPath, 0622); err != nil {
			log.Errorf("NewParentProcess mkdir %s error %v", infoPath, err)
			return nil, nil
		}
		stdLogFilePath := path.Join(infoPath, LogFileName)
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	NewWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(MergedPath, containerName)
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
func NewWorkSpace(volume, imageName, containerName string) {
	CreateLower(containerName, imageName)
	CreateDirs(containerName)
	MountOverlayFS(containerName)
	if volume != "" {
		hostPath, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("Extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mountVolume(fmt.Sprintf(MergedPath, containerName), hostPath, containerPath)
	}
}

func CreateLower(containerName, imageName string) {
	tarPath := path.Join(ImagePath, fmt.Sprintf("%s.tar", imageName))
	unTarPath := fmt.Sprintf(LowerPath, containerName)
	exist, err := PathExists(unTarPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", unTarPath, err)
	}
	// 解压tar包
	if exist == false {
		if err := os.MkdirAll(unTarPath, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", unTarPath, err)
		}
		if _, err := exec.Command("tar", "-xvf", tarPath, "-C", unTarPath).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error %v", unTarPath, err)
		}
	}
}

func CreateDirs(containerName string) {
	dirs := []string{
		fmt.Sprintf(MergedPath, containerName),
		fmt.Sprintf(UpperPath, containerName),
		fmt.Sprintf(WorkPath, containerName),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", dir, err)
		}
	}
}

func MountOverlayFS(containerName string) {
	// 拼接参数
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		fmt.Sprintf(LowerPath, containerName),
		fmt.Sprintf(UpperPath, containerName),
		fmt.Sprintf(WorkPath, containerName))

	// 拼接完整命令
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, fmt.Sprintf(MergedPath, containerName))
	log.Infof("Mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount overlayfs error: %v", err)
	}
}

// DeleteWorkSpace Delete the Overlay2 filesystem while container exit
func DeleteWorkSpace(volume, containerName string) {

	// 如果指定了 volume 则需要 umount volume
	// NOTE: 一定要要先 umount volume ，然后再删除目录，否则由于 bind mount 存在，删除临时目录会导致 volume 目录中的数据丢失。
	if volume != "" {
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
		} else {
			umountVolume(fmt.Sprintf(MergedPath, containerName), containerPath)
		}
	}
	UmountOverlayFS(containerName)
	DeleteDirs(containerName)
}

func UmountOverlayFS(containerName string) {
	containerMergePath := fmt.Sprintf(MergedPath, containerName)
	cmd := exec.Command("umount", containerMergePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Run unmount %s error: %v", containerMergePath, err)
	}
	if err := os.RemoveAll(containerMergePath); err != nil {
		log.Errorf("Remove dir %s error %v", containerMergePath, err)
	}
}

func DeleteDirs(containerName string) {
	dirs := []string{
		fmt.Sprintf(MergedPath, containerName),
		fmt.Sprintf(UpperPath, containerName),
		fmt.Sprintf(WorkPath, containerName),
		fmt.Sprintf(LowerPath, containerName),
		path.Join(RootPath, containerName),
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
