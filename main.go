package main

import (
	"cc.com/miniDocker/cgroups/subsystems"
	"cc.com/miniDocker/container"
	"cc.com/miniDocker/network"
	_ "cc.com/miniDocker/nsenter"
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
const envFlagName = "e"
const netFlagName = "net"
const portFlagName = "p"

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
			&cli.StringSliceFlag{
				Name:  envFlagName,
				Usage: "set environment; Usage -e [env]",
			},
			&cli.StringFlag{
				Name:  netFlagName,
				Usage: "container network; Usage -net [network name]",
			},
			&cli.StringSliceFlag{
				Name:  portFlagName,
				Usage: "port mapping; Usage -p [hostPort:containerPort]",
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

			imageName := cmdArray[0]
			cmdArray = cmdArray[1:]

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

			envSlice := context.StringSlice(envFlagName)

			net := context.String(netFlagName)
			portMapping := context.StringSlice(portFlagName)
			Run(tty, cmdArray, resConf, volume, name, imageName, envSlice, net, portMapping)
			return nil
		},
	}

	commitCommand := cli.Command{
		Name:  "commit",
		Usage: "Commit a container into image.\nUsage: miniDocker commit [container name] [image name]",
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 2 {
				return fmt.Errorf("missing image name or container name")
			}
			containerName := context.Args().Get(0)
			imageName := context.Args().Get(0)
			commitContainer(containerName, imageName)
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

	var execCommand = cli.Command{
		Name:  "exec",
		Usage: "Exec a command into container.\nUsage: miniDocker exec name command",
		Action: func(context *cli.Context) error {
			// This is for callback
			if os.Getenv(EnvExecPid) != "" {
				log.Infof("Callback pid %s", os.Getgid())
				return nil
			}

			if context.Args().Len() < 2 {
				return fmt.Errorf("missing container name or command")
			}
			containerName := context.Args().Get(0)
			var commandArray []string
			for _, arg := range context.Args().Tail() {
				commandArray = append(commandArray, arg)
			}
			ExecContainer(containerName, commandArray)
			return nil
		},
	}

	var stopCommand = cli.Command{
		Name:  "stop",
		Usage: "Stop a container. \n Usage: miniDocker stop [name]",
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing container name")
			}
			containerName := context.Args().Get(0)
			stopContainer(containerName)
			return nil
		},
	}

	var removeCommand = cli.Command{
		Name:  "rm",
		Usage: "Remove unused containers.\n Usage: miniDocker rm [name]",
		Action: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				return fmt.Errorf("missing container name")
			}
			containerName := context.Args().Get(0)
			removeContainer(containerName)
			return nil
		},
	}

	var networkCommand = cli.Command{
		Name:  "network",
		Usage: "Container network commands.",
		Subcommands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create a container network.\n Usage miniDocker network create -driver [driver] -subnet [subnet] [network name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "driver",
						Usage: "network driver",
					},
					&cli.StringFlag{
						Name:  "subnet",
						Usage: "subnet cidr",
					},
				},
				Action: func(context *cli.Context) error {
					if context.Args().Len() < 1 {
						return fmt.Errorf("missing network name")
					}
					if err := network.Init(); err != nil {
						return err
					}
					if err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args().Get(0)); err != nil {
						return fmt.Errorf("create network error: %+v", err)
					}
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "List container network.\n Usage: miniDocker list",
				Action: func(context *cli.Context) error {
					if err := network.Init(); err != nil {
						return err
					}
					network.ListNetwork()
					return nil
				},
			},
			{
				Name:  "remove",
				Usage: "Remove container network.\n Usage: miniDocker remove [network name]",
				Action: func(context *cli.Context) error {
					if context.Args().Len() < 1 {
						return fmt.Errorf("missing network name")
					}
					if err := network.Init(); err != nil {
						return err
					}
					if err := network.DeleteNetwork(context.Args().Get(0)); err != nil {
						return fmt.Errorf("remove network error: %+v", err)
					}
					return nil
				},
			},
		},
	}

	app.Commands = []*cli.Command{
		&initCommand,
		&runCommand,
		&commitCommand,
		&listCommand,
		&logCommand,
		&execCommand,
		&stopCommand,
		&removeCommand,
		&networkCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
