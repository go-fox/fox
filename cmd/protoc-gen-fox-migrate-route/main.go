package main

import (
	"flag"
	"fmt"
	"github.com/go-fox/fox/cmd/protoc-gen-fox-migrate-route/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var flags flag.FlagSet

var release = "0.0.0"

func main() {
	conf := generator.Configuration{
		Src:        flags.String("src", ".", "source directory e.g. ."),
		Table:      flags.String("table", "route_test", "route table name e.g. routes"),
		Package:    flags.String("package", "migratedata", "output file directory e.g. dist"),
		Filename:   flags.String("filename", "insert_route", "output file name e.g. insert_route"),
		Version:    flags.Bool("version", false, "version number text, e.g. 1.2.3"),
		EntPackage: flags.String("ent_package", "", "ent package name e.g. github.com/lolopinto/ent/ent"),
	}

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		// Enable "optional" keyword in front of type (e.g. optional string label = 1;)
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		if *conf.Version {
			fmt.Printf("protoc-gen-go-http %v\n", release)
		}
		sqlGenerator, err := generator.NewFoxRouteSQLGenerator(conf, plugin)
		if err != nil {
			return err
		}

		return sqlGenerator.Run()
	})

	//app := cli.NewApp()
	//app.Name = "protoc-gen-fox-migrate-route"
	//app.Usage = "protoc-gen-fox-migrate-route"
	//app.Version = release
	//app.Flags = []cli.Flag{
	//	&cli.StringFlag{
	//		Name:  "src",
	//		Usage: "src",
	//		Value: ".",
	//	},
	//	&cli.StringFlag{
	//		Name:  "table",
	//		Usage: "table name e.g. routes",
	//		Value: "route",
	//	},
	//	&cli.StringFlag{
	//		Name:  "dist",
	//		Usage: "output file directory",
	//		Value: "dist",
	//	},
	//	&cli.StringFlag{
	//		Name:  "filename",
	//		Usage: "output file name",
	//		Value: "insert_route",
	//	},
	//}
	//app.Action = func(c *cli.Context) error {
	//	conf := generator.Configuration{
	//		Src:      c.String("src"),
	//		Table:    generator.ToCamelCase(c.String("table"), true),
	//		Dist:     c.String("dist"),
	//		Filename: c.String("filename"),
	//	}
	//	sqlGenerator, err := generator.NewFoxRouteSQLGenerator(conf)
	//	if err != nil {
	//		return err
	//	}
	//	return sqlGenerator.Run()
	//}
	//_ = app.Run(os.Args)
}
