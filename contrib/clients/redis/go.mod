module github.com/go-fox/fox/contrib/clients/redis

go 1.22.1

require (
	github.com/go-fox/fox v0.0.1
	github.com/redis/go-redis/v9 v9.6.1
)

replace github.com/go-fox/fox => ../../../

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-fox/sugar v0.0.0-20240726072231-c5b19210270e // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
