package main

import (
	"fmt"
	"log"

	"github.com/go-kratos/kratos/v2/serverinfo"
	"github.com/go-kratos/kratos/v2/serverinfo/provider"
)

func main() {
	// 示例：使用全局注册表管理 Provider 和构建 ServerInfo
	k8sProvider := provider.NewK8sProvider()
	configProvider := provider.NewConfigProvider()
	buildProvider := provider.NewBuildProvider()
	localProvider := provider.NewLocalProvider()

	serverinfo.RegisterGlobalProvider(k8sProvider)
	serverinfo.RegisterGlobalProvider(configProvider)
	serverinfo.RegisterGlobalProvider(buildProvider)
	serverinfo.RegisterGlobalProvider(localProvider)

	// 方式1：使用新的 Get 方法（推荐）
	info := serverinfo.Get()
	fmt.Printf("通过 Get() 获取的 ServerInfo:\n")
	fmt.Printf("  ServiceName: %s\n", info.ServiceName())
	fmt.Printf("  Version: %s\n", info.Version())
	fmt.Printf("  Namespace: %s\n", info.Namespace())
	fmt.Printf("  PodName: %s\n", info.PodName())
	fmt.Printf("  PodIndex: %s\n", info.PodIndex())
	fmt.Printf("  AppId: %s\n", info.AppId())
	fmt.Printf("  ArtifactId: %s\n", info.ArtifactId())
	fmt.Printf("  RegionId: %s\n", info.RegionId())
	fmt.Printf("  ChannelId: %s\n", info.ChannelId())

	// 方式2：使用原有的 BuildGlobalServerInfo 方法（保持兼容性）
	info2, err := serverinfo.BuildGlobalServerInfo()
	if err != nil {
		log.Printf("构建 ServerInfo 失败: %v", err)
		return
	}
	fmt.Printf("\n通过 BuildGlobalServerInfo() 获取的 ServerInfo:\n")
	fmt.Printf("  ServiceName: %s\n", info2.ServiceName())
	fmt.Printf("  Namespace: %s\n", info2.Namespace())
	fmt.Printf("  PodName: %s\n", info2.PodName())

	// 方式3：通过全局注册表进行更多操作
	registry := serverinfo.GetGlobalRegistry()
	fmt.Printf("\n注册的 Provider 数量: %d\n", registry.GetProviderCount())

	providers := registry.GetProvidersByPriority()
	fmt.Printf("按优先级排序的 Provider:\n")
	for i, p := range providers {
		fmt.Printf("  %d. %s (优先级: %d)\n", i+1, p.GetName(), p.GetPriority())
	}

	// 检查特定 Provider 是否存在
	if registry.HasProvider("k8s") {
		fmt.Println("K8s Provider 已注册")
	}

	// 演示动态管理 Provider
	fmt.Printf("\n移除 config Provider 前，Provider 数量: %d\n", registry.GetProviderCount())
	registry.RemoveProvider("config")
	fmt.Printf("移除 config Provider 后，Provider 数量: %d\n", registry.GetProviderCount())

	// 重新注册
	serverinfo.RegisterGlobalProvider(configProvider)
	fmt.Printf("重新注册 config Provider 后，Provider 数量: %d\n", registry.GetProviderCount())
}
