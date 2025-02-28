package main

import (
	"flag"
	"github.com/go-fox/fox/cmd/protoc-gen-fox-migrate-route/generator"
	"github.com/urfave/cli/v2"
	"os"
)

var flags flag.FlagSet

var release = "0.0.0"

func main() {
	app := cli.NewApp()
	app.Name = "protoc-gen-fox-migrate-route"
	app.Usage = "protoc-gen-fox-migrate-route"
	app.Version = release
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "src",
			Usage: "src",
			Value: ".",
		},
		&cli.StringFlag{
			Name:  "table",
			Usage: "table name e.g. routes",
			Value: "route",
		},
		&cli.StringFlag{
			Name:  "dist",
			Usage: "output file directory",
			Value: "dist",
		},
		&cli.StringFlag{
			Name:  "filename",
			Usage: "output file name",
			Value: "insert_route",
		},
	}
	app.Action = func(c *cli.Context) error {
		conf := generator.Configuration{
			Src:      c.String("src"),
			Table:    generator.ToCamelCase(c.String("table"), true),
			Dist:     c.String("dist"),
			Filename: c.String("filename"),
		}
		sqlGenerator, err := generator.NewFoxRouteSQLGenerator(conf)
		if err != nil {
			return err
		}
		return sqlGenerator.Run()
	}
	_ = app.Run(os.Args)
}
