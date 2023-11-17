module github.com/jiangguilong2000/go-spring/examples/spring-boot-demo

go 1.16

require (
	github.com/jiangguilong2000/go-spring/spring-base v1.0.1
	github.com/jiangguilong2000/go-spring/spring-core v1.0.1
	github.com/jiangguilong2000/go-spring/starter-echo v1.0.1
	github.com/jiangguilong2000/go-spring/starter-go-redis v1.0.1
	github.com/jiangguilong2000/go-spring/starter-gorm v1.0.1
	github.com/labstack/echo/v4 v4.6.1
	github.com/spf13/viper v1.3.1
	go.mongodb.org/mongo-driver v1.7.3
	gorm.io/gorm v1.22.4
)

//replace (
//	github.com/jiangguilong2000/go-spring/spring-base => ../../spring/spring-base
//	github.com/jiangguilong2000/go-spring/spring-core => ../../spring/spring-core
//	github.com/jiangguilong2000/go-spring/spring-echo => ../../spring/spring-echo
//	github.com/jiangguilong2000/go-spring/spring-go-redis => ../../spring/spring-go-redis
//	github.com/jiangguilong2000/go-spring/starter-echo => ../../starter/starter-echo
//	github.com/jiangguilong2000/go-spring/starter-go-redis => ../../starter/starter-go-redis
//	github.com/jiangguilong2000/go-spring/starter-gorm => ../../starter/starter-gorm
//)
