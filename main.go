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
const memoryMaxFlagName = "memMax"
const cpuMaxFlagName = "cpuMax"
const cpuSetFlagName = "cpuSet"

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
		Usage: "Create a container with namespace and cgroups limit.\nminiDocker run -ti [command]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  ttyFlagName,
				Usage: "enable tty",
			},
			&cli.StringFlag{
				Name:  memoryMaxFlagName,
				Usage: "memory limit",
			},
			&cli.StringFlag{
				Name:  cpuMaxFlagName,
				Usage: "cpu limit",
			},
			&cli.StringFlag{
				Name:  cpuSetFlagName,
				Usage: "cpuSet limit",
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
			resConf := &subsystems.ResourceConfig{
				MemoryMax: context.String(memoryMaxFlagName),
				CpuSet:    context.String(cpuSetFlagName),
				CpuMax:    context.String(cpuMaxFlagName),
			}

			Run(tty, cmdArray, resConf)
			return nil
		},
	}

	app.Commands = []*cli.Command{&initCommand, &runCommand}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
