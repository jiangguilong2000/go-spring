module github.com/jiangguilong2000/go-spring/spring-core

go 1.14

require (
	github.com/antonmedv/expr v1.9.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/jiangguilong2000/go-spring/spring-base v1.0.1
	github.com/magiconair/properties v1.8.5
	github.com/pelletier/go-toml v1.9.4
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jiangguilong2000/go-spring/spring-base => ../spring-base
