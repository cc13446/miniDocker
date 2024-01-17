package cgroups

import (
	"cc.com/miniDocker/cgroups/subsystems"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
)

type Manager struct {
	// Path 创建的cgroups目录相对于root cgroups目录的路径
	Path string
	// 资源配置
	Resource *subsystems.ResourceConfig
}

func NewCgroupsManager(path string) *Manager {
	return &Manager{
		Path: path,
	}
}

// Apply 将进程pid加入到这个cgroups中
func (c *Manager) Apply(pid int) error {
	log.Infof("Apply cgroups [%s] with pid [%d]", c.Path, pid)
	if subCgroupsPath, err := subsystems.GetCgroupsPath(c.Path, false); err == nil {
		if err := os.WriteFile(path.Join(subCgroupsPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroups tacks fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup path %s error: %v", c.Path, err)
	}
}

// Set 设置 cgroups 资源限制
func (c *Manager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsInstance {
		if err := subSysIns.Set(c.Path, res); err != nil {
			return err
		}
	}
	return nil
}

// Destroy 释放 cgroups
func (c *Manager) Destroy() error {
	log.Infof("Remove cgroups [%s]", c.Path)
	if subCgroupsPath, err := subsystems.GetCgroupsPath(c.Path, false); err == nil {
		return os.RemoveAll(subCgroupsPath)
	} else {
		return err
	}
}
