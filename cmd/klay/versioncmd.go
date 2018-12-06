package main

import (
	"fmt"
	"github.com/ground-x/go-gxplatform/cmd/utils"
	"github.com/ground-x/go-gxplatform/params"
	"gopkg.in/urfave/cli.v1"
)

var versionCommand = cli.Command{
	Action:    utils.MigrateFlags(version),
	Name:      "version",
	Usage:     "Show version number",
	ArgsUsage: " ",
	Category:  "MISCELLANEOUS COMMANDS",
}

func version(ctx *cli.Context) error {
	fmt.Print("Klaytn ")
	if gitTag != "" {
		// stable version
		fmt.Println(params.Version)
	} else {
		// unstable version
		fmt.Println(params.VersionWithCommit(gitCommit))
	}
	return nil
}
