// Package main
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"flag"
	"fmt"

	"github.com/go-fox/sugar/util/sslice"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/go-fox/cmd/protoc-gen-fox/grpc"
	"github.com/go-fox/cmd/protoc-gen-fox/http"
	"github.com/go-fox/fox/api/annotations"
)

const version = "0.0.1"

var requireUnimplemented *bool
var useGenericStreams *bool
var omitempty *bool

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-fox %v\n", version)
		return
	}

	var flags flag.FlagSet
	requireUnimplemented = flags.Bool("require_unimplemented_servers", true, "set to false to match legacy behavior")
	useGenericStreams = flags.Bool("use_generic_streams_experimental", true, "set to true to use generic types for streaming client and server objects; this flag is EXPERIMENTAL and may be changed or removed in a future release")
	omitempty = flag.Bool("omitempty", true, "omit if google.api is empty")
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) | uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if checkHave(f, annotations.GenerateType_grpc) {
				grpc.GenerateFile(gen, f, requireUnimplemented, useGenericStreams, version)
			}
			if checkHave(f, annotations.GenerateType_http) {
				http.GenerateFile(gen, f, omitempty, version)
			}
		}
		return nil
	})
}

func checkHave(file *protogen.File, rule annotations.GenerateType) bool {
	for _, service := range file.Services {
		extension := proto.GetExtension(service.Desc.Options(), annotations.E_Service).(*annotations.ServiceRule)
		if len(extension.GetGenerate()) == 0 || sslice.Contain(extension.GetGenerate(), rule) {
			for _, method := range service.Methods {
				methodRule := proto.GetExtension(method.Desc.Options(), annotations.E_Method).(*annotations.MethodRule)
				if len(methodRule.GetGenerate()) == 0 || sslice.Contain(methodRule.GetGenerate(), rule) {
					return true
				}
			}
		}
	}
	return false
}
