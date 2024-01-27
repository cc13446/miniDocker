package main

import (
	"cc.com/miniDocker/cgroups"
	"cc.com/miniDocker/cgroups/subsystems"
	"cc.com/miniDocker/container"
	"cc.com/miniDocker/network"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Run 运行容器
func Run(tty bool, commandArray []string, res *subsystems.ResourceConfig, volume, containerName, imageName string, envSlice []string, nw string, portMapping []string) {
	containerId := randStringBytes(10)
	if containerName == "" {
		containerName = containerId
	}

	parent, writePipe := container.NewParentProcess(tty, volume, containerName, imageName, envSlice)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Error start parent process, error is %v", err)
	}

	// record container info
	if err := recordContainerInfo(parent.Process.Pid, commandArray, containerName, containerId, volume); err != nil {
		log.Errorf("Record container info error %v", err)
		return
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

	if nw != "" {
		// config container network
		if err := network.Init(); err != nil {
			log.Errorf("Network init error %v", err)
		}
		containerInfo := &container.Info{
			Id:          containerId,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}

	sendInitCommand(commandArray, writePipe)

	if tty {
		if err := parent.Wait(); err != nil {
			log.Errorf("Error wait parent process, error is %v", err)
		}
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}
	os.Exit(-1)
}

// sendInitCommand 向子进程发送命令
func sendInitCommand(commandArray []string, writePipe *os.File) {
	command := strings.Join(commandArray, " ")
	log.Infof("User command is %s, send to child", command)
	if _, err := writePipe.WriteString(command); err != nil {
		log.Fatalf("Write user command to child process failed, error is : %v", err)
	}
	if err := writePipe.Close(); err != nil {
		log.Errorf("Close pipe failed, error is : %v", err)
	}
}

// recordContainerInfo 记录容器信息
func recordContainerInfo(containerPID int, commandArray []string, containerName, containerId, volume string) error {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	// 生成容器信息
	containerInfo := &container.Info{
		Id:          containerId,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Parse container info to json error %v", err)
		return err
	}
	jsonStr := string(jsonBytes)

	// 生成容器信息文件
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(infoPath, 0622); err != nil {
		log.Errorf("Mkdir path %s error %v", infoPath, err)
		return err
	}
	fileName := path.Join(infoPath, container.ConfigFileName)
	file, err := os.Create(fileName)
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Errorf("Close file %s error %v", fileName, err)
		}
	}(file)
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return err
	}

	// 记录容器信息到文件
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return err
	}

	return nil
}

func deleteContainerInfo(containerName string) {
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(infoPath); err != nil {
		log.Errorf("Remove dir %s error %v", infoPath, err)
	}
}

// randStringBytes 生成16进制随机字符串
func randStringBytes(n int) string {
	letterBytes := "1234567890abcd"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
