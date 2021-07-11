/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gs

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"syscall"

	"github.com/go-spring/spring-core/arg"
	"github.com/go-spring/spring-core/bean"
	"github.com/go-spring/spring-core/cast"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/environ"
	"github.com/go-spring/spring-core/grpc"
	"github.com/go-spring/spring-core/log"
	"github.com/go-spring/spring-core/mq"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/web"
)

type ApplicationContext interface {
	Go(fn func(ctx context.Context))
	Prop(key string, opts ...conf.GetOption) interface{}
	Bind(i interface{}, opts ...conf.BindOption) error
	Get(i interface{}, selectors ...bean.Selector) error
	Wire(objOrCtor interface{}, ctorArgs ...arg.Arg) (interface{}, error)
	Invoke(fn interface{}, args ...arg.Arg) ([]interface{}, error)
}

type applicationContext struct{ app *App }

func (ctx *applicationContext) Go(fn func(ctx context.Context)) {
	ctx.app.Go(fn)
}

func (ctx *applicationContext) Prop(key string, opts ...conf.GetOption) interface{} {
	return (&pandora{ctx.app.c}).Prop(key, opts...)
}

func (ctx *applicationContext) Bind(i interface{}, opts ...conf.BindOption) error {
	return (&pandora{ctx.app.c}).Bind(i, opts...)
}

func (ctx *applicationContext) Get(i interface{}, selectors ...bean.Selector) error {
	return (&pandora{ctx.app.c}).Get(i, selectors...)
}

func (ctx *applicationContext) Wire(objOrCtor interface{}, ctorArgs ...arg.Arg) (interface{}, error) {
	return (&pandora{ctx.app.c}).Wire(objOrCtor, ctorArgs...)
}

func (ctx *applicationContext) Invoke(fn interface{}, args ...arg.Arg) ([]interface{}, error) {
	return (&pandora{ctx.app.c}).Invoke(fn, args...)
}

// ApplicationRunner 命令行启动器接口
type ApplicationRunner interface {
	Run(ctx ApplicationContext)
}

// ApplicationEvent 应用运行过程中的事件
type ApplicationEvent interface {
	OnStartApplication(ctx ApplicationContext) // 应用启动的事件
	OnStopApplication(ctx ApplicationContext)  // 应用停止的事件
}

// App 应用
type App struct {

	// 应用上下文
	c *Container

	banner string

	envIncludePatterns []string
	envExcludePatterns []string

	// 属性列表解析完成后的回调
	mapOfOnProperty map[string]interface{}

	exitChan chan struct{}

	RootRouter    web.RootRouter
	BindConsumers *BindConsumers
	GRPCServers   *GRPCServers
}

type BindConsumers struct{ consumers []mq.Consumer }

func (c *BindConsumers) Add(consumer mq.Consumer) {
	c.consumers = append(c.consumers, consumer)
}

func (c *BindConsumers) ForEach(fn func(mq.Consumer)) {
	for _, consumer := range c.consumers {
		fn(consumer)
	}
}

type GRPCServers struct{ servers map[string]*grpc.Server }

func (c *GRPCServers) Add(serviceName string, server *grpc.Server) {
	c.servers[serviceName] = server
}

func (c *GRPCServers) ForEach(fn func(string, *grpc.Server)) {
	for serviceName, server := range c.servers {
		fn(serviceName, server)
	}
}

// NewApp application 的构造函数
func NewApp() *App {
	return &App{
		c:                  New(),
		envIncludePatterns: []string{`.*`},
		mapOfOnProperty:    make(map[string]interface{}),
		exitChan:           make(chan struct{}),
		RootRouter:         web.NewRootRouter(),
		BindConsumers:      new(BindConsumers),
		GRPCServers:        new(GRPCServers),
	}
}

// Banner 自定义 banner 字符串。
func (app *App) Banner(banner string) {
	app.banner = banner
}

// EnvIncludePatterns 需要添加的环境变量。
func (app *App) EnvIncludePatterns(patterns []string) {
	app.envIncludePatterns = patterns
}

// EnvExcludePatterns 需要排除的环境变量。
func (app *App) EnvExcludePatterns(patterns []string) {
	app.envExcludePatterns = patterns
}

func (app *App) Run() {

	// 响应控制台的 Ctrl+C 及 kill 命令。
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		log.Infof("program will exit because of signal %v", sig)
		app.ShutDown()
	}()

	app.start()
	<-app.exitChan

	log.Info("application exiting")
	app.c.Close()
	log.Info("application exited")
}

