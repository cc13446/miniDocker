package main

import (
	"cc.com/miniDocker/container"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

const EnvExecPid = "mini_docker_pid"
const EnvExecCmd = "mini_docker_cmd"

func ExecContainer(containerName string, comArray []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}
	cmdStr := strings.Join(comArray, " ")
	log.Infof("Container pid %s", pid)
	log.Infof("Command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = os.Setenv(EnvExecPid, pid)
	if err != nil {
		log.Errorf("Set env %s error: %v", EnvExecPid, err)
	}
	err = os.Setenv(EnvExecCmd, cmdStr)
	if err != nil {
		log.Errorf("Set env %s error: %v", EnvExecCmd, err)
	}

	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}

// getContainerPidByName 通过容器名字获取pid
func getContainerPidByName(containerName string) (string, error) {
	infoPath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := path.Join(infoPath, container.ConfigFileName)
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.Info
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
