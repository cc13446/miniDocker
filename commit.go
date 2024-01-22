package main

import (
	"cc.com/miniDocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
)

func commitContainer(containerName, imageName string) {
	imageTar := path.Join(container.ImagePath, imageName+".tar")
	if exist, err := container.PathExists(container.ImagePath); err != nil || !exist {
		if err := os.MkdirAll(container.ImagePath, 0777); err != nil {
			log.Errorf("Mkdir %s fail", imageName)
		}
	}

	log.Infof("Commit container imageTar: %s", imageTar)
	cmd := exec.Command("tar", "-czf", imageTar, "-C", fmt.Sprintf(container.MergedPath, containerName), ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Tar folder %s error %v", fmt.Sprintf(container.MergedPath, containerName), err)
	}
}
