package main

import (
	"fmt"
	"os"

	shippercli "github.com/cygnetdigital/shipper/internal/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "shipper",
		Usage:   "deploy and release services",
		Flags:   []cli.Flag{},
		Version: "dev",
		Commands: []*cli.Command{
			shippercli.Deploy,
			shippercli.Release,
			shippercli.Remove,
			shippercli.CI,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}
