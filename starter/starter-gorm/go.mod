module github.com/jiangguilong2000/go-spring/starter-gorm

go 1.14

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/jiangguilong2000/go-spring/spring-base v1.0.1
	github.com/jiangguilong2000/go-spring/spring-core v1.0.1
	gorm.io/driver/mysql v1.2.1
	gorm.io/gorm v1.22.4
)

//replace (
//	github.com/jiangguilong2000/go-spring/spring-base => ../../spring/spring-base
//	github.com/jiangguilong2000/go-spring/spring-core => ../../spring/spring-core
//)
