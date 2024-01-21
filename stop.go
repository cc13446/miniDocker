package main

import (
	"cc.com/miniDocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"syscall"
)

func stopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Get contaienr pid by name %s error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerName, err)
		return
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerName, err)
		return
	}
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := path.Join(infoPath, container.ConfigFileName)
	if err := os.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error: %v", configFilePath, err)
	}
}

// getContainerInfoByName 通过容器名字获取容器信息
func getContainerInfoByName(containerName string) (*container.Info, error) {
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := path.Join(infoPath, container.ConfigFileName)
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.Info
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		log.Errorf("GetContainerInfoByName unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
