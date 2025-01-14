package quanx

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/go-xuan/quanx/common/constx"
	"github.com/go-xuan/quanx/core/cachex"
	"github.com/go-xuan/quanx/core/configx"
	"github.com/go-xuan/quanx/core/ginx"
	"github.com/go-xuan/quanx/core/gormx"
	"github.com/go-xuan/quanx/core/nacosx"
	"github.com/go-xuan/quanx/core/redisx"
	"github.com/go-xuan/quanx/net/ipx"
	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/os/filex"
	"github.com/go-xuan/quanx/os/fmtx"
	"github.com/go-xuan/quanx/os/logx"
	"github.com/go-xuan/quanx/os/syncx"
	"github.com/go-xuan/quanx/os/taskx"
	"github.com/go-xuan/quanx/types/anyx"
	"github.com/go-xuan/quanx/types/stringx"
	"github.com/go-xuan/quanx/utils/marshalx"
)

var engine *Engine

// Engine 服务启动器
type Engine struct {
	switches       map[Option]bool           // 服务运行开关
	config         *Config                   // 服务配置数据，使用 initAppConfig()将配置文件加载到此
	configDir      string                    // 服务配置文件夹, 使用 SetConfigDir()设置配置文件读取路径
	ginEngine      *gin.Engine               // gin框架引擎实例
	ginRouters     []func(*gin.RouterGroup)  // gin路由的预加载方法，使用 AddGinRouter()添加自行实现的路由注册方法
	ginMiddlewares []gin.HandlerFunc         // gin中间件的预加载方法，使用 AddGinRouter()添加gin中间件
	customFuncs    []func()                  // 自定义初始化函数 使用 AddCustomFunc()添加自定义函数
	configurators  []configx.Configurator    // 配置器，使用 AddConfigurator()添加配置器对象，被添加对象必须为指针类型，且需要实现 configx.Configurator 接口
	gormTables     map[string][]gormx.Tabler // gorm表结构对象，使用 AddTable() / AddSourceTable() 添加至表结构初始化任务列表，需要实现 gormx.Tabler 接口
	queue          *taskx.QueueScheduler     // Engine启动时的队列任务
}

// GetEngine 获取当前Engine
func GetEngine() *Engine {
	if engine == nil {
		engine = NewEngine(
			EnableDebug(),
		)
	}
	return engine
}

// NewEngine 初始化Engine
func NewEngine(opts ...EngineOptionFunc) *Engine {
	if engine == nil {
		engine = &Engine{
			config:         &Config{},
			configDir:      constx.DefaultConfDir,
			customFuncs:    make([]func(), 0),
			configurators:  make([]configx.Configurator, 0),
			ginMiddlewares: make([]gin.HandlerFunc, 0),
			gormTables:     make(map[string][]gormx.Tabler),
			switches:       make(map[Option]bool),
		}
		gin.SetMode(gin.ReleaseMode)
		engine.SetOptions(opts...)
	}
	// 设置默认日志输出
	log.SetOutput(logx.DefaultWriter())
	log.SetFormatter(logx.DefaultFormatter())
	// 设置服务启动队列
	engine.enableQueue()
	return engine
}

// RUN 服务运行
func (e *Engine) RUN() {
	if e.switches[enableQueue] { // 任务队列方式启动
		e.queue.Execute()
	} else { // 默认方式启动
		syncx.OnceDo(e.initAppConfig)     // 1.初始化应用配置
		syncx.OnceDo(e.initInnerConfig)   // 2.初始化内置组件（log/nacos/gorm/redis/cache）
		syncx.OnceDo(e.initOuterConfig)   // 3.初始化外置组件
		syncx.OnceDo(e.runCustomFunction) // 4.运行自定义函数
		syncx.OnceDo(e.startServer)       // 5.启动服务
	}
	e.switches[running] = true
}

func (e *Engine) checkRunning() {
	if e.switches[running] {
		panic("engine has already running")
	}
}

// Queue task id
const (
	taskInitAppConfig     = "init_app_config"     // 初始化应用配置
	taskInitInnerConfig   = "init_inner_config"   // 初始化内置组件（log/nacos/gorm/redis/cache）
	taskInitOuterConfig   = "init_outer_config"   // 初始化外置组件
	taskRunCustomFunction = "run_custom_function" // 运行自定义函数
	taskStartServer       = "start_server"        // 启动web服务
)

// 是否启用队列
func (e *Engine) enableQueue() {
	if e.switches[enableQueue] && e.queue == nil {
		queue := taskx.Queue()
		queue.Add(engine.initAppConfig, taskInitAppConfig)         // 1.初始化应用配置
		queue.Add(engine.initInnerConfig, taskInitInnerConfig)     // 2.初始化内置组件（log/nacos/gorm/redis/cache）
		queue.Add(engine.initOuterConfig, taskInitOuterConfig)     // 3.初始化外置组件
		queue.Add(engine.runCustomFunction, taskRunCustomFunction) // 4.运行自定义函数
		queue.Add(engine.startServer, taskStartServer)             // 5.启动服务
		engine.queue = queue
	}
}

