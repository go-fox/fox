package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/go-fox/fox/cmd/protoc-gen-fox-route/generator"
)

var flags flag.FlagSet

var release = "0.0.0"

func main() {
	conf := generator.Configuration{
		Src:         flags.String("src", ".", "source directory e.g. ."),
		Table:       flags.String("table", "routes", "route table name e.g. routes"),
		Filename:    flags.String("filename", "insert_route", "output file name e.g. insert_route"),
		Version:     flags.Bool("version", false, "version number text, e.g. 1.2.3"),
		IDColumn:    flags.String("id_column", "id", "id column name e.g. id"),
		TitleColumn: flags.String("title_column", "title", "title column name e.g. title"),
	}

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		// Enable "optional" keyword in front of type (e.g. optional string label = 1;)
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		if *conf.Version {
			fmt.Printf("protoc-gen-fox-route %v\n", release)
		}
		sqlGenerator, err := generator.NewFoxRouteSQLGenerator(conf, plugin)
		if err != nil {
			return err
		}

		return sqlGenerator.Run()
	})
}