func (app *App) start() {

	app.Object(app.BindConsumers)
	app.Object(app.GRPCServers)

	e := newEnvironment()
	err := e.prepare(
		envIncludePatterns(app.envIncludePatterns),
		envExcludePatterns(app.envExcludePatterns),
	)
	util.Panic(err).When(err != nil)

	configLocation := func() string {
		s := e.Get(environ.SpringConfigLocation, conf.Def("config/"))
		return cast.ToString(s)
	}()

	showBanner := cast.ToBool(e.Get(environ.SpringBannerVisible))
	if showBanner {
		PrintBanner(app.getBanner(configLocation))
	}

	profile := cast.ToString(e.Get(environ.SpringActiveProfile))
	p, err := app.profile(configLocation, profile)
	util.Panic(err).When(err != nil)

	// 保存从配置文件加载的属性
	for _, k := range p.Keys() {
		app.c.p.Set(k, p.Get(k))
	}

	// 保存从环境变量和命令行解析的属性
	for _, k := range e.p.Keys() {
		app.c.p.Set(k, e.p.Get(k))
	}

	for key, f := range app.mapOfOnProperty {
		t := reflect.TypeOf(f)
		in := reflect.New(t.In(0)).Elem()
		err = app.c.p.Bind(in, conf.Key(key))
		util.Panic(err).When(err != nil)
		reflect.ValueOf(f).Call([]reflect.Value{in})
	}

	app.c.Refresh()

	ctx := &applicationContext{app}

	var runners []ApplicationRunner
	err = ctx.Get(&runners)
	util.Panic(err).When(err != nil)

	// 执行命令行启动器
	for _, r := range runners {
		r.Run(ctx)
	}

	var events []ApplicationEvent
	err = ctx.Get(&events)
	util.Panic(err).When(err != nil)

	// 通知应用启动事件
	for _, e := range events {
		e.OnStartApplication(ctx)
	}

	// 通知应用停止事件
	app.Go(func(c context.Context) {
		select {
		case <-c.Done():
			for _, e := range events {
				e.OnStopApplication(ctx)
			}
		}
	})

	log.Info("application started successfully")
}

func (app *App) getBanner(configLocation string) string {
	if app.banner != "" {
		return app.banner
	}
	file := path.Join(configLocation, "banner.txt")
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return DefaultBanner
	}
	return string(b)
}