// 加载服务配置文件
func (e *Engine) initAppConfig() {
	e.checkRunning()
	var config = e.config
	var server = &Server{}
	// 先设置默认值
	if err := anyx.SetDefaultValue(server); err != nil {
		panic(errorx.Wrap(err, "set default value error"))
	} else {
		config.Server = server
	}
	// 读取配置文件
	if path := e.GetConfigPath(constx.DefaultConfigFilename); filex.Exists(path) {
		if err := marshalx.UnmarshalFromFile(path, config); err != nil {
			panic(errorx.Wrap(err, "unmarshal file failed: "+path))
		}
	} else if err := marshalx.WriteYaml(path, config); err != nil {
		panic(errorx.Wrap(err, "set default value error"))
	}
	if config.Server.Host == "" {
		config.Server.Host = ipx.GetLocalIP()
	}
	// 从nacos加载配置
	if e.switches[enableNacos] && config.Nacos != nil {
		e.ExecuteConfigurator(config.Nacos, true)
		if config.Nacos.EnableNaming() {
			// 注册nacos服务Nacos
			if err := nacosx.Register(config.Server.Instance()); err != nil {
				panic(errorx.Wrap(err, "nacos register error"))
			}
		}
	} else {
		e.switches[enableNacos] = false
	}
	e.config = config
}

// 初始化服务基础组件（log/gorm/redis）
func (e *Engine) initInnerConfig() {
	e.checkRunning()
	// 初始化日志
	var serverName = stringx.IfZero(e.config.Server.Name, "app")
	logConf := anyx.IfZero(e.config.Log, &logx.Config{FileName: serverName + ".log"})
	e.ExecuteConfigurator(logConf, true)

	// 初始化数据库连接
	if e.config.Database != nil {
		e.ExecuteConfigurator(e.config.Database)
	} else if e.switches[multiDatabase] {
		var database = &gormx.MultiConfig{}
		e.ExecuteConfigurator(database, true)
		e.config.Database = database
	} else {
		var database = &gormx.Config{}
		e.ExecuteConfigurator(database)
		e.config.Database = &gormx.MultiConfig{database}
	}

	// 初始化表结构
	if gormx.Initialized() {
		for _, source := range gormx.Sources() {
			if dst, ok := e.gormTables[source]; ok {
				if err := gormx.InitTable(source, dst...); err != nil {
					panic(errorx.Wrap(err, "init table struct and data failed"))
				}
			}
		}
	}

	// 初始化redis连接
	if e.config.Redis != nil {
		e.ExecuteConfigurator(e.config.Redis)
	} else if e.switches[multiRedis] {
		var redis = &redisx.MultiConfig{}
		e.ExecuteConfigurator(redis, true)
		e.config.Redis = redis
	} else {
		var redis = &redisx.Config{}
		e.ExecuteConfigurator(redis)
		e.config.Redis = &redisx.MultiConfig{redis}
	}

	// 初始化缓存
	if redisx.Initialized() {
		if e.config.Cache != nil {
			e.ExecuteConfigurator(e.config.Cache)
		} else if e.switches[multiCache] {
			var cache = &cachex.MultiConfig{}
			e.ExecuteConfigurator(cache, true)
			e.config.Cache = cache
		} else {
			var cache = &cachex.Config{}
			e.ExecuteConfigurator(cache)
			e.config.Cache = &cachex.MultiConfig{cache}
		}
	}
}

// 运行自定义配置器
func (e *Engine) initOuterConfig() {
	e.checkRunning()
	for _, configurator := range e.configurators {
		e.ExecuteConfigurator(configurator)
	}
}

// 运行自定义函数
func (e *Engine) runCustomFunction() {
	e.checkRunning()
	for _, customFunc := range e.customFuncs {
		customFunc()
	}
}

func (e *Engine) startServer() {
	e.startWebServer()
}

// 启动web服务
func (e *Engine) startWebServer() {
	e.checkRunning()
	defer PanicRecover()
	if e.switches[enableDebug] {
		gin.SetMode(gin.DebugMode)
	}
	if e.ginEngine == nil {
		e.ginEngine = gin.New()
	}
	e.ginEngine.Use(gin.Recovery(), ginx.RequestLogFmt)
	e.ginEngine.Use(e.ginMiddlewares...)

	host := e.config.Server.Host
	_ = e.ginEngine.SetTrustedProxies([]string{host})

	// 注册服务根路由
	group := e.ginEngine.Group(e.config.Server.ApiPrefix())
	e.initGinRouter(group)

	// 获取服务端口
	port := strconv.Itoa(e.config.Server.Port)
	if e.switches[customPort] {
		port = os.Getenv("PORT")
	}

	// 启动服务
	log.Infof(`API接口请求地址: http://%s:%s`, host, port)
	if err := e.ginEngine.Run(":" + port); err != nil {
		panic(errorx.Wrap(err, "gin engine run failed"))
	}
}

// initGinRouter 执行gin的路由加载函数
func (e *Engine) initGinRouter(group *gin.RouterGroup) {
	if len(e.ginRouters) > 0 {
		for _, router := range e.ginRouters {
			router(group)
		}
	} else {
		log.Warn("gin router is empty")
	}
}

