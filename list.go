package main

import (
	"cc.com/miniDocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"
)

// ListContainers 列出所有容器
func ListContainers() {
	infoPath := path.Join(fmt.Sprintf(container.DefaultInfoLocation, ""))
	files, err := ioutil.ReadDir(infoPath)
	if err != nil {
		log.Errorf("Read dir %s error %v", infoPath, err)
		return
	}

	// 读取所有的容器信息
	var containers []*container.Info
	for _, file := range files {
		if tmpContainer, err := getContainerInfo(file); err != nil {
			log.Errorf("Get container info error %v", err)
		} else {
			containers = append(containers, tmpContainer)
		}
	}

	// 输出容器信息
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n"); err != nil {
		log.Errorf("Print table head error: %v", err)
	}
	for _, item := range containers {
		if _, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime); err != nil {
			log.Errorf("Print container info error: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

// getContainerInfo 获取容器信息
func getContainerInfo(file os.FileInfo) (*container.Info, error) {
	containerName := file.Name()
	configFilePath := path.Join(fmt.Sprintf(container.DefaultInfoLocation, containerName), container.ConfigFileName)
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.Info
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}

	return &containerInfo, nil
}
