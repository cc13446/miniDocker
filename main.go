package main

import (
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
			cmd := context.Args().Get(0)
			// run init process
			return container.RunContainerInitProcess(cmd, nil)
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
		},
		// 解析参数，然后运行容器
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing container command")
			}
			cmd := context.Args().Get(0)
			tty := context.Bool(ttyFlagName)
			Run(tty, cmd)
			return nil
		},
	}

	app.Commands = []*cli.Command{&initCommand, &runCommand}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
