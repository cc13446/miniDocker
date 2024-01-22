package main

import (
	"cc.com/miniDocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP {
		log.Errorf("Couldn't remove running container")
		return
	}
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(infoPath); err != nil {
		log.Errorf("Remove file %s error %v", infoPath, err)
		return
	}
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
}
