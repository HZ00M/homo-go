package serverinfo

// Provider 提供者接口
type Provider interface {
	// GetName 返回提供者名称
	GetName() string

	// GetPriority 返回提供者优先级，数字越小优先级越高
	GetPriority() int

	// CanProvide 检查是否可以提供指定字段
	CanProvide(field string) bool

	// Provide 提供指定字段的值
	Provide(field string) (string, error)
}
