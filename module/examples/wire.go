// Package examples 提供 ModuleKit 的使用示例
package examples

// App 应用结构体
type App struct {
	LogModule      *LogModule
	ConfigModule   *ConfigModule
	BusinessModule *BusinessModule
}

// NewApp 创建应用实例
func NewApp(log *LogModule, config *ConfigModule, business *BusinessModule) *App {
	return &App{
		LogModule:      log,
		ConfigModule:   config,
		BusinessModule: business,
	}
}

// 注意：这个文件展示了如何组织模块依赖关系
// 如果项目使用 Wire 框架，可以添加以下依赖注入代码：
//
// import "github.com/google/wire"
//
// var ProviderSet = wire.NewSet(
//     NewLogModule,
//     NewConfigModule,
//     NewBusinessModule,
//     NewApp,
// )
//
// func InitializeApp() (*App, error) {
//     wire.Build(ProviderSet)
//     return nil, nil
// }
