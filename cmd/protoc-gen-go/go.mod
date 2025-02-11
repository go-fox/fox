module github.com/go-fox/fox/cmd/protoc-gen-go

go 1.22.1

require (
	github.com/go-fox/protobuf-go v0.0.0-20250211133054-7cb3c3ab6b17
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/go-fox/fox v0.0.0-20250211103327-9752fa043e79 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/go-fox/fox => ../../
	google.golang.org/protobuf => github.com/go-fox/protobuf-go v0.0.0-20250211141352-16bd050b13dc
)
