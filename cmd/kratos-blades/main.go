package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/exuan/kratos-blades/template"
)

func main() {
	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "kratos-blades",
		Usage:                "a go-kratos scaffold",
		Commands: []*cli.Command{
			template.ProtoCmd,
			template.ServiceCmd,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
