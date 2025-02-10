module github.com/go-fox/cmd/protoc-gen-go

go 1.22.1

require google.golang.org/protobuf v1.36.5

require (
	github.com/go-fox/fox v0.0.0-20250210143009-df3a88f5f826 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace google.golang.org/protobuf => github.com/go-fox/protobuf-go v0.0.0-20250210141321-83ceb72da1b5
