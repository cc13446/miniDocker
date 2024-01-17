package cgroups

import (
	"cc.com/miniDocker/cgroups/subsystems"
	"os"
	"path"
	"testing"
)

func TestCgroups(t *testing.T) {
	resConfig := subsystems.ResourceConfig{
		CpuMax:    "1000",
		CpuSet:    "1",
		MemoryMax: "1000m",
	}
	testCgroups := "test_cgroups"
	manager := NewCgroupsManager(testCgroups)

	if err := manager.Set(&resConfig); err != nil {
		t.Fatalf("cgroup set fail %v", err)
	}

	mountPoint, err := subsystems.FindCgroupsMountPoint()
	if err != nil {
		t.Fatalf("get cgroups mount point fail %v", err)
	}

	stat, _ := os.Stat(path.Join(mountPoint, testCgroups))
	t.Logf("cgroup stats: %+v", stat)

	if err := manager.Apply(os.Getpid()); err != nil {
		t.Fatalf("cgroup apply fail %v", err)
	}

	// 将进程移回到根Cgroups节点
	defaultManager := NewCgroupsManager("")
	if err := defaultManager.Apply(os.Getpid()); err != nil {
		t.Fatalf("cgroup apply fail %v", err)
	}

	if err := manager.Destroy(); err != nil {
		t.Fatalf("cgroup remove fail %v", err)
	}
}
