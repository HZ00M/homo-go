// Package module 提供模块生命周期管理功能
package module

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModuleManager 测试 ModuleManager
func TestModuleManager(t *testing.T) {
	// 测试模块设置
	modules := []Module{
		&TestModule{name: "module1", order: 2},
		&TestModule{name: "module2", order: 1},
		&TestModule{name: "module3", order: 3},
	}

	mm := NewModuleManager(modules)
	mm.SetModules(modules)
	assert.Len(t, mm.GetModules(), 3)

	// 测试模块排序
	sortedModules := mm.GetModules()
	assert.Equal(t, "module2", sortedModules[0].Name()) // order 1
	assert.Equal(t, "module1", sortedModules[1].Name()) // order 2
	assert.Equal(t, "module3", sortedModules[2].Name()) // order 3
}

// TestModuleManagerInitAll 测试模块初始化
func TestModuleManagerInitAll(t *testing.T) {
	mm := NewModuleManager([]Module{})
	mm.SetModules([]Module{})

	ctx := context.Background()

	err := mm.InitAll(ctx)
	assert.NoError(t, err)
}

// TestModuleManagerInitAllWithError 测试初始化错误
func TestModuleManagerInitAllWithError(t *testing.T) {
	// 创建一个会返回错误的模块
	errorModule := &TestModule{
		name:    "error_module",
		order:   1,
		initErr: assert.AnError,
	}

	mm := NewModuleManager([]Module{errorModule})
	mm.SetModules([]Module{errorModule})

	ctx := context.Background()

	err := mm.InitAll(ctx)
	assert.Error(t, err)
}
