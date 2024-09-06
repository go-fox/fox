module github.com/go-fox/fox/contrib/cache

go 1.22.1

require (
	github.com/go-fox/fox v0.0.1
	github.com/go-fox/fox/contrib/clients/redis v0.0.0
	github.com/redis/go-redis/v9 v9.6.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-fox/sugar v0.0.0-20240726072231-c5b19210270e // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/urfave/cli/v2 v2.27.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-fox/fox => ../../

replace github.com/go-fox/fox/contrib/clients/redis => ../clients/redis