func (app *App) profile(configLocation string, profile string) (*conf.Properties, error) {

	p := conf.New()
	if err := app.loadConfigFile(p, configLocation, ""); err != nil {
		return nil, err
	}

	if profile != "" {
		if err := app.loadConfigFile(p, configLocation, profile); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (app *App) loadConfigFile(p *conf.Properties, configLocation string, profile string) error {

	filename := "application"
	if len(profile) > 0 {
		filename += "-" + profile
	}

	extArray := []string{".properties", ".yaml", ".toml"}
	for _, ext := range extArray {
		err := p.Load(filepath.Join(configLocation, filename+ext))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// ShutDown 关闭执行器
func (app *App) ShutDown() {
	select {
	case <-app.exitChan:
		// chan 已关闭，无需再次关闭。
	default:
		close(app.exitChan)
	}
}

// Property 设置 key 对应的属性值，如果 key 对应的属性值已经存在则 Set 方法会
// 覆盖旧值。Set 方法除了支持 string 类型的属性值，还支持 int、uint、bool 等
// 其他基础数据类型的属性值。特殊情况下，Set 方法也支持 slice 、map 与基础数据
// 类型组合构成的属性值，其处理方式是将组合结构层层展开，可以将组合结构看成一棵树，
// 那么叶子结点的路径就是属性的 key，叶子结点的值就是属性的值。
func (app *App) Property(key string, value interface{}) {
	app.c.Property(key, value)
}

func (app *App) OnProperty(key string, fn interface{}) {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		panic(errors.New("fn should be a func(value_type)"))
	}
	if t.NumIn() != 1 || !util.IsValueType(t.In(0)) || t.NumOut() != 0 {
		panic(errors.New("fn should be a func(value_type)"))
	}
	app.mapOfOnProperty[key] = fn
}

// Object 注册对象形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (app *App) Object(i interface{}) *BeanDefinition {
	return app.c.register(NewBean(reflect.ValueOf(i)))
}

// Provide 注册构造函数形式的 bean ，需要注意的是该方法在注入开始后就不能再调用了。
func (app *App) Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return app.c.register(NewBean(ctor, args...))
}

// Go 创建安全可等待的 goroutine，fn 要求的 ctx 对象由 IoC 容器提供，当 IoC 容
// 器关闭时 ctx会 发出 Done 信号， fn 在接收到此信号后应当立即退出。
func (app *App) Go(fn func(ctx context.Context)) {
	app.c.Go(fn)
}

// Route 返回和 Mapping 绑定的路由分组
func (app *App) Route(basePath string) *web.Router {
	return app.RootRouter.Route(basePath)
}

// HandleRequest 注册任意 HTTP 方法处理函数
func (app *App) HandleRequest(method uint32, path string, fn web.Handler) *web.Mapper {
	return app.RootRouter.HandleRequest(method, path, fn)
}

// RequestMapping 注册任意 HTTP 方法处理函数
func (app *App) RequestMapping(method uint32, path string, fn web.HandlerFunc) *web.Mapper {
	return app.RootRouter.RequestMapping(method, path, fn)
}

// RequestBinding 注册任意 HTTP 方法处理函数
func (app *App) RequestBinding(method uint32, path string, fn interface{}) *web.Mapper {
	return app.RootRouter.RequestBinding(method, path, fn)
}

// HandleGet 注册 GET 方法处理函数
func (app *App) HandleGet(path string, fn web.Handler) *web.Mapper {
	return app.RootRouter.HandleGet(path, fn)
}

// GetMapping 注册 GET 方法处理函数
func (app *App) GetMapping(path string, fn web.HandlerFunc) *web.Mapper {
	return app.RootRouter.GetMapping(path, fn)
}

// GetBinding 注册 GET 方法处理函数
func (app *App) GetBinding(path string, fn interface{}) *web.Mapper {
	return app.RootRouter.GetBinding(path, fn)
}

// HandlePost 注册 POST 方法处理函数
func (app *App) HandlePost(path string, fn web.Handler) *web.Mapper {
	return app.RootRouter.HandlePost(path, fn)
}

// PostMapping 注册 POST 方法处理函数
func (app *App) PostMapping(path string, fn web.HandlerFunc) *web.Mapper {
	return app.RootRouter.PostMapping(path, fn)
}

// PostBinding 注册 POST 方法处理函数
func (app *App) PostBinding(path string, fn interface{}) *web.Mapper {
	return app.RootRouter.PostBinding(path, fn)
}

// HandlePut 注册 PUT 方法处理函数
func (app *App) HandlePut(path string, fn web.Handler) *web.Mapper {
	return app.RootRouter.HandlePut(path, fn)
}

// PutMapping 注册 PUT 方法处理函数
func (app *App) PutMapping(path string, fn web.HandlerFunc) *web.Mapper {
	return app.RootRouter.PutMapping(path, fn)
}

// PutBinding 注册 PUT 方法处理函数
func (app *App) PutBinding(path string, fn interface{}) *web.Mapper {
	return app.RootRouter.PutBinding(path, fn)
}

// HandleDelete 注册 DELETE 方法处理函数
func (app *App) HandleDelete(path string, fn web.Handler) *web.Mapper {
	return app.RootRouter.HandleDelete(path, fn)
}

// DeleteMapping 注册 DELETE 方法处理函数
func (app *App) DeleteMapping(path string, fn web.HandlerFunc) *web.Mapper {
	return app.RootRouter.DeleteMapping(path, fn)
}

// DeleteBinding 注册 DELETE 方法处理函数
func (app *App) DeleteBinding(path string, fn interface{}) *web.Mapper {
	return app.RootRouter.DeleteBinding(path, web.BIND(fn))
}

// Filter 注册 web.Filter 对象。
func (app *App) Filter(objOrCtor interface{}, ctorArgs ...arg.Arg) *BeanDefinition {
	b := NewBean(objOrCtor, ctorArgs...)
	return app.c.register(b).Export((*web.Filter)(nil))
}

// Consume 注册 MQ 消费者。
func (app *App) Consume(fn interface{}, topics ...string) {
	app.BindConsumers.Add(mq.Bind(fn, topics...))
}

// GRPCClient 注册 gRPC 服务客户端，fn 是 gRPC 自动生成的客户端构造函数。
func (app *App) GRPCClient(fn interface{}, endpoint string) *BeanDefinition {
	return app.c.register(NewBean(fn, endpoint))
}

// GRPCServer 注册 gRPC 服务提供者，fn 是 gRPC 自动生成的服务注册函数，
// serviceName 是服务名称，必须对应 *_grpc.pg.go 文件里面 grpc.ServerDesc
// 的 ServiceName 字段，server 是服务提供者对象。
func (app *App) GRPCServer(serviceName string, fn interface{}, service interface{}) {
	app.GRPCServers.Add(serviceName, &grpc.Server{
		Service:  service,
		Register: fn,
	})
}
