package main

import (
	"cc.com/miniDocker/cgroups"
	"cc.com/miniDocker/cgroups/subsystems"
	"cc.com/miniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Run 运行容器
func Run(tty bool, commandArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Error start parent process, error is %v", err)
	}
	// 新建 cgroups
	cgroupsManager := cgroups.NewCgroupsManager("miniDocker-cgroups")

	defer func(cgroupsManager *cgroups.Manager) {
		if err := cgroupsManager.Destroy(); err != nil {
			log.Errorf("Destroy cgroups fail, error is : %v", err)
		}
	}(cgroupsManager)

	if err := cgroupsManager.Set(res); err != nil {
		log.Fatal("Set cgroups config fail, error is %v", err)
	}

	if err := cgroupsManager.Apply(parent.Process.Pid); err != nil {
		log.Fatal("Apply cgroups config fail, error is %v", err)
	}

	sendInitCommand(commandArray, writePipe)

	if err := parent.Wait(); err != nil {
		log.Errorf("Error wait parent process, error is %v", err)
	}
	os.Exit(-1)
}

// sendInitCommand 向子进程发送命令
func sendInitCommand(commandArray []string, writePipe *os.File) {
	command := strings.Join(commandArray, " ")
	log.Infof("User command is %s", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Fatalf("Write user command to child process failed, error is : %v", err)
	}
	if err := writePipe.Close(); err != nil {
		log.Errorf("Close pipe failed, error is : %v", err)
	}
}
