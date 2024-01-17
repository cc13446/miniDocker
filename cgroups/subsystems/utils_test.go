package subsystems

import (
	"testing"
)

func TestFindCgroupsMountPoint(t *testing.T) {

	cgroupsPath, err := FindCgroupsMountPoint()
	if err != nil {
		t.Fatalf("Fail find cgroups mount point %v", err)
	}
	t.Logf("Cgroups mount point %v\n", cgroupsPath)
}
