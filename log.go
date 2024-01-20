package main

import (
	"cc.com/miniDocker/container"
	"fmt"
	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

// logContainer 打印容器日志
func logContainer(containerName string) {
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFilePath := path.Join(infoPath, container.LogFileName)
	file, err := os.Open(logFilePath)
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Errorf("Close file %s error %v", logFilePath, err)
		}
	}(file)
	if err != nil {
		log.Errorf("Log container open file %s error %v", logFilePath, err)
		return
	}
	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}
