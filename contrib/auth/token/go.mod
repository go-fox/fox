module github.com/go-fox/fox/contrib/auth/token

require (
	github.com/duke-git/lancet/v2 v2.3.4
	github.com/go-fox/fox v0.0.0-20250210153006-90b39c7c7809
	github.com/go-fox/sugar v0.0.0-20241003034413-d0ef6605084f
	github.com/google/uuid v1.6.0
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/kr/text v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20250207012021-f9890c6ad9f3 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-fox/fox => ../../../

go 1.23.6
