package main

import (
	"cc.com/miniDocker/cgroups/subsystems"
	"cc.com/miniDocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

const usage = `MiniDocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works.
Enjoy it, just for fun.
`

const ttyFlagName = "ti"
const detachFlagName = "d"
const memoryMaxFlagName = "memMax"
const cpuMaxFlagName = "cpuMax"
const cpuSetFlagName = "cpuSet"
const volumeFlagName = "v"
const containerNameFlagName = "name"

func main() {

	app := cli.NewApp()
	app.Name = "miniDocker"
	app.Usage = usage

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	initCommand := cli.Command{
		Name:  "init",
		Usage: "Init container process and run user's process in container. Do not call it outside.",
		Action: func(context *cli.Context) error {
			log.Infof("Begin init")
			// run init process
			return container.RunContainerInitProcess()
		},
	}

	runCommand := cli.Command{
		Name:  "run",
		Usage: "Create a container with namespace and cgroups limit.\nUsage: miniDocker run -ti [command]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  ttyFlagName,
				Usage: "enable tty; Usage -ti",
			},
			&cli.BoolFlag{
				Name:  detachFlagName,
				Usage: "detach container; Usage -d",
			},
			&cli.StringFlag{
				Name:  memoryMaxFlagName,
				Usage: "memory limit; Usage -memMax 100m",
			},
			&cli.StringFlag{
				Name:  cpuMaxFlagName,
				Usage: "cpu limit; Usage -cpuMax 1000",
			},
			&cli.StringFlag{
				Name:  cpuSetFlagName,
				Usage: "cpuSet limit; Usage -cpuSet 2",
			},
			&cli.StringFlag{
				Name:  volumeFlagName,
				Usage: "volume; Usage: -v /etc/conf:/etc/conf",
			},
			&cli.StringFlag{
				Name:  containerNameFlagName,
				Usage: "container name; Usage: -name [name]",
			},
		},
		// 解析参数，然后运行容器
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing container command")
			}

			var cmdArray []string
			for _, arg := range context.Args().Slice() {
				cmdArray = append(cmdArray, arg)
			}

			tty := context.Bool(ttyFlagName)
			detach := context.Bool(detachFlagName)
			if tty && detach {
				return fmt.Errorf("ti and d parameter can not both provided")
			}

			resConf := &subsystems.ResourceConfig{
				MemoryMax: context.String(memoryMaxFlagName),
				CpuSet:    context.String(cpuSetFlagName),
				CpuMax:    context.String(cpuMaxFlagName),
			}
			log.Info("Resolve cgroups conf :", resConf)

			volume := context.String(volumeFlagName)
			log.Infof("Resolve volume conf : %s", volume)

			name := context.String(containerNameFlagName)
			Run(tty, cmdArray, resConf, volume, name)
			return nil
		},
	}

	commitCommand := cli.Command{
		Name:  "commit",
		Usage: "Commit a container into image.\nUsage: miniDocker commit [image name]",
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing image name")
			}
			imageName := context.Args().Get(0)
			commitContainer(imageName)
			return nil
		},
	}

	var listCommand = cli.Command{
		Name:  "ps",
		Usage: "List all the containers.\nUsage: miniDocker ps",
		Action: func(context *cli.Context) error {
			ListContainers()
			return nil
		},
	}

	var logCommand = cli.Command{
		Name:  "logs",
		Usage: "Print logs of a container.\n Usage: miniDocker logs [containerName]",
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("please input your container name")
			}
			containerName := context.Args().Get(0)
			logContainer(containerName)
			return nil
		},
	}

	app.Commands = []*cli.Command{&initCommand, &runCommand, &commitCommand, &listCommand, &logCommand}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
