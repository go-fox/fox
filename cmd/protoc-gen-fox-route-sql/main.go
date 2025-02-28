package main

import (
	"flag"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/go-fox/fox/cmd/protoc-gen-fox-route-sql/generator"
)

var flags flag.FlagSet

var release = "0.0.0"

func main() {
	conf := generator.Configuration{
		Version:    flags.String("version", "0.0.1", "version number text, e.g. 1.2.3"),
		OutputMode: flags.String("output_mode", "merged", `output generation mode. By default, a single openapi.yaml is generated at the out folder. Use "source_relative' to generate a separate '[inputfile].openapi.yaml' next to each '[inputfile].proto'.`),
		Table:      flags.String("table", "test_route", "table name"),
		Filename:   flags.String("filename", "fox_route.sql", "output file name"),
	}

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		if *conf.OutputMode == "source_relative" {
			for _, file := range plugin.Files {
				if !file.Generate {
					continue
				}
				outfileName := strings.TrimSuffix(file.Desc.Path(), filepath.Ext(file.Desc.Path())) + "." + *conf.Table + ".sql"
				outputFile := plugin.NewGeneratedFile(outfileName, "")
				gen := generator.NewFoxRouteSQLGenerator(plugin, conf, []*protogen.File{file})
				if err := gen.Run(outputFile); err != nil {
					return err
				}
			}
		} else {
			var files []*protogen.File
			for _, file := range plugin.Files {
				if !file.Generate {
					continue
				}
				files = append(files, file)
			}
			filename := func() string {
				if len(*conf.Filename) > 0 {
					return generator.NewVersion() + "_" + *conf.Filename + ".sql"
				}
				return *conf.Table + ".sql"
			}()
			outputFile := plugin.NewGeneratedFile(filename, "")
			return generator.NewFoxRouteSQLGenerator(plugin, conf, files).Run(outputFile)
		}
		return nil
	})
}