// AddCustomFunc 添加自定义函数
func (e *Engine) AddCustomFunc(funcs ...func()) {
	e.checkRunning()
	if len(funcs) > 0 {
		e.customFuncs = append(e.customFuncs, funcs...)
	}
}

// AddConfigurator 新增自定义配置器
func (e *Engine) AddConfigurator(configurators ...configx.Configurator) {
	e.checkRunning()
	if len(configurators) > 0 {
		e.configurators = append(e.configurators, configurators...)
	}
}

// ExecuteConfigurator 运行配置器
func (e *Engine) ExecuteConfigurator(configurator configx.Configurator, must ...bool) {
	e.checkRunning()
	var configFrom, mustRun = "local@" + e.GetConfigPath(constx.DefaultConfigFilename) + " or tag@default",
		anyx.Default(false, must...)
	if reader := configurator.Reader(); reader != nil {
		if e.switches[enableNacos] {
			group := stringx.IfZero(reader.NacosGroup, e.config.Server.Name)
			reader.NacosGroup = group
			if err := nacosx.This().ScanConfig(configurator, group, reader.NacosDataId, reader.Listen); err == nil {
				configFrom, mustRun = "nacos@"+group+"@"+reader.NacosDataId, true
			}
		} else {
			path := e.GetConfigPath(reader.FilePath)
			if err := marshalx.UnmarshalFromFile(e.GetConfigPath(reader.FilePath), configurator); err == nil {
				configFrom, mustRun = "local@"+path, true
			}
		}
	}
	if mustRun {
		if e.switches[enableDebug] {
			log.Info("configurator data: ", configurator.Format())
		}
		if err := configx.Execute(configurator); err != nil {
			log.WithField("configFrom", configFrom).
				Error(fmtx.Red.String("configurator execute failed ==> "), err)
		} else {
			log.WithField("configFrom", configFrom).
				Info(fmtx.Green.String("configurator execute success"))
		}
	}
}

// LoadingLocalConfig 加载本地配置项（立即加载）
func (e *Engine) LoadingLocalConfig(v any, path string) {
	if err := marshalx.UnmarshalFromFile(path, v); err != nil {
		panic(errorx.Wrap(err, "unmarshal config file failed"))
	}
}

// LoadingNacosConfig 加载nacos配置（以自定义函数的形式延迟加载）
func (e *Engine) LoadingNacosConfig(v any, dataId string, listen ...bool) {
	e.AddCustomFunc(func() {
		if err := nacosx.This().ScanConfig(v, e.config.Server.Name, dataId, listen...); err != nil {
			panic(errorx.Wrap(err, "scan nacos config failed"))
		}
	})
}

// SetOptions 设置启动项
func (e *Engine) SetOptions(funcs ...EngineOptionFunc) {
	e.checkRunning()
	if len(funcs) > 0 {
		for _, f := range funcs {
			f(e)
		}
	}
}

// SetConfigDir 设置配置文件
func (e *Engine) SetConfigDir(dir string) {
	e.checkRunning()
	e.configDir = dir
}

// GetConfigPath 设置配置文件
func (e *Engine) GetConfigPath(filename string) string {
	if dir := e.configDir; dir != "" {
		return filepath.Join(dir, filename)
	} else {
		return filename
	}
}

// AddTable 添加 gormx.Tabler 模型
func (e *Engine) AddTable(dst ...gormx.Tabler) {
	e.AddSourceTable(constx.DefaultSource, dst...)
}

// AddSourceTable 添加某个数据源的 gormx.Table 模型
func (e *Engine) AddSourceTable(source string, dst ...gormx.Tabler) {
	e.checkRunning()
	if len(dst) > 0 {
		e.gormTables[source] = append(e.gormTables[source], dst...)
	}
}

// AddGinMiddleware 添加gin中间件
func (e *Engine) AddGinMiddleware(middleware ...gin.HandlerFunc) {
	e.checkRunning()
	if len(middleware) > 0 {
		e.ginMiddlewares = append(e.ginMiddlewares, middleware...)
	}
}

// AddGinRouter 添加gin的路由加载函数
func (e *Engine) AddGinRouter(router ...func(*gin.RouterGroup)) {
	e.checkRunning()
	if len(router) > 0 {
		e.ginRouters = append(e.ginRouters, router...)
	}
}

// AddQueueTask 使用后，会自动启用队列方式启动服务，且当前添加的任务会在 startServer 之前执行
func (e *Engine) AddQueueTask(task func(), id string) {
	e.checkRunning()
	if id == "" {
		log.Error(`add queue task failed, cause: the task id is required`)
	} else {
		e.switches[enableQueue] = true
		e.enableQueue()
		e.queue.AddBefore(task, id, taskStartServer)
		log.Info(`add queue task successfully, task id:`, id)
	}
	return
}

// PanicRecover 服务保活
func PanicRecover() {
	if err := recover(); err != nil {
		log.Error("server run panic: ", err)
		return
	}
	select {}
}
