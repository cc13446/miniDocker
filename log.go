package main

import (
	"cc.com/miniDocker/container"
	"fmt"
	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
	"path"
)

// logContainer 打印容器日志
func logContainer(containerName string) {
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFilePath := path.Join(infoPath, container.LogFileName)

	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	if err != nil {
		log.Errorf("Tail file %s error %v", logFilePath, err)
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}
