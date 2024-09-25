module github.com/go-fox/cmd/protoc-gen-go

go 1.22.1

require google.golang.org/protobuf v1.34.2

require github.com/go-fox/fox v0.0.0-20240911041716-918b71cb3969 // indirect

replace google.golang.org/protobuf => github.com/go-fox/protobuf-go v0.0.0-20240925024828-89e5667145ee

replace github.com/go-fox/fox => ../../
