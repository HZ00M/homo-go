// Package serverinfo 提供服务器信息管理功能
package serverinfo

// ProviderSet 定义依赖注入集合（注释掉 wire 相关代码，因为项目中没有 wire 依赖）
// import "github.com/google/wire"
// var ProviderSet = wire.NewSet(
// 	// 提供 ServerInfo 实例
// 	Get(),
// )

// ProvideServerInfoForWire 创建新的 ServerInfo 实例（用于依赖注入）
func ProvideServerInfoForWire() *ServerInfo {
	// 使用 Get 方法获取全局唯一的 ServerInfo 实例
	return Get()
}

// ProvideServerInfo 提供 ServerInfo 实例（兼容性函数）
func ProvideServerInfo() (*ServerInfo, error) {
	// 从全局注册表构建 ServerInfo
	return BuildGlobalServerInfo()
}
