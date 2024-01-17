package subsystems

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

// GetCgroupsPath 获取cgroups文件的路径
func GetCgroupsPath(cgroupsPath string, autoCreate bool) (string, error) {
	// 获取 cgroups 挂载目录
	cgroupsRoot, err := FindCgroupsMountPoint()
	if err != nil {
		return "", err
	}

	// 创建新的 cgroups 空间
	if _, err := os.Stat(path.Join(cgroupsRoot, cgroupsPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupsRoot, cgroupsPath), 0755); err != nil {
				return "", fmt.Errorf("create cgroup path error %v", err)
			}
		}
		return path.Join(cgroupsRoot, cgroupsPath), nil
	}
	return "", fmt.Errorf("get cgroup path error %v", err)
}

// FindCgroupsMountPoint 获取 cgroups 文件夹挂载地址
func FindCgroupsMountPoint() (string, error) {
	// 打开 mount 文件
	const mountPath = "/proc/self/mountinfo"
	f, err := os.Open(mountPath)
	if err != nil {
		return "", err
	}

	// 函数结束关闭文件
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Infof("Close [%s] fail, error is %v", mountPath, err)
		}
	}(f)

	// 扫描文件内容
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		// 找到 cgroups 挂载的那一行，返回路径
		if fields[len(fields)-2] == "cgroup2" {
			return fields[4], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("not found cgroups path")
}
