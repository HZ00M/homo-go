// Package serverinfo 提供服务器信息管理功能
package serverinfo

import (
	"time"
)

// ServerInfo 服务器信息
type ServerInfo struct {
	// 基础服务信息
	serviceName string
	version     string

	// 环境信息
	namespace string
	podName   string
	podIndex  string

	// 业务配置信息
	appId      string
	artifactId string
	regionId   string
	channelId  string

	// 元数据
	metadata map[string]string

	// 时间信息
	createdAt int64
	updatedAt int64
}

// NewServerInfo 创建新的 ServerInfo 实例（同模块内部使用）
func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		metadata:  make(map[string]string),
		createdAt: time.Now().Unix(),
		updatedAt: time.Now().Unix(),
	}
}

// 同模块内部使用的设置方法
func (s *ServerInfo) setServiceName(name string) {
	s.serviceName = name
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setVersion(version string) {
	s.version = version
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setNamespace(namespace string) {
	s.namespace = namespace
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setPodName(podName string) {
	s.podName = podName
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setPodIndex(podIndex string) {
	s.podIndex = podIndex
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setAppId(appId string) {
	s.appId = appId
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setArtifactId(artifactId string) {
	s.artifactId = artifactId
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setRegionId(regionId string) {
	s.regionId = regionId
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setChannelId(channelId string) {
	s.channelId = channelId
	s.updatedAt = time.Now().Unix()
}

func (s *ServerInfo) setMetadata(key, value string) {
	if s.metadata == nil {
		s.metadata = make(map[string]string)
	}
	s.metadata[key] = value
	s.updatedAt = time.Now().Unix()
}

// 外部只读访问方法
func (s *ServerInfo) ServiceName() string { return s.serviceName }
func (s *ServerInfo) Version() string     { return s.version }
func (s *ServerInfo) Namespace() string   { return s.namespace }
func (s *ServerInfo) PodName() string     { return s.podName }
func (s *ServerInfo) PodIndex() string    { return s.podIndex }
func (s *ServerInfo) AppId() string       { return s.appId }
func (s *ServerInfo) ArtifactId() string  { return s.artifactId }
func (s *ServerInfo) RegionId() string    { return s.regionId }
func (s *ServerInfo) ChannelId() string   { return s.channelId }
func (s *ServerInfo) CreatedAt() int64    { return s.createdAt }
func (s *ServerInfo) UpdatedAt() int64    { return s.updatedAt }

// GetMetadata 获取元数据（只读）
func (s *ServerInfo) GetMetadata(key string) string {
	if s.metadata == nil {
		return ""
	}
	return s.metadata[key]
}

// GetAllMetadata 获取所有元数据（只读）
func (s *ServerInfo) GetAllMetadata() map[string]string {
	if s.metadata == nil {
		return make(map[string]string)
	}
	// 返回副本，防止外部修改
	result := make(map[string]string)
	for k, v := range s.metadata {
		result[k] = v
	}
	return result
}

// GetInstanceName 获取实例名称
func (s *ServerInfo) GetInstanceName() string {
	if s.podName != "" {
		return s.podName
	}
	return s.serviceName
}

// IsLocalDebug 判断是否为本地调试环境
func (s *ServerInfo) IsLocalDebug() bool {
	return s.namespace == "" || s.namespace == "local" || s.namespace == "debug"
}
