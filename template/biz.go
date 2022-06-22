package template

import (
	"errors"
	"github.com/urfave/cli/v2"
)

var bizCmd = &cli.Command{
	Name: "biz",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "proto", Aliases: []string{"p"}, Usage: "proto file", Required: true},
		&cli.StringFlag{Name: "target_dir", Aliases: []string{"t"}, DefaultText: "internal/biz", Value: "internal/biz"},
	},
	Action: bizRun,
}

func bizRun(ctx *cli.Context) error {
	return errors.New("todo")
}
