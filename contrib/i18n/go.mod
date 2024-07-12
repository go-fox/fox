module github.com/go-fox/contrilb/i18n

require (
	dario.cat/mergo v1.0.0
	github.com/go-fox/fox v0.0.1
	github.com/go-fox/sugar v0.0.0-20240606100759-2030575881d7
	google.golang.org/protobuf v1.34.2
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
