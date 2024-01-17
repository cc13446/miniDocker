package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

// ResourceConfig 用于传递资源限制的结构体
type ResourceConfig struct {
	MemoryMax string
	CpuMax    string
	CpuSet    string
}

// Subsystem 子系统接口，这里将cgroups抽象为path
type Subsystem interface {

	// Set 设置某个cgroups在这个子系统的资源限制
	Set(path string, res *ResourceConfig) error
}

// resourceConfigGetFunc 获取资源配置函数
type resourceConfigGetFunc func(config *ResourceConfig) string

// SimpleSubSystem 子系统接口实现类
type SimpleSubSystem struct {
	name          string
	fileName      string
	configGetFunc resourceConfigGetFunc
}

func (s *SimpleSubSystem) Set(cgroupsPath string, res *ResourceConfig) error {
	log.Infof("Set cgroups [%s] type [%s] with [%s]", cgroupsPath, s.name, s.configGetFunc(res))
	if subCgroupsPath, err := GetCgroupsPath(cgroupsPath, true); err == nil {
		if s.configGetFunc(res) != "" {
			if err := os.WriteFile(path.Join(subCgroupsPath, s.fileName), []byte(s.configGetFunc(res)), 0644); err != nil {
				return fmt.Errorf("set cgroups fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

var SubsystemsInstance = []Subsystem{
	&SimpleSubSystem{name: "cpu", fileName: "cpu.max", configGetFunc: func(config *ResourceConfig) string {
		return config.CpuMax
	}},
	&SimpleSubSystem{name: "cpuset", fileName: "cpuset.cpus", configGetFunc: func(config *ResourceConfig) string {
		return config.CpuSet
	}},
	&SimpleSubSystem{name: "memory", fileName: "memory.max", configGetFunc: func(config *ResourceConfig) string {
		return config.MemoryMax
	}},
}
