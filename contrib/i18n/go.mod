module github.com/go-fox/fox/contrib/i18n

require (
	dario.cat/mergo v1.0.1
	github.com/go-fox/fox v0.0.0-20250210153006-90b39c7c7809
	github.com/go-fox/sugar v0.0.0-20241003034413-d0ef6605084f
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

require (
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/google/safetext v0.0.0-20240722112252-5a72de7e7962
	github.com/google/uuid v1.6.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-fox/fox => ../../

go 1.22.1
