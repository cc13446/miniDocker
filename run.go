package main

import (
	"cc.com/miniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

// Run 运行容器
func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Errorf("Error start parent process, error is %v", err)
	}
	if err := parent.Wait(); err != nil {
		log.Errorf("Error wait parent process, error is %v", err)
	}
	os.Exit(-1)
}
