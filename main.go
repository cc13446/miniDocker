package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const usage = `MiniDocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works.
Enjoy it, just for fun.
`

func main() {

	app := cli.NewApp()
	app.Name = "miniDocker"
	app.Usage = usage

	runCommand := cli.Command{
		Name:  "run",
		Usage: "Create a container with namespace and cgroups limit.\nminiDocker run -ti [command]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "ti",
				Usage: "enable tty",
			},
		},
	}
	app.Commands = []*cli.Command{&runCommand}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
