package module

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModule 测试模块实现
type TestModule struct {
	name               string
	order              int
	initErr            error
	closeErr           error
	initCalled         bool
	afterAllInitCalled bool
	afterStartCalled   bool
	beforeCloseCalled  bool
	afterStopCalled    bool
}

func (tm *TestModule) Name() string { return tm.name }
func (tm *TestModule) Order() int   { return tm.order }

func (tm *TestModule) Init(ctx context.Context) error {
	tm.initCalled = true
	return tm.initErr
}

func (tm *TestModule) AfterAllInit(ctx context.Context) {
	tm.afterAllInitCalled = true
}

func (tm *TestModule) AfterStart(ctx context.Context) {
	tm.afterStartCalled = true
}

func (tm *TestModule) BeforeClose(ctx context.Context) error {
	tm.beforeCloseCalled = true
	return tm.closeErr
}

func (tm *TestModule) AfterStop(ctx context.Context) {
	tm.afterStopCalled = true
}

// TestModuleInterface 测试 Module 接口
func TestModuleInterface(t *testing.T) {
	module := &TestModule{
		name:  "test",
		order: 1,
	}

	assert.Equal(t, "test", module.Name())
	assert.Equal(t, 1, module.Order())

	// 测试方法调用
	ctx := context.Background()

	err := module.Init(ctx)
	assert.NoError(t, err)
	assert.True(t, module.initCalled)

	module.AfterAllInit(ctx)
	assert.True(t, module.afterAllInitCalled)

	module.AfterStart(ctx)
	assert.True(t, module.afterStartCalled)

	err = module.BeforeClose(ctx)
	assert.NoError(t, err)
	assert.True(t, module.beforeCloseCalled)

	module.AfterStop(ctx)
	assert.True(t, module.afterStopCalled)
}
