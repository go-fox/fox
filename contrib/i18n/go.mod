module github.com/go-fox/contrilb/i18n

require (
	dario.cat/mergo v1.0.0
	github.com/go-fox/fox v0.0.0-20240914094022-3a0dec96e7ce
	github.com/go-fox/sugar v0.0.0-20240726072231-c5b19210270e
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.uber.org/automaxprocs v1.5.3 // indirect
	golang.org/x/sys v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
	google.golang.org/grpc v1.65.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/google/safetext v0.0.0-20240104143208-7a7d9b3d812f
	github.com/google/uuid v1.6.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-fox/fox => ../../

go 1.22.1
